package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	uispinner "github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
	appsv1 "k8s.io/api/apps/v1"
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

	// Derive a cancellable context from the parent. The parent is already
	// signal-cancelled (root ExecuteContext), so Ctrl-C / SIGTERM propagates
	// here immediately — no local signal handler required.
	localCtx, localCancel := context.WithCancel(ctx)
	defer localCancel()

	// Check if we should start the spinner (skip if context is cancelled or expiring soon)
	shouldSkipSpinner := localCtx.Err() != nil

	// Check if context is cancelled

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

	// Initial repo-server health check - catch issues early
	initialIssue := m.checkRepoServerHealth(localCtx, true)
	if initialIssue != nil {
		// If repo-server has already restarted, proactively restart it to clear any stuck state
		// This helps CI environments where the pod may have OOM'd during initial setup
		if initialIssue.Type == "resource" && initialIssue.Recoverable {
			m.triggerRepoServerRecovery(localCtx, "")
		}
	}
	// Show initial verbose info if enabled
	if config.Verbose {
		pterm.Info.Println("Starting ArgoCD application synchronization...")
		pterm.Debug.Println("  - Waiting for applications to be created by app-of-apps")
		pterm.Debug.Println("  - Each application must reach Healthy + Synced status")
		pterm.Debug.Println("  - Progress updates every 10 seconds in verbose mode")
	}

	// Start pterm spinner only if not in silent/non-interactive mode
	var spinner *uispinner.Spinner
	if !config.Silent {
		spinner = uispinner.New().WithTimer()
		spinner.Start("Installing ArgoCD applications...")
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
		if !spinnerStopped && spinner != nil {
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

	// Bootstrap wait (30 seconds) with periodic cluster health checks
	bootstrapEnd := time.Now().Add(30 * time.Second)
	bootstrapHealthCheckInterval := 5 * time.Second
	lastBootstrapHealthCheck := time.Now()
	consecutiveFailures := 0
	maxConsecutiveFailures := 5 // Increased from 3 for better WSL resilience in CI environments

	// Check every 10ms for immediate response
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Bootstrap phase - check cluster health every 5 seconds
	for time.Now().Before(bootstrapEnd) {
		select {
		case <-localCtx.Done():
			return fmt.Errorf("operation cancelled: %w", localCtx.Err())
		case <-ticker.C:
			// Check cluster health periodically during bootstrap
			if time.Since(lastBootstrapHealthCheck) >= bootstrapHealthCheckInterval {
				lastBootstrapHealthCheck = time.Now()
				if err := m.checkClusterConnectivity(localCtx, config.Verbose); err != nil {
					consecutiveFailures++
					if consecutiveFailures >= maxConsecutiveFailures {
						stopSpinner()
						m.printClusterDiagnostics(localCtx)
						return fmt.Errorf("cluster became unreachable during bootstrap wait: %w", err)
					}
				} else {
					consecutiveFailures = 0
				}
			}
		}
	}

	// Main monitoring phase
	startTime := time.Now()
	timeout := 60 * time.Minute
	checkInterval := 2 * time.Second
	lastCheck := time.Now()
	clusterHealthCheckInterval := 10 * time.Second
	clusterHealthCheckIntervalFast := 2 * time.Second // Faster checks when errors occur
	lastClusterHealthCheck := time.Now()
	resourceCheckInterval := 5 * time.Minute // Check system resources every 5 minutes
	lastResourceCheck := time.Now()
	consecutiveFailures = 0 // Reset for main loop

	// Get expected applications count
	totalAppsExpected := m.getTotalExpectedApplications(localCtx, config)
	if totalAppsExpected == 0 {
		totalAppsExpected = -1
	}

	maxAppsSeenTotal := 0
	maxAppsSeenReady := 0
	consecutiveAllReady := 0
	stabilizationChecks := m.StabilizationChecks
	if stabilizationChecks <= 0 {
		stabilizationChecks = 15 // default: 15 * 2s = 30s
	}

	// Track applications that have ever been ready (healthy + synced) during this session
	// Once an app is ready, it stays counted even if it temporarily goes out of sync
	everReadyApps := make(map[string]bool)

	// Repo-server issue tracking for recovery logic
	repoServerRecoveryAttempts := 0
	maxRepoServerRecoveryAttempts := 3 // Increased from 2 for CI resilience
	lastRepoServerDiagnostic := time.Time{}
	repoServerDiagnosticInterval := 2 * time.Minute  // Reduced from 3 min for faster CI recovery
	appsWithRepoServerIssues := make(map[string]int) // Track consecutive failures per app
	lastRepoServerResourceCheck := time.Now()
	repoServerResourceCheckInterval := 30 * time.Second // Reduced from 1 min for faster issue detection

	// Main loop
	for {
		select {
		case <-localCtx.Done():
			return fmt.Errorf("operation cancelled: %w", localCtx.Err())
		case <-ticker.C:
			// Check timeout
			if time.Since(startTime) > timeout {
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil {
					spinner.Fail(fmt.Sprintf("Timeout after %v", timeout))
					spinnerStopped = true
				}
				spinnerMutex.Unlock()
				return fmt.Errorf("timeout waiting for ArgoCD applications after %v", timeout)
			}

			// Periodic cluster health check
			// Use faster interval (2s) when we've seen failures, normal interval (10s) otherwise
			currentHealthCheckInterval := clusterHealthCheckInterval
			if consecutiveFailures > 0 {
				currentHealthCheckInterval = clusterHealthCheckIntervalFast
			}
			if time.Since(lastClusterHealthCheck) >= currentHealthCheckInterval {
				lastClusterHealthCheck = time.Now()
				if err := m.checkClusterConnectivity(localCtx, false); err != nil {
					consecutiveFailures++

					// On Windows, try WSL recovery before giving up
					if runtime.GOOS == "windows" && consecutiveFailures >= maxConsecutiveFailures-1 {
						if wslErr := executor.TryRecoverWSL(); wslErr == nil {
							// Give WSL a moment to stabilize
							time.Sleep(3 * time.Second)
							// Retry the connectivity check
							if retryErr := m.checkClusterConnectivity(localCtx, false); retryErr == nil {
								consecutiveFailures = 0
								continue
							}
						}
					}

					if consecutiveFailures >= maxConsecutiveFailures {
						stopSpinner()
						m.printClusterDiagnostics(localCtx)
						return fmt.Errorf("cluster became unreachable while waiting for applications: %w", err)
					}

					// Add exponential backoff delay between failures to avoid hammering WSL
					backoffDelay := time.Duration(consecutiveFailures) * 2 * time.Second
					if backoffDelay > 10*time.Second {
						backoffDelay = 10 * time.Second
					}
					time.Sleep(backoffDelay)
				} else {
					consecutiveFailures = 0
				}
			}

			// Periodic resource check (every 5 minutes) - helps diagnose resource exhaustion
			if time.Since(lastResourceCheck) >= resourceCheckInterval {
				lastResourceCheck = time.Now()
				m.logResourceStatus(localCtx, config.Verbose)

				// Also check repo-server health proactively
				m.checkRepoServerHealth(localCtx, false)
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

				// Check if this is a cluster connectivity error (including WSL errors)
				errStr := err.Error()
				isConnectivityError := strings.Contains(errStr, "connection refused") ||
					strings.Contains(errStr, "cluster unreachable") ||
					strings.Contains(errStr, "was refused") ||
					strings.Contains(errStr, "Unable to connect") ||
					strings.Contains(errStr, "WSL error")

				if isConnectivityError {
					consecutiveFailures++
					pterm.Warning.Printf("Application query failed - cluster may be unreachable (%d/%d): %v\n",
						consecutiveFailures, maxConsecutiveFailures, err)

					// On Windows, try WSL recovery before giving up
					if runtime.GOOS == "windows" && consecutiveFailures >= maxConsecutiveFailures-1 {
						pterm.Info.Println("Attempting WSL recovery before giving up...")
						if wslErr := executor.TryRecoverWSL(); wslErr != nil {
							pterm.Warning.Printf("WSL recovery failed: %v\n", wslErr)
						} else {
							pterm.Success.Println("WSL recovery successful")
							// Give WSL a moment to stabilize
							time.Sleep(3 * time.Second)
						}
					}

					if consecutiveFailures >= maxConsecutiveFailures {
						stopSpinner()
						m.printClusterDiagnostics(localCtx)
						return fmt.Errorf("cluster became unreachable while waiting for applications: %w", err)
					}

					// Add backoff delay between failures
					backoffDelay := time.Duration(consecutiveFailures) * 2 * time.Second
					if backoffDelay > 10*time.Second {
						backoffDelay = 10 * time.Second
					}
					time.Sleep(backoffDelay)
				}

				// Retry on other errors (with normal interval via lastCheck)
				continue
			}

			// Reset consecutive failures on successful query
			if consecutiveFailures > 0 {
				pterm.Success.Println("Application queries restored")
				consecutiveFailures = 0
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
			assess := assessApplications(apps, everReadyApps)
			currentlyReady := assess.ready
			healthyApps := assess.healthyNames
			notReadyApps := assess.notReady

			// Show verbose logging if enabled
			if config.Verbose && totalApps > 0 {
				elapsed := time.Since(startTime)

				// Update spinner message with current status
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil {
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

						// Check for applications stuck in "Unknown" status or with repo-server issues
						unknownApps, appsWithConditionErrors := classifyAppIssues(apps, appsWithRepoServerIssues)

						// Check for repo-server issues and attempt recovery
						if len(appsWithConditionErrors) > 0 && elapsed > 2*time.Minute {
							// More frequent resource checks when repo-server issues are detected
							if time.Since(lastRepoServerResourceCheck) >= repoServerResourceCheckInterval {
								lastRepoServerResourceCheck = time.Now()
								m.checkRepoServerHealth(localCtx, false)
							}

							// Check if any app has had consistent repo-server issues
							for _, app := range appsWithConditionErrors {
								consecutiveIssues := appsWithRepoServerIssues[app.Name]

								// After 2 consecutive checks with repo-server issues, run diagnostics
								if consecutiveIssues >= 2 && time.Since(lastRepoServerDiagnostic) >= repoServerDiagnosticInterval {
									lastRepoServerDiagnostic = time.Now()

									// Attempt recovery if we haven't exceeded max attempts
									if repoServerRecoveryAttempts < maxRepoServerRecoveryAttempts {
										repoServerRecoveryAttempts++
										if m.triggerRepoServerRecovery(localCtx, app.Name) {
											// Reset the issue counter for this app to give it a fresh start
											delete(appsWithRepoServerIssues, app.Name)
										}
									} else if repoServerRecoveryAttempts == maxRepoServerRecoveryAttempts {
										repoServerRecoveryAttempts++ // Increment to prevent repeated attempts
									}
									break // Only recover one app at a time
								}
							}
						}

						// After 5 minutes, warn about Unknown status as it may indicate ArgoCD controller issues
						if len(unknownApps) > 0 && elapsed > 5*time.Minute {
							// Show detailed info for each unknown app using data we already have
							pterm.Warning.Printf("  Applications with 'Unknown' status (%d):\n", len(unknownApps))
							for _, app := range unknownApps {
								pterm.Warning.Printf("\n  --- %s (Health: %s, Sync: %s) ---\n", app.Name, app.Health, app.Sync)

								// Show source info
								if app.RepoURL != "" {
									pterm.Info.Printf("    Source: %s", app.RepoURL)
									if app.Path != "" {
										pterm.Printf(" path=%s", app.Path)
									}
									if app.TargetRevision != "" {
										pterm.Printf(" revision=%s", app.TargetRevision)
									}
									pterm.Println()
								}

								// Show condition error (this is usually the most important info)
								if app.Condition != "" {
									condType := app.ConditionType
									if condType == "" {
										condType = "Error"
									}
									pterm.Warning.Printf("    %s: %s\n", condType, app.Condition)
								}

								// Show operation state if present
								if app.OperationPhase != "" {
									pterm.Info.Printf("    Operation: %s", app.OperationPhase)
									if app.OperationMessage != "" {
										pterm.Printf(" - %s", app.OperationMessage)
									}
									pterm.Println()
								}

								// Show health message if present
								if app.HealthMessage != "" {
									pterm.Info.Printf("    Health details: %s\n", app.HealthMessage)
								}

								// Show last reconciliation time
								if app.ReconciledAt != "" {
									pterm.Info.Printf("    Last reconciled: %s\n", app.ReconciledAt)
								} else {
									pterm.Warning.Println("    Not yet reconciled (ArgoCD hasn't processed this app)")
								}
							}

							pterm.Warning.Println("\n  Possible causes: Controller pod not ready, Git repo access issues, or resource constraints.")
						}

						// DEBUG: Show pod details for stuck applications after 7 min, every 5 minutes
						if elapsed > 7*time.Minute && int(elapsed.Seconds())%300 == 0 {
							stuckApps := []Application{}
							for _, app := range apps {
								if app.Health != ArgoCDHealthHealthy && app.Health != ArgoCDHealthMissing {
									stuckApps = append(stuckApps, app)
								}
							}

							if len(stuckApps) > 0 {
								pterm.Info.Printf("\n=== DEBUG: Found %d stuck application(s) ===\n", len(stuckApps))

								for _, app := range stuckApps {
									pterm.Info.Printf("\n--- %s (Health: %s, Sync: %s) ---\n", app.Name, app.Health, app.Sync)

									// Get namespace using explicit context.
									// Use -o json and parse in Go to avoid Windows WSL escaping issues with jsonpath.
									nsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "app", app.Name, "-o", "json")
									nsResult, err := m.executor.Execute(localCtx, "kubectl", nsArgs...)
									if err != nil || nsResult == nil || nsResult.Stdout == "" {
										pterm.Warning.Printf("Could not get namespace for %s\n", app.Name)
										continue
									}
									var nsApp argoApp
									if err := json.Unmarshal([]byte(nsResult.Stdout), &nsApp); err != nil || nsApp.Spec.Destination.Namespace == "" {
										pterm.Warning.Printf("Could not get namespace for %s\n", app.Name)
										continue
									}
									ns := nsApp.Spec.Destination.Namespace

									// Get all pods as JSON to avoid Windows WSL escaping issues with jsonpath
									allPodsArgs := m.getKubectlArgs("-n", ns, "get", "pods", "-o", "json")
									allPodsResult, _ := m.executor.Execute(localCtx, "kubectl", allPodsArgs...)

									problemPods := make(map[string]bool)
									allPods := make(map[string]string) // podName -> status summary

									// Parse pods from JSON and identify problematic ones
									var podList corev1.PodList
									if allPodsResult != nil && allPodsResult.Stdout != "" {
										if err := json.Unmarshal([]byte(allPodsResult.Stdout), &podList); err == nil {
											for _, pod := range podList.Items {
												// Build status summary for all pods
												statusSummary := string(pod.Status.Phase)
												if len(pod.Status.ContainerStatuses) > 0 {
													var containerStates []string
													for _, cs := range pod.Status.ContainerStatuses {
														if cs.State.Waiting != nil {
															containerStates = append(containerStates, cs.State.Waiting.Reason)
														} else if cs.State.Terminated != nil {
															containerStates = append(containerStates, cs.State.Terminated.Reason)
														} else if cs.Ready {
															containerStates = append(containerStates, "Ready")
														}
													}
													if len(containerStates) > 0 {
														statusSummary += "/" + strings.Join(containerStates, ",")
													}
												}
												// Check init container status
												for _, ics := range pod.Status.InitContainerStatuses {
													if ics.State.Waiting != nil {
														statusSummary += " (init:" + ics.State.Waiting.Reason + ")"
													} else if ics.State.Running != nil {
														statusSummary += " (init:Running)"
													}
												}
												allPods[pod.Name] = statusSummary

												// Non-running pods are problematic
												if pod.Status.Phase != corev1.PodRunning {
													problemPods[pod.Name] = true
													continue
												}
												// Pods stuck in init containers are problematic
												for _, ics := range pod.Status.InitContainerStatuses {
													if ics.State.Waiting != nil || ics.State.Running != nil {
														problemPods[pod.Name] = true
														break
													}
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

									// Get StatefulSets to identify missing pods
									stsArgs := m.getKubectlArgs("-n", ns, "get", "statefulsets", "-o", "json")
									stsResult, _ := m.executor.Execute(localCtx, "kubectl", stsArgs...)
									missingPods := []string{}
									if stsResult != nil && stsResult.Stdout != "" {
										var stsList appsv1.StatefulSetList
										if err := json.Unmarshal([]byte(stsResult.Stdout), &stsList); err == nil {
											for _, sts := range stsList.Items {
												expectedReplicas := int32(1)
												if sts.Spec.Replicas != nil {
													expectedReplicas = *sts.Spec.Replicas
												}
												for i := int32(0); i < expectedReplicas; i++ {
													expectedPodName := fmt.Sprintf("%s-%d", sts.Name, i)
													if _, exists := allPods[expectedPodName]; !exists {
														missingPods = append(missingPods, expectedPodName)
													}
												}
											}
										}
									}

									// Get Deployments to identify missing pods
									deployArgs := m.getKubectlArgs("-n", ns, "get", "deployments", "-o", "json")
									deployResult, _ := m.executor.Execute(localCtx, "kubectl", deployArgs...)
									if deployResult != nil && deployResult.Stdout != "" {
										var deployList appsv1.DeploymentList
										if err := json.Unmarshal([]byte(deployResult.Stdout), &deployList); err == nil {
											for _, deploy := range deployList.Items {
												expectedReplicas := int32(1)
												if deploy.Spec.Replicas != nil {
													expectedReplicas = *deploy.Spec.Replicas
												}
												// Count pods matching this deployment
												matchingPods := 0
												for podName := range allPods {
													if strings.HasPrefix(podName, deploy.Name+"-") {
														matchingPods++
													}
												}
												if int32(matchingPods) < expectedReplicas {
													missingPods = append(missingPods, fmt.Sprintf("%s (want %d, have %d)", deploy.Name, expectedReplicas, matchingPods))
												}
											}
										}
									}

									// Show namespace summary first
									pterm.Info.Printf("  Namespace %s summary:\n", ns)
									if len(allPods) == 0 {
										pterm.Warning.Println("    No pods found in namespace!")
									} else {
										// Group pods by status
										readyPods := []string{}
										pendingPods := []string{}
										otherPods := []string{}
										for podName, status := range allPods {
											if strings.Contains(status, "Running") && strings.Contains(status, "Ready") && !strings.Contains(status, "init:") {
												readyPods = append(readyPods, podName)
											} else if strings.Contains(status, "Pending") || strings.Contains(status, "init:") {
												pendingPods = append(pendingPods, fmt.Sprintf("%s (%s)", podName, status))
											} else {
												otherPods = append(otherPods, fmt.Sprintf("%s (%s)", podName, status))
											}
										}
										pterm.Info.Printf("    Ready: %d, Pending/Init: %d, Other: %d\n", len(readyPods), len(pendingPods), len(otherPods))
										if len(pendingPods) > 0 {
											pterm.Info.Printf("    Pending/Init pods: %v\n", pendingPods)
										}
										if len(otherPods) > 0 {
											pterm.Info.Printf("    Other pods: %v\n", otherPods)
										}
									}

									// Show missing pods (expected but not created)
									if len(missingPods) > 0 {
										pterm.Warning.Printf("    Missing pods (expected but not created): %v\n", missingPods)

										// For missing StatefulSet pods, check the StatefulSet events and PVC status
										for _, missingPod := range missingPods {
											// Skip deployment-style missing pods (they have different format)
											if strings.Contains(missingPod, "(want") {
												continue
											}
											// Extract StatefulSet name from pod name (e.g., "tactical-backend-0" -> "tactical-backend")
											parts := strings.Split(missingPod, "-")
											if len(parts) >= 2 {
												stsName := strings.Join(parts[:len(parts)-1], "-")

												pterm.Info.Printf("\n    Checking StatefulSet %s:\n", stsName)

												// Get StatefulSet status
												stsStatusArgs := m.getKubectlArgs("-n", ns, "get", "statefulset", stsName, "-o", "json")
												stsStatusResult, _ := m.executor.Execute(localCtx, "kubectl", stsStatusArgs...)
												if stsStatusResult != nil && stsStatusResult.Stdout != "" {
													var sts appsv1.StatefulSet
													if err := json.Unmarshal([]byte(stsStatusResult.Stdout), &sts); err == nil {
														pterm.Info.Printf("      Replicas: %d desired, %d ready, %d current\n",
															func() int32 {
																if sts.Spec.Replicas != nil {
																	return *sts.Spec.Replicas
																}
																return 1
															}(),
															sts.Status.ReadyReplicas, sts.Status.CurrentReplicas)
														if len(sts.Status.Conditions) > 0 {
															for _, cond := range sts.Status.Conditions {
																if cond.Status != "True" || cond.Type == "Available" {
																	pterm.Info.Printf("      Condition: %s=%s (%s)\n", cond.Type, cond.Status, cond.Message)
																}
															}
														}
													}
												}

												// Get events for the StatefulSet
												stsEventsArgs := m.getKubectlArgs("-n", ns, "get", "events", "--field-selector", "involvedObject.name="+stsName, "--sort-by=.lastTimestamp", "-o", "custom-columns=TIME:.lastTimestamp,REASON:.reason,MESSAGE:.message", "--no-headers")
												stsEventsResult, _ := m.executor.Execute(localCtx, "kubectl", stsEventsArgs...)
												if stsEventsResult != nil && strings.TrimSpace(stsEventsResult.Stdout) != "" {
													eventLines := strings.Split(strings.TrimSpace(stsEventsResult.Stdout), "\n")
													if len(eventLines) > 3 {
														eventLines = eventLines[len(eventLines)-3:]
													}
													pterm.Info.Println("      Recent Events:")
													for _, event := range eventLines {
														if event != "" {
															pterm.Info.Printf("        %s\n", event)
														}
													}
												}

												// Check PVC status for this StatefulSet
												pvcName := stsName // Common pattern: PVC name matches StatefulSet name
												pvcArgs := m.getKubectlArgs("-n", ns, "get", "pvc", pvcName, "-o", "json")
												pvcResult, _ := m.executor.Execute(localCtx, "kubectl", pvcArgs...)
												if pvcResult != nil && pvcResult.Stdout != "" {
													var pvc corev1.PersistentVolumeClaim
													if err := json.Unmarshal([]byte(pvcResult.Stdout), &pvc); err == nil {
														pterm.Info.Printf("      PVC %s: Phase=%s", pvcName, pvc.Status.Phase)
														if pvc.Status.Phase != corev1.ClaimBound {
															pterm.Warning.Printf(" (not bound!)")
														}
														pterm.Println()
													}
												} else {
													// Try volumeClaimTemplate pattern: data-{stsname}-0
													pvcName = "data-" + stsName + "-0"
													pvcArgs = m.getKubectlArgs("-n", ns, "get", "pvc", pvcName, "-o", "json")
													pvcResult, _ = m.executor.Execute(localCtx, "kubectl", pvcArgs...)
													if pvcResult != nil && pvcResult.Stdout != "" {
														var pvc corev1.PersistentVolumeClaim
														if err := json.Unmarshal([]byte(pvcResult.Stdout), &pvc); err == nil {
															pterm.Info.Printf("      PVC %s: Phase=%s", pvcName, pvc.Status.Phase)
															if pvc.Status.Phase != corev1.ClaimBound {
																pterm.Warning.Printf(" (not bound!)")
															}
															pterm.Println()
														}
													}
												}
											}
										}
									}

									if len(problemPods) == 0 && len(missingPods) == 0 {
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

			// Check if deployment is complete — ALL currently detected apps must be
			// healthy and synced (not just "ever ready"), guarded by the high-water
			// mark of the app count (see isDeploymentComplete).
			allReady := isDeploymentComplete(totalApps, currentlyReady, maxAppsSeenTotal)
			if !allReady && totalApps > 0 && totalApps < maxAppsSeenTotal && config.Verbose {
				pterm.Warning.Printf("Application count dropped: %d visible vs %d previously seen — waiting for all apps to reappear\n", totalApps, maxAppsSeenTotal)
			}

			// Update ready count for display purposes (still use everReady for progress tracking)
			if currentlyReady > maxAppsSeenReady {
				maxAppsSeenReady = currentlyReady
			}

			// Stabilization window: require multiple consecutive all-ready checks
			// to handle inter-wave gaps where next-wave apps haven't been created yet
			if allReady {
				consecutiveAllReady++
				if config.Verbose {
					pterm.Debug.Printf("All apps ready (%d/%d stabilization checks)\n", consecutiveAllReady, stabilizationChecks)
				}
				if consecutiveAllReady >= stabilizationChecks {
					spinnerMutex.Lock()
					if !spinnerStopped && spinner != nil {
						spinner.Stop()
						spinnerStopped = true
					}
					spinnerMutex.Unlock()
					pterm.Success.Println("All ArgoCD applications installed")
					return nil
				}
			} else {
				if consecutiveAllReady > 0 && config.Verbose {
					pterm.Debug.Printf("Stabilization reset: was %d/%d, app became not-ready\n", consecutiveAllReady, stabilizationChecks)
				}
				consecutiveAllReady = 0
			}
		}
	}
}

// waitForArgoCDReady waits for ArgoCD CRD and pods to be ready using native Go clients
// This reduces reliance on external kubectl binary
func (m *Manager) waitForArgoCDReady(ctx context.Context, verbose bool, skipCRDs bool) error {
	// On Windows/WSL2, always use kubectl because native Go client can't reliably
	// reach the cluster running inside WSL due to networking bridge issues
	if runtime.GOOS == "windows" {
		if verbose {
			pterm.Info.Println("Using kubectl for ArgoCD readiness check (Windows/WSL2 mode)")
		}
		return m.waitForArgoCDReadyViaKubectl(ctx, verbose, skipCRDs)
	}

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
		podList, err := m.kubeClient.CoreV1().Pods(ArgoCDNamespace).List(ctx, metav1.ListOptions{
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

		podList, err := m.kubeClient.CoreV1().Pods(ArgoCDNamespace).List(ctx, metav1.ListOptions{
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
		checkArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods",
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

	// Wait for pods to be ready using a polling approach instead of kubectl wait
	// This allows us to properly handle completed Job pods (like argocd-redis-secret-init)
	// which don't have a Ready condition but are considered "done"
	podReadyTimeout := 5 * time.Minute
	podReadyStart := time.Now()

	for time.Since(podReadyStart) < podReadyTimeout {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Get all ArgoCD pods as JSON
		podsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods",
			"-l", "app.kubernetes.io/part-of=argocd",
			"-o", "json")
		podsResult, err := m.executor.Execute(ctx, "kubectl", podsArgs...)

		if err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to list pods: %v\n", err)
			}
			time.Sleep(retryInterval)
			continue
		}

		if podsResult == nil || podsResult.Stdout == "" {
			time.Sleep(retryInterval)
			continue
		}

		var podList corev1.PodList
		if err := json.Unmarshal([]byte(podsResult.Stdout), &podList); err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to parse pod list: %v\n", err)
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
