package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

									// Get all pods as JSON to avoid Windows WSL escaping issues with jsonpath
									allPodsArgs := m.getKubectlArgs("-n", ns, "get", "pods", "-o", "json")
									allPodsResult, _ := m.executor.Execute(localCtx, "kubectl", allPodsArgs...)

									problemPods := make(map[string]bool)

									// Parse pods from JSON and identify problematic ones
									if allPodsResult != nil && allPodsResult.Stdout != "" {
										var podList corev1.PodList
										if err := json.Unmarshal([]byte(allPodsResult.Stdout), &podList); err == nil {
											for _, pod := range podList.Items {
												// Non-running pods are problematic
												if pod.Status.Phase != corev1.PodRunning {
													problemPods[pod.Name] = true
													continue
												}
												// Check for restarts in running pods
												for _, cs := range pod.Status.ContainerStatuses {
													if cs.RestartCount > 0 {
														problemPods[pod.Name] = true
														break
													}
												}
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

										// Get pod status as JSON to avoid Windows WSL escaping issues
										podStatusArgs := m.getKubectlArgs("-n", ns, "get", "pod", podName, "-o", "json")
										podStatusResult, _ := m.executor.Execute(localCtx, "kubectl", podStatusArgs...)
										if podStatusResult != nil && podStatusResult.Stdout != "" {
											var pod corev1.Pod
											if err := json.Unmarshal([]byte(podStatusResult.Stdout), &pod); err == nil {
												// Build status string similar to the old jsonpath output
												var states []string
												for _, cs := range pod.Status.ContainerStatuses {
													if cs.State.Waiting != nil {
														states = append(states, fmt.Sprintf("waiting(%s)", cs.State.Waiting.Reason))
													} else if cs.State.Running != nil {
														states = append(states, "running")
													} else if cs.State.Terminated != nil {
														states = append(states, fmt.Sprintf("terminated(%s)", cs.State.Terminated.Reason))
													}
												}
												pterm.Info.Printf("  Status: %s/%s\n", pod.Status.Phase, strings.Join(states, ","))
											}
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

// waitForArgoCDReady waits for ArgoCD CRD and pods to be ready using native Go clients
// This reduces reliance on external kubectl binary
func (m *Manager) waitForArgoCDReady(ctx context.Context, verbose bool, skipCRDs bool) error {
	maxRetries := 100 // 100 retries * 3 seconds = 5 minutes max
	retryInterval := 3 * time.Second

	// Initialize Kubernetes clients for native API access
	if err := m.initKubernetesClients(); err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to initialize native clients, falling back to kubectl: %v\n", err)
		}
		return m.waitForArgoCDReadyViaKubectl(ctx, verbose, skipCRDs)
	}

	// Skip CRD wait if CRDs installation was skipped (e.g., in non-interactive/CI mode)
	if skipCRDs {
		if verbose {
			pterm.Info.Println("Skipping ArgoCD CRD wait (CRDs managed by Helm chart)")
		}
	} else {
		// Wait for ArgoCD CRD to be available using native apiextensions client
		if verbose {
			pterm.Info.Println("Waiting for ArgoCD CRD applications.argoproj.io...")
		}

		for i := 0; i < maxRetries; i++ {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled: %w", ctx.Err())
			default:
			}

			// Check CRD existence using native client
			_, err := m.apiextClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "applications.argoproj.io", metav1.GetOptions{})
			if err == nil {
				if verbose {
					pterm.Success.Println("ArgoCD CRD applications.argoproj.io is ready")
				}
				break
			}

			if !k8serrors.IsNotFound(err) {
				// Non-404 error - might be connectivity issue
				if verbose {
					pterm.Warning.Printf("Cluster connectivity issue detected: %v (attempt %d/%d)\n", err, i+1, maxRetries)
				}
			}

			if i == maxRetries-1 {
				return fmt.Errorf("timeout waiting for ArgoCD CRD applications.argoproj.io")
			}

			if verbose && i%5 == 0 {
				pterm.Info.Println("Waiting for ArgoCD CRD applications.argoproj.io...")
			}

			time.Sleep(retryInterval)
		}
	}

	// Wait for ArgoCD pods to be ready using native Kubernetes client
	if verbose {
		pterm.Info.Println("Waiting for ArgoCD pods to be ready...")
	}

	podExistenceTimeout := 120 * time.Second
	podExistenceInterval := 3 * time.Second
	podExistenceStart := time.Now()
	podsExist := false

	for time.Since(podExistenceStart) < podExistenceTimeout {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// List ArgoCD pods using native client
		podList, err := m.kubeClient.CoreV1().Pods("argocd").List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/part-of=argocd",
		})

		if err == nil && len(podList.Items) > 0 {
			if verbose {
				pterm.Info.Printf("Found %d ArgoCD pod(s), waiting for them to be ready...\n", len(podList.Items))
			}
			podsExist = true
			break
		}

		if verbose && int(time.Since(podExistenceStart).Seconds())%15 == 0 {
			pterm.Info.Println("Waiting for ArgoCD pods to be created...")
		}

		time.Sleep(podExistenceInterval)
	}

	if !podsExist {
		pterm.Warning.Println("No ArgoCD pods found after waiting. Collecting diagnostics...")
		m.printArgoCDPodDiagnostics(ctx)
		return fmt.Errorf("timeout waiting for ArgoCD pods to be created (no pods found with label app.kubernetes.io/part-of=argocd)")
	}

	// Wait for all pods to be Ready using native client
	podReadyTimeout := 5 * time.Minute
	podReadyStart := time.Now()

	for time.Since(podReadyStart) < podReadyTimeout {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		podList, err := m.kubeClient.CoreV1().Pods("argocd").List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/part-of=argocd",
		})

		if err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to list pods: %v\n", err)
			}
			time.Sleep(retryInterval)
			continue
		}

		allReady := true
		for _, pod := range podList.Items {
			if !isPodReady(&pod) {
				allReady = false
				break
			}
		}

		if allReady && len(podList.Items) > 0 {
			if verbose {
				pterm.Success.Println("ArgoCD pods are ready")
			}
			return nil
		}

		time.Sleep(retryInterval)
	}

	m.printArgoCDPodDiagnostics(ctx)
	return fmt.Errorf("timeout waiting for ArgoCD pods to be ready")
}

