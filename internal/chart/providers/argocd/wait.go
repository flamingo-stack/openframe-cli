package argocd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/pterm/pterm"
)

// WaitForApplications waits for all ArgoCD applications to be Healthy and Synced
func (m *Manager) WaitForApplications(ctx context.Context, config config.ChartInstallConfig) error {
	// Skip waiting in dry-run mode for testing
	if config.DryRun {
		return nil
	}

	// Set cluster name from config for explicit context usage (important for Windows/WSL)
	if config.ClusterName != "" {
		m.clusterName = config.ClusterName
	}

	// Check if already cancelled before starting
	if ctx.Err() != nil {
		return fmt.Errorf("operation already cancelled: %w", ctx.Err())
	}

	// Early exit if context has a short deadline (indicates timeout scenario)
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < 5*time.Second {
			// Context will expire soon - skip ArgoCD applications wait
			return nil
		}
	}

	// Create a derived context that responds to both parent cancellation AND direct signals
	// This ensures immediate response to Ctrl+C even if parent context isn't propagating fast enough
	localCtx, localCancel := context.WithCancel(ctx)
	defer localCancel()

	// Handle direct interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		localCancel() // Cancel our local context immediately
	}()

	// Check if we should start the spinner (skip if context is cancelled or expiring soon)
	shouldSkipSpinner := false

	// Check if context is cancelled
	if localCtx.Err() != nil {
		shouldSkipSpinner = true
	}

	// Check if original context is cancelled
	if ctx.Err() != nil {
		shouldSkipSpinner = true
	}

	// Check if context deadline is very close (less than 10 seconds)
	if deadline, ok := ctx.Deadline(); ok {
		timeLeft := time.Until(deadline)
		if timeLeft < 10*time.Second {
			shouldSkipSpinner = true
		}
	}

	if shouldSkipSpinner {
		// Context is cancelled or expiring soon - skip ArgoCD applications wait entirely
		return nil
	}

	// Wait for ArgoCD CRD and pods to be ready before checking applications
	if err := m.waitForArgoCDReady(localCtx, config.Verbose, config.SkipCRDs); err != nil {
		return fmt.Errorf("ArgoCD not ready: %w", err)
	}

	// Show initial verbose info if enabled
	if config.Verbose {
		pterm.Info.Println("Starting ArgoCD application synchronization...")
		pterm.Debug.Println("  - Waiting for applications to be created by app-of-apps")
		pterm.Debug.Println("  - Each application must reach Healthy + Synced status")
		pterm.Debug.Println("  - Progress updates every 10 seconds in verbose mode")
	}

	// Start pterm spinner only if not in silent/non-interactive mode
	var spinner *pterm.SpinnerPrinter
	if !config.Silent {
		spinner, _ = pterm.DefaultSpinner.
			WithRemoveWhenDone(false).
			WithShowTimer(true).
			Start("Installing ArgoCD applications...")
	} else {
		// In non-interactive mode, just show a simple info message
		pterm.Info.Println("Installing ArgoCD applications...")
	}

	var spinnerMutex sync.Mutex
	spinnerStopped := false

	// Function to stop spinner safely
	stopSpinner := func() {
		spinnerMutex.Lock()
		defer spinnerMutex.Unlock()
		if !spinnerStopped && spinner != nil && spinner.IsActive {
			spinner.Stop()
			spinnerStopped = true
		}
	}

	// Monitor for context cancellation (includes interrupt signals from parent or direct signals)
	go func() {
		<-localCtx.Done()
		stopSpinner()
	}()

	// Ensure spinner is stopped when function exits
	defer stopSpinner()

	// Bootstrap wait (30 seconds)
	bootstrapEnd := time.Now().Add(30 * time.Second)

	// Check every 10ms for immediate response
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Bootstrap phase
	for time.Now().Before(bootstrapEnd) {
		select {
		case <-localCtx.Done():
			return fmt.Errorf("operation cancelled: %w", localCtx.Err())
		case <-ticker.C:
			// Continue waiting
		}
	}

	// Main monitoring phase
	startTime := time.Now()
	timeout := 60 * time.Minute
	checkInterval := 2 * time.Second
	lastCheck := time.Now()

	// Get expected applications count
	totalAppsExpected := m.getTotalExpectedApplications(localCtx, config)
	if totalAppsExpected == 0 {
		totalAppsExpected = -1
	}

	maxAppsSeenTotal := 0
	maxAppsSeenReady := 0

	// Track applications that have ever been ready (healthy + synced) during this session
	// Once an app is ready, it stays counted even if it temporarily goes out of sync
	everReadyApps := make(map[string]bool)

	// Main loop
	for {
		select {
		case <-localCtx.Done():
			return fmt.Errorf("operation cancelled: %w", localCtx.Err())
		case <-ticker.C:
			// Check timeout
			if time.Since(startTime) > timeout {
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil && spinner.IsActive {
					spinner.Fail(fmt.Sprintf("Timeout after %v", timeout))
					spinnerStopped = true
				}
				spinnerMutex.Unlock()
				return fmt.Errorf("timeout waiting for ArgoCD applications after %v", timeout)
			}

			// Check applications every 2 seconds
			if time.Since(lastCheck) < checkInterval {
				continue
			}
			lastCheck = time.Now()

			// Parse applications
			apps, err := m.parseApplications(localCtx, config.Verbose)
			if err != nil {
				if localCtx.Err() != nil {
					return fmt.Errorf("operation cancelled: %w", localCtx.Err())
				}
				// Ignore parse errors and retry
				continue
			}

			totalApps := len(apps)
			if totalApps > maxAppsSeenTotal {
				maxAppsSeenTotal = totalApps
				// Show initial application count when first detected (verbose mode)
				if config.Verbose && totalApps > 0 {
					pterm.Info.Printf("Detected %d ArgoCD applications to synchronize\n", totalApps)
				}
			}

			if totalAppsExpected == -1 || maxAppsSeenTotal > totalAppsExpected {
				totalAppsExpected = maxAppsSeenTotal
			}

			// Track applications that have ever been ready during this session
			currentHealthyCount := 0
			currentlyReady := 0
			healthyApps := make([]string, 0)
			syncedApps := make([]string, 0)
			notReadyApps := make([]string, 0)

			for _, app := range apps {
				// Count currently healthy apps for monitoring
				if app.Health == "Healthy" {
					currentHealthyCount++
					healthyApps = append(healthyApps, app.Name)
				}

				if app.Sync == "Synced" {
					syncedApps = append(syncedApps, app.Name)
				}

				// Count currently ready apps (both healthy and synced)
				if app.Health == "Healthy" && app.Sync == "Synced" {
					currentlyReady++
					// Mark apps as "ever ready" if they are currently healthy and synced
					// Once marked, they stay counted even if they go out of sync later
					everReadyApps[app.Name] = true
				} else {
					// Track apps that are not yet ready with more detailed status
					if app.Health != "Healthy" || app.Sync != "Synced" {
						// Show the most important status issue
						var status string
						if app.Health != "Healthy" && app.Sync != "Synced" {
							status = fmt.Sprintf("%s/%s", app.Health, app.Sync)
						} else if app.Health != "Healthy" {
							status = fmt.Sprintf("Health: %s", app.Health)
						} else {
							status = fmt.Sprintf("Sync: %s", app.Sync)
						}
						notReadyApps = append(notReadyApps, fmt.Sprintf("%s (%s)", app.Name, status))
					}
				}
			}

			// Show verbose logging if enabled
			if config.Verbose && totalApps > 0 {
				elapsed := time.Since(startTime)

				// Update spinner message with current status
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil && spinner.IsActive {
					progress := ""
					if totalApps > 0 {
						progressPercent := float64(currentlyReady) / float64(totalApps) * 100
						progress = fmt.Sprintf(" (%.0f%%)", progressPercent)
					}
					spinner.UpdateText(fmt.Sprintf("Installing ArgoCD applications... %d/%d ready%s [%s]",
						currentlyReady, totalApps, progress, elapsed.Round(time.Second)))
				}
				spinnerMutex.Unlock()

				// Only show detailed status every 10 seconds to avoid spam
				if int(elapsed.Seconds())%10 == 0 {
					pterm.Info.Printf("ArgoCD Sync Progress: %d/%d applications ready (%s elapsed)\n",
						currentlyReady, totalApps, elapsed.Round(time.Second))

					// Always show pending applications when there are any
					if len(notReadyApps) > 0 {
						if len(notReadyApps) <= 8 {
							pterm.Info.Printf("  Still waiting for: %v\n", notReadyApps)
						} else {
							pterm.Info.Printf("  Still waiting for %d applications (showing first 5): %v...\n",
								len(notReadyApps), notReadyApps[:5])
						}

						// DEBUG: Show pod details for stuck applications after 7 min, every 5 minutes
						if elapsed > 7*time.Minute && int(elapsed.Seconds())%300 == 0 {
							stuckApps := []Application{}
							for _, app := range apps {
								if app.Health != "Healthy" && app.Health != "Missing" {
									stuckApps = append(stuckApps, app)
								}
							}

							if len(stuckApps) > 0 {
								pterm.Info.Printf("\n=== DEBUG: Found %d stuck application(s) ===\n", len(stuckApps))

								for _, app := range stuckApps {
									pterm.Info.Printf("\n--- %s (Health: %s, Sync: %s) ---\n", app.Name, app.Health, app.Sync)

									// Get namespace using explicit context
									nsArgs := m.getKubectlArgs("-n", "argocd", "get", "app", app.Name, "-o", "jsonpath={.spec.destination.namespace}")
									nsResult, err := m.executor.Execute(localCtx, "kubectl", nsArgs...)
									if err != nil || nsResult == nil || nsResult.Stdout == "" {
										pterm.Warning.Printf("Could not get namespace for %s\n", app.Name)
										continue
									}
									ns := strings.TrimSpace(nsResult.Stdout)

									// Get pods with issues: not Running or with restarts
									podQuery := "jsonpath={range .items[?(@.status.phase!=\"Running\")]}{.metadata.name}{\"\\t\"}{.status.phase}{\"\\t\"}{.status.containerStatuses[0].restartCount}{\"\\n\"}{end}"
									problemPodsArgs := m.getKubectlArgs("-n", ns, "get", "pods", "-o", podQuery)
									problemPodsResult, _ := m.executor.Execute(localCtx, "kubectl", problemPodsArgs...)

									// Also get pods with restarts but Running
									restartPodsQuery := "jsonpath={range .items[?(@.status.phase==\"Running\")]}{.metadata.name}{\"\\t\"}{.status.containerStatuses[0].restartCount}{\"\\n\"}{end}"
									restartPodsArgs := m.getKubectlArgs("-n", ns, "get", "pods", "-o", restartPodsQuery)
									restartPodsResult, _ := m.executor.Execute(localCtx, "kubectl", restartPodsArgs...)

									problemPods := make(map[string]bool)

									// Parse non-running pods
									if problemPodsResult != nil && problemPodsResult.Stdout != "" {
										for _, line := range strings.Split(strings.TrimSpace(problemPodsResult.Stdout), "\n") {
											if line != "" {
												podName := strings.Split(line, "\t")[0]
												problemPods[podName] = true
											}
										}
									}

									// Parse pods with restarts
									if restartPodsResult != nil && restartPodsResult.Stdout != "" {
										for _, line := range strings.Split(strings.TrimSpace(restartPodsResult.Stdout), "\n") {
											if line == "" {
												continue
											}
											parts := strings.Split(line, "\t")
											if len(parts) >= 2 && parts[1] != "0" && parts[1] != "" {
												problemPods[parts[0]] = true
											}
										}
									}

									if len(problemPods) == 0 {
										pterm.Info.Println("  No problematic pods found (may be an ArgoCD sync issue)")
										continue
									}

									pterm.Info.Printf("  Found %d pod(s) with issues\n", len(problemPods))

									for podName := range problemPods {
										pterm.Info.Printf("\n  Pod: %s\n", podName)

										// Get pod status summary using explicit context
										statusArgs := m.getKubectlArgs("-n", ns, "get", "pod", podName, "-o", "jsonpath={.status.phase}{'/'}{.status.containerStatuses[*].state}")
										statusResult, _ := m.executor.Execute(localCtx, "kubectl", statusArgs...)
										if statusResult != nil && statusResult.Stdout != "" {
											pterm.Info.Printf("  Status: %s\n", statusResult.Stdout)
										}

										// Get recent events for this pod using explicit context
										eventsArgs := m.getKubectlArgs("-n", ns, "get", "events", "--field-selector", "involvedObject.name="+podName, "--sort-by=.lastTimestamp", "-o", "custom-columns=TIME:.lastTimestamp,REASON:.reason,MESSAGE:.message", "--no-headers")
										eventsResult, _ := m.executor.Execute(localCtx, "kubectl", eventsArgs...)
										if eventsResult != nil && eventsResult.Stdout != "" {
											eventLines := strings.Split(strings.TrimSpace(eventsResult.Stdout), "\n")
											if len(eventLines) > 5 {
												eventLines = eventLines[len(eventLines)-5:]
											}
											pterm.Info.Println("  Recent Events:")
											for _, event := range eventLines {
												if event != "" {
													pterm.Info.Printf("    %s\n", event)
												}
											}
										}

										// Get last 20 lines of logs using explicit context
										logsArgs := m.getKubectlArgs("-n", ns, "logs", podName, "--tail=20", "--all-containers=true", "--prefix=true")
										logsResult, _ := m.executor.Execute(localCtx, "kubectl", logsArgs...)
										if logsResult != nil && logsResult.Stdout != "" {
											pterm.Info.Println("  Recent Logs:")
											for _, line := range strings.Split(logsResult.Stdout, "\n") {
												if line != "" {
													pterm.Info.Printf("    %s\n", line)
												}
											}
										}
									}
								}
								pterm.Info.Println("\n=== End Debug ===")
							}
						}
					}

					// Show recently completed applications
					if len(healthyApps) > 0 && len(healthyApps) <= 5 {
						startIdx := 0
						if len(healthyApps) > 5 {
							startIdx = len(healthyApps) - 5
						}
						pterm.Debug.Printf("  Recently completed: %v\n", healthyApps[startIdx:])
					}
				}
			}

			// Use the high water mark of applications that have ever been ready
			readyCount := len(everReadyApps)

			if readyCount > maxAppsSeenReady {
				maxAppsSeenReady = readyCount
			}

			// Check if deployment is complete - ALL currently detected apps must be healthy and synced
			// All apps must be currently ready (not just "ever ready")
			allReady := false
			if totalApps > 0 && currentlyReady == totalApps {
				allReady = true
			}

			// Update ready count for display purposes (still use everReady for progress tracking)
			if currentlyReady > maxAppsSeenReady {
				maxAppsSeenReady = currentlyReady
			}

			if allReady {
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil && spinner.IsActive {
					spinner.Stop()
					spinnerStopped = true
				}
				spinnerMutex.Unlock()
				pterm.Success.Println("All ArgoCD applications installed")
				return nil
			}
		}
	}
}

// waitForArgoCDReady waits for ArgoCD CRD and pods to be ready
func (m *Manager) waitForArgoCDReady(ctx context.Context, verbose bool, skipCRDs bool) error {
	maxRetries := 100 // 100 retries * 3 seconds = 5 minutes max
	retryInterval := 3 * time.Second

	// Skip CRD wait if CRDs installation was skipped (e.g., in non-interactive/CI mode)
	// In this case, the ArgoCD Helm chart will install CRDs as part of the release
	if skipCRDs {
		if verbose {
			pterm.Info.Println("Skipping ArgoCD CRD wait (CRDs managed by Helm chart)")
		}
	} else {
		// Wait for ArgoCD CRD to be available
		if verbose {
			pterm.Info.Println("Waiting for ArgoCD CRD applications.argoproj.io...")
		}

		for i := 0; i < maxRetries; i++ {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled: %w", ctx.Err())
			default:
			}

			// First verify cluster is still reachable (important for Windows/WSL stability)
			clusterCheckArgs := m.getKubectlArgs("cluster-info")
			clusterResult, clusterErr := m.executor.Execute(ctx, "kubectl", clusterCheckArgs...)
			if clusterErr != nil || clusterResult.ExitCode != 0 {
				if verbose {
					pterm.Warning.Printf("Cluster connectivity issue detected, waiting... (attempt %d/%d)\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Check if CRD exists using explicit context
			crdArgs := m.getKubectlArgs("get", "crd", "applications.argoproj.io")
			result, err := m.executor.Execute(ctx, "kubectl", crdArgs...)
			if err == nil && result.ExitCode == 0 {
				if verbose {
					pterm.Success.Println("ArgoCD CRD applications.argoproj.io is ready")
				}
				break
			}

			if i == maxRetries-1 {
				return fmt.Errorf("timeout waiting for ArgoCD CRD applications.argoproj.io")
			}

			if verbose && i%5 == 0 { // Log every 15 seconds
				pterm.Info.Println("🕒 Waiting for ArgoCD CRD applications.argoproj.io...")
			}

			time.Sleep(retryInterval)
		}
	}

	// Wait for ArgoCD pods to be ready using explicit context
	if verbose {
		pterm.Info.Println("Waiting for ArgoCD pods to be ready...")
	}

	// First, wait for pods to exist before trying to wait for them to be Ready
	// kubectl wait returns an error if no pods match the selector
	podExistenceTimeout := 120 * time.Second // 2 minutes for pods to be created
	podExistenceInterval := 3 * time.Second
	podExistenceStart := time.Now()
	podsExist := false

	for time.Since(podExistenceStart) < podExistenceTimeout {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Check if any ArgoCD pods exist
		checkArgs := m.getKubectlArgs("-n", "argocd", "get", "pods",
			"-l", "app.kubernetes.io/part-of=argocd",
			"-o", "jsonpath={.items[*].metadata.name}")
		checkResult, checkErr := m.executor.Execute(ctx, "kubectl", checkArgs...)

		if checkErr == nil && checkResult != nil && strings.TrimSpace(checkResult.Stdout) != "" {
			podNames := strings.Fields(checkResult.Stdout)
			if len(podNames) > 0 {
				if verbose {
					pterm.Info.Printf("Found %d ArgoCD pod(s), waiting for them to be ready...\n", len(podNames))
				}
				podsExist = true
				break
			}
		}

		if verbose && int(time.Since(podExistenceStart).Seconds())%15 == 0 {
			pterm.Info.Println("Waiting for ArgoCD pods to be created...")
		}

		time.Sleep(podExistenceInterval)
	}

	if !podsExist {
		// No pods found - collect diagnostics and return error
		pterm.Warning.Println("No ArgoCD pods found after waiting. Collecting diagnostics...")
		m.printArgoCDPodDiagnostics(ctx)
		return fmt.Errorf("timeout waiting for ArgoCD pods to be created (no pods found with label app.kubernetes.io/part-of=argocd)")
	}

	// Now wait for pods to be Ready using kubectl wait with timeout
	waitArgs := m.getKubectlArgs("-n", "argocd", "wait",
		"--for=condition=Ready", "pod",
		"-l", "app.kubernetes.io/part-of=argocd",
		"--timeout=300s")
	result, err := m.executor.Execute(ctx, "kubectl", waitArgs...)

	if err != nil || result.ExitCode != 0 {
		// Collect diagnostic info before returning error
		m.printArgoCDPodDiagnostics(ctx)
		return fmt.Errorf("timeout waiting for ArgoCD pods to be ready: %w", err)
	}

	if verbose {
		pterm.Success.Println("ArgoCD pods are ready")
	}

	return nil
}

// printArgoCDPodDiagnostics prints diagnostic information about ArgoCD pods when they fail to become ready
func (m *Manager) printArgoCDPodDiagnostics(ctx context.Context) {
	pterm.Warning.Println("ArgoCD pods failed to become ready. Collecting diagnostics...")

	// First check Helm release status to understand if ArgoCD was installed correctly
	helmStatusArgs := []string{"status", "argo-cd", "-n", "argocd"}
	helmResult, _ := m.executor.Execute(ctx, "helm", helmStatusArgs...)
	if helmResult != nil && helmResult.Stdout != "" {
		pterm.Info.Println("\nHelm release status:")
		// Show just the first few lines of status
		statusLines := strings.Split(helmResult.Stdout, "\n")
		for i, line := range statusLines {
			if i < 10 {
				pterm.Info.Printf("  %s\n", line)
			}
		}
	} else {
		pterm.Warning.Println("Could not get Helm release status for argo-cd")
	}

	// Check for deployments in argocd namespace
	deployArgs := m.getKubectlArgs("-n", "argocd", "get", "deployments", "-o", "wide")
	deployResult, _ := m.executor.Execute(ctx, "kubectl", deployArgs...)
	if deployResult != nil && deployResult.Stdout != "" {
		pterm.Info.Println("\nArgoCD deployments:")
		for _, line := range strings.Split(strings.TrimSpace(deployResult.Stdout), "\n") {
			pterm.Info.Printf("  %s\n", line)
		}
	}

	// Get all pods in argocd namespace with their status
	podArgs := m.getKubectlArgs("-n", "argocd", "get", "pods", "-o", "wide")
	podResult, _ := m.executor.Execute(ctx, "kubectl", podArgs...)
	if podResult != nil && podResult.Stdout != "" {
		pterm.Info.Println("ArgoCD pods status:")
		for _, line := range strings.Split(strings.TrimSpace(podResult.Stdout), "\n") {
			pterm.Info.Printf("  %s\n", line)
		}
	}

	// Get pods that are not ready and show their details
	notReadyQuery := "jsonpath={range .items[?(@.status.phase!=\"Running\")]}{.metadata.name}{\"\\n\"}{end}"
	notReadyArgs := m.getKubectlArgs("-n", "argocd", "get", "pods", "-o", notReadyQuery)
	notReadyResult, _ := m.executor.Execute(ctx, "kubectl", notReadyArgs...)

	var problemPods []string
	if notReadyResult != nil && notReadyResult.Stdout != "" {
		for _, pod := range strings.Split(strings.TrimSpace(notReadyResult.Stdout), "\n") {
			if pod != "" {
				problemPods = append(problemPods, pod)
			}
		}
	}

	// Also check for pods that are Running but not Ready (container issues)
	runningNotReadyQuery := "jsonpath={range .items[?(@.status.phase==\"Running\")]}{.metadata.name}{\" \"}{.status.conditions[?(@.type==\"Ready\")].status}{\"\\n\"}{end}"
	runningNotReadyArgs := m.getKubectlArgs("-n", "argocd", "get", "pods", "-o", runningNotReadyQuery)
	runningNotReadyResult, _ := m.executor.Execute(ctx, "kubectl", runningNotReadyArgs...)
	if runningNotReadyResult != nil && runningNotReadyResult.Stdout != "" {
		for _, line := range strings.Split(strings.TrimSpace(runningNotReadyResult.Stdout), "\n") {
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[1] != "True" && parts[0] != "" {
				problemPods = append(problemPods, parts[0])
			}
		}
	}

	// Show details for problem pods
	for _, podName := range problemPods {
		pterm.Info.Printf("\n--- Pod: %s ---\n", podName)

		// Get pod describe summary (conditions and events)
		describeArgs := m.getKubectlArgs("-n", "argocd", "get", "pod", podName, "-o",
			"jsonpath={\"Phase: \"}{.status.phase}{\"\\nConditions:\\n\"}{range .status.conditions[*]}{\"  \"}{.type}{\"=\"}{.status}{\" (\"}{.reason}{\")\\n\"}{end}")
		describeResult, _ := m.executor.Execute(ctx, "kubectl", describeArgs...)
		if describeResult != nil && describeResult.Stdout != "" {
			pterm.Info.Println(describeResult.Stdout)
		}

		// Get container statuses
		containerArgs := m.getKubectlArgs("-n", "argocd", "get", "pod", podName, "-o",
			"jsonpath={\"Containers:\\n\"}{range .status.containerStatuses[*]}{\"  \"}{.name}{\": ready=\"}{.ready}{\", restarts=\"}{.restartCount}{\"\\n\"}{end}")
		containerResult, _ := m.executor.Execute(ctx, "kubectl", containerArgs...)
		if containerResult != nil && containerResult.Stdout != "" {
			pterm.Info.Println(containerResult.Stdout)
		}

		// Get recent events for this pod
		eventsArgs := m.getKubectlArgs("-n", "argocd", "get", "events",
			"--field-selector", "involvedObject.name="+podName,
			"--sort-by=.lastTimestamp",
			"-o", "custom-columns=TIME:.lastTimestamp,TYPE:.type,REASON:.reason,MESSAGE:.message",
			"--no-headers")
		eventsResult, _ := m.executor.Execute(ctx, "kubectl", eventsArgs...)
		if eventsResult != nil && eventsResult.Stdout != "" {
			eventLines := strings.Split(strings.TrimSpace(eventsResult.Stdout), "\n")
			// Show last 5 events
			if len(eventLines) > 5 {
				eventLines = eventLines[len(eventLines)-5:]
			}
			pterm.Info.Println("Recent Events:")
			for _, event := range eventLines {
				if event != "" {
					pterm.Info.Printf("  %s\n", event)
				}
			}
		}

		// Get container logs if pod exists
		logsArgs := m.getKubectlArgs("-n", "argocd", "logs", podName, "--tail=10", "--all-containers=true")
		logsResult, _ := m.executor.Execute(ctx, "kubectl", logsArgs...)
		if logsResult != nil && logsResult.Stdout != "" {
			pterm.Info.Println("Recent Logs (last 10 lines):")
			for _, line := range strings.Split(strings.TrimSpace(logsResult.Stdout), "\n") {
				pterm.Info.Printf("  %s\n", line)
			}
		}
	}

	// If no specific problem pods found, show general namespace events
	if len(problemPods) == 0 {
		pterm.Info.Println("\nNo specific problem pods found. Showing recent namespace events:")
		nsEventsArgs := m.getKubectlArgs("-n", "argocd", "get", "events",
			"--sort-by=.lastTimestamp",
			"-o", "custom-columns=TIME:.lastTimestamp,TYPE:.type,OBJECT:.involvedObject.name,REASON:.reason,MESSAGE:.message",
			"--no-headers")
		nsEventsResult, _ := m.executor.Execute(ctx, "kubectl", nsEventsArgs...)
		if nsEventsResult != nil && nsEventsResult.Stdout != "" {
			eventLines := strings.Split(strings.TrimSpace(nsEventsResult.Stdout), "\n")
			if len(eventLines) > 10 {
				eventLines = eventLines[len(eventLines)-10:]
			}
			for _, event := range eventLines {
				if event != "" {
					pterm.Info.Printf("  %s\n", event)
				}
			}
		}
	}
}