// isPodReady checks if a pod has the Ready condition set to True
func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// waitForArgoCDReadyViaKubectl is the fallback method using kubectl
func (m *Manager) waitForArgoCDReadyViaKubectl(ctx context.Context, verbose bool, skipCRDs bool) error {
	maxRetries := 100
	retryInterval := 3 * time.Second

	if skipCRDs {
		if verbose {
			pterm.Info.Println("Skipping ArgoCD CRD wait (CRDs managed by Helm chart)")
		}
	} else {
		if verbose {
			pterm.Info.Println("Waiting for ArgoCD CRD applications.argoproj.io...")
		}

		for i := 0; i < maxRetries; i++ {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled: %w", ctx.Err())
			default:
			}

			clusterCheckArgs := m.getKubectlArgs("cluster-info")
			clusterResult, clusterErr := m.executor.Execute(ctx, "kubectl", clusterCheckArgs...)
			if clusterErr != nil || clusterResult.ExitCode != 0 {
				if verbose {
					pterm.Warning.Printf("Cluster connectivity issue detected, waiting... (attempt %d/%d)\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

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

			if verbose && i%5 == 0 {
				pterm.Info.Println("Waiting for ArgoCD CRD applications.argoproj.io...")
			}

			time.Sleep(retryInterval)
		}
	}

	if verbose {
		pterm.Info.Println("Waiting for ArgoCD pods to be ready...")
	}

	podExistenceTimeout := 120 * time.Second
	podExistenceInterval := 3 * time.Second
	podExistenceStart := time.Now()
	podsExist := false

	for time.Since(podExistenceStart) < podExistenceTimeout {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Use -o json to avoid Windows WSL escaping issues with jsonpath
		checkArgs := m.getKubectlArgs("-n", "argocd", "get", "pods",
			"-l", "app.kubernetes.io/part-of=argocd",
			"-o", "json")
		checkResult, checkErr := m.executor.Execute(ctx, "kubectl", checkArgs...)

		if checkErr == nil && checkResult != nil && strings.TrimSpace(checkResult.Stdout) != "" {
			var podList corev1.PodList
			podNames := []string{}
			if jsonErr := json.Unmarshal([]byte(checkResult.Stdout), &podList); jsonErr == nil {
				for _, p := range podList.Items {
					podNames = append(podNames, p.Name)
				}
			}
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
		pterm.Warning.Println("No ArgoCD pods found after waiting. Collecting diagnostics...")
		m.printArgoCDPodDiagnostics(ctx)
		return fmt.Errorf("timeout waiting for ArgoCD pods to be created (no pods found with label app.kubernetes.io/part-of=argocd)")
	}

	waitArgs := m.getKubectlArgs("-n", "argocd", "wait",
		"--for=condition=Ready", "pod",
		"-l", "app.kubernetes.io/part-of=argocd",
		"--timeout=300s")
	result, err := m.executor.Execute(ctx, "kubectl", waitArgs...)

	if err != nil || result.ExitCode != 0 {
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
	// Use --field-selector instead of jsonpath to avoid Windows WSL escaping issues
	notReadyArgs := m.getKubectlArgs("-n", "argocd", "get", "pods", "--field-selector=status.phase!=Running", "-o", "name")
	notReadyResult, _ := m.executor.Execute(ctx, "kubectl", notReadyArgs...)

	var problemPods []string
	if notReadyResult != nil && notReadyResult.Stdout != "" {
		for _, pod := range strings.Split(strings.TrimSpace(notReadyResult.Stdout), "\n") {
			if pod != "" {
				// Strip "pod/" prefix from -o name output
				podName := strings.TrimPrefix(pod, "pod/")
				problemPods = append(problemPods, podName)
			}
		}
	}

	// Also check for pods that are Running but not Ready (container issues)
	// Use -o json and parse in Go to avoid Windows WSL escaping issues with jsonpath
	runningPodsArgs := m.getKubectlArgs("-n", "argocd", "get", "pods", "--field-selector=status.phase=Running", "-o", "json")
	runningPodsResult, _ := m.executor.Execute(ctx, "kubectl", runningPodsArgs...)
	if runningPodsResult != nil && runningPodsResult.Stdout != "" {
		var podList corev1.PodList
		if err := json.Unmarshal([]byte(runningPodsResult.Stdout), &podList); err == nil {
			for _, pod := range podList.Items {
				// Check if the Ready condition is not True
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status != corev1.ConditionTrue {
						problemPods = append(problemPods, pod.Name)
						break
					}
				}
			}
		}
	}

	// Show details for problem pods
	for _, podName := range problemPods {
		pterm.Info.Printf("\n--- Pod: %s ---\n", podName)

		// Get pod details as JSON to avoid Windows WSL escaping issues with jsonpath
		podArgs := m.getKubectlArgs("-n", "argocd", "get", "pod", podName, "-o", "json")
		podResult, _ := m.executor.Execute(ctx, "kubectl", podArgs...)
		if podResult != nil && podResult.Stdout != "" {
			var pod corev1.Pod
			if err := json.Unmarshal([]byte(podResult.Stdout), &pod); err == nil {
				// Print phase
				pterm.Info.Printf("Phase: %s\n", pod.Status.Phase)

				// Print conditions
				pterm.Info.Println("Conditions:")
				for _, cond := range pod.Status.Conditions {
					reason := string(cond.Reason)
					if reason == "" {
						reason = "-"
					}
					pterm.Info.Printf("  %s=%s (%s)\n", cond.Type, cond.Status, reason)
				}

				// Print container statuses
				pterm.Info.Println("Containers:")
				for _, cs := range pod.Status.ContainerStatuses {
					pterm.Info.Printf("  %s: ready=%t, restarts=%d\n", cs.Name, cs.Ready, cs.RestartCount)
				}
			}
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
