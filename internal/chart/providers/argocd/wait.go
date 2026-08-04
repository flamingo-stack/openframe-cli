package argocd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	uispinner "github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
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

	// If the deadline is too close to meaningfully verify the applications, do
	// NOT report success. Returning nil here would mark the platform "ready"
	// while apps are still syncing (and let cleanup delete the temp values);
	// surface it as a timeout so the caller sees the truth.
	if deadline, ok := ctx.Deadline(); ok {
		if left := time.Until(deadline); left < 10*time.Second {
			return fmt.Errorf("insufficient time to verify ArgoCD applications before the deadline (%s left)", left.Round(time.Second))
		}
	}

	// Derive a cancellable context from the parent. The parent is already
	// signal-cancelled (root ExecuteContext), so Ctrl-C / SIGTERM propagates
	// here immediately — no local signal handler required.
	localCtx, localCancel := context.WithCancel(ctx)
	defer localCancel()

	// Wait for ArgoCD CRD and pods to be ready before checking applications
	if err := m.waitForArgoCDReady(localCtx, config.Verbose, config.SkipCRDs); err != nil {
		return fmt.Errorf("ArgoCD not ready: %w", err)
	}

	// Initial repo-server health check - catch issues early
	initialIssue := m.checkRepoServerHealth(localCtx, true)
	if initialIssue != nil {
		// RepoServerIssue.Message explains what is wrong (restart count, OOMKilled,
		// CrashLoopBackOff). Every caller used to discard it, so the CLI knew the
		// repo-server was crash-looping and said nothing.
		pterm.Warning.Printfln("ArgoCD repo-server: %s", initialIssue.Message)
		// If repo-server has already restarted, proactively restart it to clear any stuck state
		// This helps CI environments where the pod may have OOM'd during initial setup
		if initialIssue.Type == "resource" && initialIssue.Recoverable {
			if age, ok := m.repoServerAge(localCtx); ok && age < repoServerColdStartGrace {
				// Cold-start grace: a freshly started repo-server produces exactly
				// these symptoms while it warms up; restarting it only prolongs that.
				pterm.Info.Printfln("ArgoCD repo-server is only %s old; giving it %s to settle before considering restarts.",
					age.Round(time.Second), repoServerColdStartGrace)
			} else {
				pterm.Info.Println("Restarting the ArgoCD repo-server to clear the stuck state...")
				m.triggerRepoServerRecovery(localCtx, "")
			}
		} else if !initialIssue.Recoverable {
			pterm.Warning.Println("This is not automatically recoverable — the installation may fail. " +
				"Check resources with: kubectl describe pods -n argocd -l app.kubernetes.io/component=repo-server")
		}
	}
	// Show initial verbose info if enabled
	if config.Verbose {
		pterm.Info.Println("Starting ArgoCD application synchronization...")
		pterm.Debug.Println("  - Waiting for applications to be created by app-of-apps")
		pterm.Debug.Println("  - Each application must reach Healthy + Synced status")
		pterm.Debug.Println("  - Progress updates every 10 seconds in verbose mode")
	}

	// Display: the live dashboard (interactive terminal, non-verbose) shows
	// the progress bar and per-app health in place; otherwise the spinner
	// carries a one-line summary; --silent gets a single info line. Exactly
	// one of dash/spinner is active — every site below guards on nil.
	var spinner *uispinner.Spinner
	var dash *appDashboard
	if !config.Silent && !config.Verbose {
		dash = newAppDashboard()
	}
	switch {
	case dash != nil:
		dash.Update(0, 0, nil)
	case !config.Silent:
		spinner = uispinner.New().WithTimer()
		spinner.Start("Installing ArgoCD applications...")
	default:
		// In non-interactive mode, just show a simple info message
		pterm.Info.Println("Installing ArgoCD applications...")
	}

	var spinnerMutex sync.Mutex
	spinnerStopped := false

	// waitNote routes one-off in-wait announcements: pinned under the live
	// dashboard when it is active (a plain print would be visually swallowed
	// by the area redraw within 2s), ordinary silence-aware prints otherwise.
	waitNote := func(styled string) {
		if dash != nil {
			dash.Note(styled)
			return
		}
		pterm.DefaultBasicText.Println(styled)
	}

	// Function to stop spinner safely
	stopSpinner := func() {
		dash.Stop()
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
	timeout := m.waitTimeout
	if timeout <= 0 {
		timeout = defaultAppWaitTimeout()
	}
	// Cap the budget to the caller's context deadline, with margin: the install
	// path runs the WHOLE flow (ArgoCD install, clone, app-of-apps, this wait)
	// under one 60m deadline, and this wait's own default is also 60m measured
	// from a strictly later point — so the outer deadline always expired first
	// and the diagnostic timeoutError below (not-ready apps, per-app health)
	// was unreachable; users got a bare "context deadline exceeded".
	timeout = capToDeadline(localCtx, timeout, startTime)
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
	consecutiveAllReady := 0
	stabilizationChecks := m.StabilizationChecks
	if stabilizationChecks <= 0 {
		stabilizationChecks = 15 // default: 15 * 2s = 30s
	}

	// Track applications that have ever been ready (healthy + synced) during this session
	// Once an app is ready, it stays counted even if it temporarily goes out of sync
	everReadyApps := make(map[string]bool)

	// Stall tracking (finding N3, per-application — see stall.go).
	stall := newStallTracker()
	stragglerSyncTriggered := false
	stallHintShown := false

	// Deterministic manifest-error tracking (see fatalmanifest.go): a legacy
	// ref whose chart path does not exist at the pinned revision fails fast
	// instead of riding the full timeout.
	fatalManifest := newFatalManifestTracker()

	// Terminally-Degraded app tracking (see degraded.go): an app that stays
	// Degraded+Synced with a genuinely-stuck pod (CrashLoop / ImagePull) fails
	// fast with that pod's crash logs, instead of hanging to the timeout.
	degraded := newDegradedTracker()

	// Repo-server issue tracking for recovery logic
	repoServerRecoveryAttempts := 0
	maxRepoServerRecoveryAttempts := 3 // Increased from 2 for CI resilience
	lastRepoServerDiagnostic := time.Time{}
	repoServerDiagnosticInterval := 2 * time.Minute  // Reduced from 3 min for faster CI recovery
	appsWithRepoServerIssues := make(map[string]int) // Track consecutive failures per app
	lastRepoServerResourceCheck := time.Now()
	repoServerResourceCheckInterval := 30 * time.Second // Reduced from 1 min for faster issue detection
	lastRepoServerMessage := ""                         // de-duplicates the repeated diagnosis line

	// Periodic-output throttles. These are time-based on purpose: the previous
	// code gated on `int(elapsed.Seconds())%10 == 0`, but the status check runs
	// every checkInterval (2s), so whether elapsed ever landed on an exact
	// multiple of 10 was luck. A skipped tick silently skipped that whole cycle.
	lastProgressPrint := time.Now()
	lastUnknownWarn := time.Time{}
	lastStuckSummary := time.Time{}

	// Last observed state, kept so the timeout error can name the applications
	// that never became ready. The loop had this all along and threw it away:
	// "timeout waiting for ArgoCD applications after 1h0m0s" told the user
	// nothing about which of the apps was stuck, or what to run next.
	var lastNotReadyApps []string    // decorated "name (Health: X)" labels, for the list
	var lastNotReadyNames []string   // bare names, for the kubectl example
	var lastNotReadyDetails []string // per-app health messages, for the timeout diagnostic
	var lastAppsSeen []Application   // full last parse, for the timeout workload diagnosis
	lastReadyCount, lastTotalApps := 0, 0
	heartbeatLastReady := 0
	// The spinner already animates for interactive users, so the textual line is
	// mainly a heartbeat for logs and CI; verbose users want it more often.
	progressPrintInterval := 30 * time.Second
	if config.Verbose {
		progressPrintInterval = 10 * time.Second
	}

	// Main loop
	for {
		select {
		case <-localCtx.Done():
			// Last-known-state line for the sequential modes: a Ctrl+C or a
			// cancelled CI job records where the install stood instead of
			// ending on a bare "operation cancelled".
			if dash == nil && lastTotalApps > 0 {
				line := fmt.Sprintf("interrupted at %d/%d applications ready", lastReadyCount, lastTotalApps)
				if len(lastNotReadyNames) > 0 {
					line += fmt.Sprintf(" · pending: %s", strings.Join(lastNotReadyNames, ", "))
				}
				pterm.DefaultBasicText.Println(line)
			}
			return fmt.Errorf("operation cancelled: %w", localCtx.Err())
		case <-ticker.C:
			// Check timeout
			if time.Since(startTime) > timeout {
				dash.Fail(fmt.Sprintf("Timeout after %v", timeout))
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil {
					spinner.Fail(fmt.Sprintf("Timeout after %v", timeout))
					spinnerStopped = true
				}
				spinnerMutex.Unlock()
				// Deep workload diagnosis for the apps that never became ready:
				// failing pods with reasons, their last crash-log lines, and
				// recent warning events. This is the same inspection the
				// Degraded fail-fast runs — the timeout path used to skip it,
				// leaving only app-level health lines. Bounded (capToDeadline
				// reserved a margin for exactly this) and best-effort.
				workloadDiag := ""
				if stuck := notReadyApplications(lastAppsSeen, maxAppsDiagnosedAtTimeout); len(stuck) > 0 {
					pterm.Info.Printf("Collecting diagnostics for %d not-ready application(s)...\n", len(stuck))
					workloadDiag, _ = m.diagnoseFailingApps(localCtx, stuck)
				}
				return timeoutError(timeout, lastReadyCount, lastTotalApps, lastNotReadyApps, lastNotReadyNames, lastNotReadyDetails, workloadDiag)
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

					// Add exponential backoff delay between failures to avoid hammering
					// WSL. Ctx-aware: a plain Sleep here held Ctrl+C hostage for up to
					// 10s per failure — exactly during the flaky-cluster episodes where
					// users reach for it.
					backoffDelay := time.Duration(consecutiveFailures) * 2 * time.Second
					if backoffDelay > 10*time.Second {
						backoffDelay = 10 * time.Second
					}
					select {
					case <-localCtx.Done():
						return fmt.Errorf("operation cancelled: %w", localCtx.Err())
					case <-time.After(backoffDelay):
					}
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

					// Add backoff delay between failures. Ctx-aware for the same
					// reason as the connectivity backoff above.
					backoffDelay := time.Duration(consecutiveFailures) * 2 * time.Second
					if backoffDelay > 10*time.Second {
						backoffDelay = 10 * time.Second
					}
					select {
					case <-localCtx.Done():
						return fmt.Errorf("operation cancelled: %w", localCtx.Err())
					case <-time.After(backoffDelay):
					}
				}

				// Retry on other errors (with normal interval via lastCheck)
				continue
			}

			// Reset consecutive failures on successful query
			if consecutiveFailures > 0 {
				waitNote(pterm.Success.Sprint("Application queries restored"))
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
			lastNotReadyApps, lastReadyCount, lastTotalApps = notReadyApps, currentlyReady, totalApps
			lastNotReadyNames = assess.notReadyNames
			lastNotReadyDetails = notReadyDiagnostics(apps)
			lastAppsSeen = apps

			// Fail fast on deterministic manifest errors (see fatalmanifest.go):
			// once an app has shown the same "content missing at this revision"
			// ComparisonError past the persistence thresholds, no amount of
			// waiting or repo-server recovery changes the outcome. Checked before
			// stall/recovery handling so neither wastes effort on a lost cause.
			// One timestamp for all observe calls so recorded "since" values and
			// staleness checks use the same tick.
			now := time.Now()
			if fatal := fatalManifest.observe(apps, now); len(fatal) > 0 {
				dash.Fail("Applications cannot render manifests from the deployed revision")
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil {
					spinner.Fail("Applications cannot render manifests from the deployed revision")
					spinnerStopped = true
				}
				spinnerMutex.Unlock()

				requestedRef := ""
				if config.AppOfApps != nil {
					requestedRef = config.AppOfApps.GitHubBranch
				}
				return fatalManifestError(requestedRef, fatal)
			}

			// Terminally-Degraded fail-fast: an app that has stayed Degraded+Synced
			// long enough is only aborted once we confirm a genuinely-stuck pod
			// (CrashLoop / ImagePull / repeated crashes) — so a slow-but-recovering
			// workload is never cut off. The error carries the pod's crash logs and
			// events, so it says WHY, not just that it hung.
			if cand := degraded.observe(apps, now); len(cand) > 0 {
				if diag, stuck := m.diagnoseFailingApps(localCtx, cand); len(stuck) > 0 {
					dash.Fail("An application is Degraded with a workload that will not recover")
					spinnerMutex.Lock()
					if !spinnerStopped && spinner != nil {
						spinner.Fail("An application is Degraded with a workload that will not recover")
						spinnerStopped = true
					}
					spinnerMutex.Unlock()
					// Only the apps whose OWN pods are terminally stuck — not
					// every candidate that happens to share their namespace.
					return degradedAppError(stuck, diag)
				}
			}

			// Stall handling (finding N3, per-application): an app that has sat
			// OutOfSync-but-Healthy, bit-for-bit identical, for stallAfter will not
			// move on its own (autoSync off). Judged per-app so a noisy neighbour
			// flapping status can't keep resetting the clock on a stuck app (V5).
			// On the upgrade path, sync exactly those stragglers once; otherwise
			// print an actionable hint once instead of burning the full timeout.
			stall.observe(apps, now)
			if stragglers := stall.stalledStragglers(apps, now); len(stragglers) > 0 {
				// The two branches are chosen by MODE, not by whether the sync has
				// already fired: on the upgrade path (SyncStragglersOnStall) the hint
				// must never print — it tells the user to run `app upgrade --sync`,
				// which is exactly the path they are already on. Gating the hint on
				// !stragglerSyncTriggered instead would print that contradictory
				// advice on every stall tick after the one-shot sync.
				stallNote := waitNote
				if config.SyncStragglersOnStall {
					if !stragglerSyncTriggered {
						stragglerSyncTriggered = true
						stallNote(pterm.Warning.Sprintf("No progress for %s; triggering sync of %d OutOfSync application(s): %v",
							stallAfter.Round(time.Second), len(stragglers), stragglers))
						patched, failedCount, syncErr := m.syncApplicationsByName(localCtx, stragglers, false)
						if failedCount > 0 {
							stallNote(pterm.Warning.Sprintf("Straggler sync: %d triggered, %d failed (first error: %v)", patched, failedCount, syncErr))
						}
					}
				} else if !stallHintShown {
					stallHintShown = true
					stallNote(pterm.Warning.Sprintf("No progress for %s; %d application(s) are OutOfSync and may have auto-sync disabled: %v",
						stallAfter.Round(time.Second), len(stragglers), stragglers))
					stallNote(pterm.Info.Sprint("They will not sync on their own — run `openframe app upgrade --sync` (or sync them in ArgoCD) to roll them out."))
				}
			}

			elapsed := time.Since(startTime)

			// Progress belongs in the spinner text, not behind --verbose. Without
			// this the default experience was a static "Installing ArgoCD
			// applications..." for up to the full 60m timeout, with no way to tell
			// a working install from a wedged one.
			dash.Update(currentlyReady, totalApps, apps)
			if totalApps > 0 {
				spinnerMutex.Lock()
				if !spinnerStopped && spinner != nil {
					percent := float64(currentlyReady) / float64(totalApps) * 100
					spinner.UpdateText(fmt.Sprintf("Installing ArgoCD applications... %d/%d ready (%.0f%%) [%s]",
						currentlyReady, totalApps, percent, elapsed.Round(time.Second)))
				}
				spinnerMutex.Unlock()
			}

			// Repo-server recovery and issue classification used to sit INSIDE the
			// `if config.Verbose` block, so a user who did not pass --verbose never
			// got the recovery at all: a wedged repo-server simply burned the whole
			// timeout. Recovery is a corrective action, not a diagnostic — it runs
			// regardless of verbosity, and announces itself when it fires.
			if totalApps > 0 && len(notReadyApps) > 0 {
				unknownApps, appsWithConditionErrors := classifyAppIssues(apps, appsWithRepoServerIssues)

				if len(appsWithConditionErrors) > 0 && elapsed > 2*time.Minute {
					if time.Since(lastRepoServerResourceCheck) >= repoServerResourceCheckInterval {
						lastRepoServerResourceCheck = time.Now()
						if issue := m.checkRepoServerHealth(localCtx, false); issue != nil && issue.Message != lastRepoServerMessage {
							// Print each distinct diagnosis once: the check runs every
							// 30s and would otherwise repeat the same line forever.
							lastRepoServerMessage = issue.Message
							pterm.Warning.Printfln("ArgoCD repo-server: %s", issue.Message)
						}
					}

					for _, app := range appsWithConditionErrors {
						consecutiveIssues := appsWithRepoServerIssues[app.Name]

						// After 2 consecutive checks with repo-server issues, recover.
						if consecutiveIssues >= 2 && time.Since(lastRepoServerDiagnostic) >= repoServerDiagnosticInterval {
							lastRepoServerDiagnostic = time.Now()

							// Cold-start grace: never restart a repo-server that is still
							// warming up — on a fresh install the first manifest renders
							// legitimately fail with the same "connection refused"/EOF
							// conditions this detector keys on, and each restart re-zeroes
							// rendering for every app, restarting the carousel. This also
							// spaces successive recovery attempts at least the grace apart.
							if age, ok := m.repoServerAge(localCtx); ok && age < repoServerColdStartGrace {
								pterm.Info.Printfln("ArgoCD repo-server is only %s old; waiting for it to settle (%s grace) before considering a restart.",
									age.Round(time.Second), repoServerColdStartGrace)
								break
							}

							if repoServerRecoveryAttempts < maxRepoServerRecoveryAttempts {
								repoServerRecoveryAttempts++
								// Restarting the repo-server takes the apps through a
								// visible wobble; say why, or it reads as a new failure.
								pterm.Warning.Printfln("ArgoCD repo-server looks stuck (application %q cannot fetch its manifests); restarting it (attempt %d/%d)",
									app.Name, repoServerRecoveryAttempts, maxRepoServerRecoveryAttempts)
								if m.triggerRepoServerRecovery(localCtx, app.Name) {
									pterm.Info.Println("ArgoCD repo-server restarted; applications will re-sync shortly.")
									delete(appsWithRepoServerIssues, app.Name)
									// The restarted repo-server has a cold manifest cache, so
									// every app stuck in Unknown (not just the trigger) needs a
									// HARD refresh to regenerate — otherwise they ride the wait
									// out to its timeout. triggerRepoServerRecovery already
									// hard-refreshed app.Name; cover the rest.
									if refreshed := m.hardRefreshApplications(localCtx, appNames(unknownApps)); refreshed > 0 {
										waitNote(pterm.Info.Sprintf("Hard-refreshed %d application(s) stuck in Unknown.", refreshed))
									}
								} else {
									waitNote(pterm.Warning.Sprint("Could not restart the ArgoCD repo-server; continuing to wait."))
								}
							} else if repoServerRecoveryAttempts == maxRepoServerRecoveryAttempts {
								repoServerRecoveryAttempts++ // prevent repeated attempts
								waitNote(pterm.Warning.Sprintf("ArgoCD repo-server did not recover after %d restarts; continuing to wait for the timeout.",
									maxRepoServerRecoveryAttempts))
							}
							break // Only recover one app at a time
						}
					}
				}

				// Applications stuck in Unknown for 5 minutes usually mean the ArgoCD
				// controller is unhealthy or git is unreachable. Warn at any verbosity
				// (throttled); the per-application dump stays behind --verbose.
				if len(unknownApps) > 0 && elapsed > 5*time.Minute && time.Since(lastUnknownWarn) >= 5*time.Minute {
					lastUnknownWarn = time.Now()
					waitNote(pterm.Warning.Sprintf("%d application(s) have 'Unknown' status after %s. Possible causes: controller pod not ready, git repository unreachable, or resource constraints.",
						len(unknownApps), elapsed.Round(time.Second)))
					if config.Verbose {
						describeUnknownApps(unknownApps)
					} else if dash == nil {
						pterm.Info.Println("  Re-run with --verbose for per-application detail.")
					}
				}

				// A concise summary of stuck applications, every 5 minutes after the
				// 7-minute mark (in-memory status; no kubectl resource dump). The
				// dashboard already shows exactly this, live — sequential modes only.
				if dash == nil && elapsed > 7*time.Minute && time.Since(lastStuckSummary) >= 5*time.Minute {
					lastStuckSummary = time.Now()
					for _, app := range apps {
						if app.Health != ArgoCDHealthHealthy && app.Health != ArgoCDHealthMissing {
							line := fmt.Sprintf("  Stuck app %s: health=%s sync=%s", app.Name, app.Health, app.Sync)
							if app.Condition != "" {
								line += " condition=" + app.Condition
							}
							pterm.Warning.Println(line)
						}
					}
				}
			}

			// Textual progress heartbeat — what a log, a piped session, --plain,
			// or a --verbose run sees instead of the dashboard. Each line is
			// self-sufficient for scrollback reading: wall clock (correlate
			// with cluster events), the delta since the previous beat (a zero
			// delta screams "stalled" without diffing counts by eye), and the
			// pending apps WITH their health, not just names. Never alongside
			// the live dashboard — a foreign print every 30s would garble the
			// area redraw, and the dashboard already shows all of it.
			if dash == nil && totalApps > 0 && time.Since(lastProgressPrint) >= progressPrintInterval {
				lastProgressPrint = time.Now()
				delta := currentlyReady - heartbeatLastReady
				heartbeatLastReady = currentlyReady
				pterm.Info.Printf("[%s] apps %d/%d ready (%+d since last check) · elapsed %s\n",
					time.Now().Format("15:04:05"), currentlyReady, totalApps, delta, elapsed.Round(time.Second))
				if p := pendingSummary(apps, 6); p != "" {
					pterm.Info.Printf("  pending: %s\n", p)
				}
				if config.Verbose && len(healthyApps) > 0 && len(healthyApps) <= 5 {
					pterm.Debug.Printf("  Recently completed: %v\n", healthyApps)
				}
			}

			// Check if deployment is complete — ALL currently detected apps must be
			// healthy and synced (not just "ever ready"), guarded by the high-water
			// mark of the app count AND the planned count from the app-of-apps
			// (see isDeploymentComplete).
			allReady := isDeploymentComplete(totalApps, currentlyReady, maxAppsSeenTotal, totalAppsExpected)
			if !allReady && totalApps > 0 && totalApps < maxAppsSeenTotal && config.Verbose {
				pterm.Warning.Printf("Application count dropped: %d visible vs %d previously seen — waiting for all apps to reappear\n", totalApps, maxAppsSeenTotal)
			}
			if !allReady && config.Verbose && totalAppsExpected > 0 && totalApps < totalAppsExpected && currentlyReady == totalApps && totalApps > 0 {
				pterm.Debug.Printf("All %d visible apps ready, but the app-of-apps plans %d — waiting for the remaining Application CRs\n", totalApps, totalAppsExpected)
			}

			// Stabilization window: require multiple consecutive all-ready checks
			// to handle inter-wave gaps where next-wave apps haven't been created yet
			if allReady {
				consecutiveAllReady++
				if config.Verbose {
					pterm.Debug.Printf("All apps ready (%d/%d stabilization checks)\n", consecutiveAllReady, stabilizationChecks)
				}
				if consecutiveAllReady >= stabilizationChecks {
					// Everything is Healthy+Synced — but "ready" is not "correct".
					// If a ref was requested, confirm ArgoCD is actually tracking it
					// before declaring success; a legacy branch's chart silently
					// deploys main and this is the only place that catches it (V3).
					// Decide the spinner's final state from the outcome: FAIL on a
					// mismatch (matching the timeout path), a neutral Stop on success
					// — never a neutral stop immediately before returning an error.
					var mm []refMismatch
					if config.AppOfApps != nil {
						mm = verifyRefPinning(apps, config.AppOfApps.GitHubRepo, config.AppOfApps.GitHubBranch)
					}

					spinnerMutex.Lock()
					if !spinnerStopped && spinner != nil {
						if len(mm) > 0 {
							spinner.Fail("Deployed ref does not match the requested ref")
						} else {
							spinner.Stop()
						}
						spinnerStopped = true
					}
					spinnerMutex.Unlock()

					if len(mm) > 0 {
						dash.Fail("Deployed ref does not match the requested ref")
						return refMismatchError(config.AppOfApps.GitHubBranch, mm)
					}

					if dash != nil {
						// The dashboard's success line carries the slowest-apps
						// timings it collected along the way.
						dash.FinishSuccess("All ArgoCD applications installed")
					} else {
						pterm.Success.Println("All ArgoCD applications installed")
					}
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
	// On Windows the cluster lives in WSL2 and must be reached from inside WSL.
	if err := platform.WSLClusterHint("wait for ArgoCD to be ready"); err != nil {
		return err
	}

	maxRetries := 100 // 100 retries * 3 seconds = 5 minutes max
	retryInterval := 3 * time.Second

	// Initialize Kubernetes clients for native API access
	if err := m.initKubernetesClients(); err != nil {
		return fmt.Errorf("failed to initialize the Kubernetes client: %w", err)
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
				// The pod-wait path below prints diagnostics on timeout; this one
				// returned a bare sentence with nothing to act on. The CRD is
				// installed by the ArgoCD chart, so its absence means the release
				// itself never landed.
				return fmt.Errorf("timeout waiting for the ArgoCD CRD applications.argoproj.io to appear.\n"+
					"The CRD is installed by the ArgoCD Helm release, so this usually means the release failed.\n"+
					"Check it with: helm status %s -n %s\n"+
					"And the controller pods with: kubectl get pods -n %s",
					ArgoCDReleaseName, ArgoCDNamespace, ArgoCDNamespace)
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
	// Time-based throttle, not `elapsed%15 == 0`: the loop advances by 3s plus
	// API latency, so exact multiples of 15 land by luck (see the main loop's
	// periodic-output throttles for the same fix).
	lastPodWaitNote := time.Time{}

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

		if verbose && time.Since(lastPodWaitNote) >= 15*time.Second {
			lastPodWaitNote = time.Now()
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

// describeUnknownApps prints the per-application detail for applications stuck
// in Unknown: source, condition, operation state, health message, and last
// reconciliation. It is the --verbose expansion of the one-line warning the
// wait loop emits; the condition line is usually the one that explains it.
func describeUnknownApps(unknownApps []Application) {
	for _, app := range unknownApps {
		pterm.Warning.Printf("\n  --- %s (Health: %s, Sync: %s) ---\n", app.Name, app.Health, app.Sync)

		if app.RepoURL != "" {
			pterm.Info.Printf("    Source: %s", app.RepoURL)
			if app.Path != "" {
				pterm.DefaultBasicText.Printf(" path=%s", app.Path)
			}
			if app.TargetRevision != "" {
				pterm.DefaultBasicText.Printf(" revision=%s", app.TargetRevision)
			}
			pterm.DefaultBasicText.Println()
		}

		if app.Condition != "" {
			condType := app.ConditionType
			if condType == "" {
				condType = "Error"
			}
			pterm.Warning.Printf("    %s: %s\n", condType, app.Condition)
		}

		if app.OperationPhase != "" {
			pterm.Info.Printf("    Operation: %s", app.OperationPhase)
			if app.OperationMessage != "" {
				pterm.DefaultBasicText.Printf(" - %s", app.OperationMessage)
			}
			pterm.DefaultBasicText.Println()
		}

		if app.HealthMessage != "" {
			pterm.Info.Printf("    Health details: %s\n", app.HealthMessage)
		}

		if app.ReconciledAt != "" {
			pterm.Info.Printf("    Last reconciled: %s\n", app.ReconciledAt)
		} else {
			pterm.Warning.Println("    Not yet reconciled (ArgoCD hasn't processed this app)")
		}
	}
}

// maxAppsInTimeoutError bounds the application list in the timeout message: a
// large platform can leave dozens pending, and an unbounded list buries the
// next-step hint that follows it.
const maxAppsInTimeoutError = 10

// maxAppsDiagnosedAtTimeout bounds the deep workload diagnosis at timeout: each
// diagnosed app costs a pod list, a log stream per failing container, and an
// events list, and the whole gather must fit the margin capToDeadline reserved.
const maxAppsDiagnosedAtTimeout = 5

// notReadyApplications returns up to limit applications that are not
// Healthy+Synced, for the timeout workload diagnosis.
func notReadyApplications(apps []Application, limit int) []Application {
	var out []Application
	for _, app := range apps {
		if app.Health == ArgoCDHealthHealthy && app.Sync == ArgoCDSyncSynced {
			continue
		}
		out = append(out, app)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// timeoutError builds the error returned when the wait budget is exhausted.
//
// The old message was "timeout waiting for ArgoCD applications after 1h0m0s" —
// true, and useless: the loop knew exactly which applications never became
// ready and discarded that. This names them and points at the command that
// shows why.
//
// notReadyLabels are decorated "name (Health: X)" strings for the human list;
// notReadyNames are the BARE application names for the kubectl example. They
// must be kept separate: feeding a decorated label into `kubectl describe
// application` produced `kubectl describe application argocd-apps (Health:
// Progressing) -n argocd`, which is not a runnable command.
func timeoutError(timeout time.Duration, ready, total int, notReadyLabels, notReadyNames, notReadyDetails []string, workloadDiag string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "timeout after %s waiting for ArgoCD applications", timeout)
	if total > 0 {
		fmt.Fprintf(&b, " (%d/%d ready)", ready, total)
	}

	if len(notReadyLabels) > 0 {
		shown := notReadyLabels
		suffix := ""
		if len(shown) > maxAppsInTimeoutError {
			suffix = fmt.Sprintf(" (and %d more)", len(shown)-maxAppsInTimeoutError)
			shown = shown[:maxAppsInTimeoutError]
		}
		fmt.Fprintf(&b, "; still not ready: %s%s", strings.Join(shown, ", "), suffix)
	}

	// Per-app health messages and unhealthy child resources say WHY (e.g. the
	// exact Deployment that missed its progress deadline inside a Degraded app),
	// so the timeout is actionable rather than opaque.
	if len(notReadyDetails) > 0 {
		b.WriteString("\nWhy they are not ready:")
		for _, d := range notReadyDetails {
			b.WriteString("\n" + d)
		}
	}

	// The pod-level view: failing containers with reasons, their last crash-log
	// lines, and recent warning events — gathered at the moment of timeout.
	if d := strings.TrimSpace(workloadDiag); d != "" {
		b.WriteString("\nFailing workloads (pod errors, last logs, warning events):")
		b.WriteString("\n" + indentLines(d, "  "))
	}

	b.WriteString("\nInspect them with: kubectl get applications -n argocd")
	if len(notReadyNames) > 0 {
		fmt.Fprintf(&b, "\nDetails for one: kubectl describe application %s -n argocd", notReadyNames[0])
	}
	if len(notReadyDetails) > 0 || strings.TrimSpace(workloadDiag) != "" {
		// The per-app health messages above ARE the diagnosis; suppress the
		// generic handler's pattern-matched hint (it misfires on their content).
		return selfDiagnosedError(b.String())
	}
	return fmt.Errorf("%s", b.String())
}

// capToDeadline shrinks a wait budget so it elapses BEFORE the context
// deadline, leaving a margin to gather and render the timeout diagnostic.
// Returns the configured budget unchanged when the context has no deadline or
// the deadline is farther away. A (nearly) exhausted deadline yields 0 — the
// wait times out on its first tick rather than riding into ctx.Err().
func capToDeadline(ctx context.Context, configured time.Duration, from time.Time) time.Duration {
	const margin = time.Minute
	dl, ok := ctx.Deadline()
	if !ok {
		return configured
	}
	remaining := dl.Sub(from) - margin
	if remaining < 0 {
		return 0
	}
	if remaining < configured {
		return remaining
	}
	return configured
}

// defaultAppWaitTimeout is the install/bootstrap application-wait budget: 60m,
// overridable via OPENFRAME_APP_WAIT_TIMEOUT (a Go duration, e.g. "35m"). A CI
// job with a shorter step cap sets it so the CLI hits its OWN timeout — and
// prints its diagnostic — before the job is killed opaquely. The upgrade path
// is unaffected: it sets an explicit WithWaitTimeout (>0), so this default
// branch never runs there.
func defaultAppWaitTimeout() time.Duration {
	const fallback = 60 * time.Minute
	if v := strings.TrimSpace(os.Getenv("OPENFRAME_APP_WAIT_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}

// maxResourcesPerAppInDiagnostics bounds the unhealthy-children list per app:
// one wedged node can degrade dozens of resources at once, and the point is to
// name the failing pieces, not to inventory the cluster.
const maxResourcesPerAppInDiagnostics = 8

// notReadyDiagnostics returns one diagnostic block per not-ready application:
// ArgoCD's app-level health message, plus the unhealthy CHILD resources with
// their per-resource health messages — which is what names the exact
// Deployment/StatefulSet/child-Application that is failing inside a Degraded
// app (the app-level message is frequently empty). Trimmed to stay readable.
func notReadyDiagnostics(apps []Application) []string {
	const maxMsg = 300
	var out []string
	for _, app := range apps {
		if app.Health == ArgoCDHealthHealthy && app.Sync == ArgoCDSyncSynced {
			continue
		}
		line := fmt.Sprintf("  - %s: health=%s sync=%s", app.Name, app.Health, app.Sync)
		if msg := strings.TrimSpace(app.HealthMessage); msg != "" {
			if len(msg) > maxMsg {
				msg = msg[:maxMsg] + "…"
			}
			line += "\n      " + strings.ReplaceAll(msg, "\n", "\n      ")
		}
		shown := app.UnhealthyResources
		suffix := ""
		if len(shown) > maxResourcesPerAppInDiagnostics {
			suffix = fmt.Sprintf("\n      … and %d more unhealthy resource(s)", len(shown)-maxResourcesPerAppInDiagnostics)
			shown = shown[:maxResourcesPerAppInDiagnostics]
		}
		for _, res := range shown {
			resLine := fmt.Sprintf("\n      %s %s", res.Kind, res.Name)
			if res.Namespace != "" {
				resLine = fmt.Sprintf("\n      %s %s/%s", res.Kind, res.Namespace, res.Name)
			}
			resLine += ": " + res.Health
			if msg := strings.TrimSpace(res.Message); msg != "" {
				if len(msg) > maxMsg {
					msg = msg[:maxMsg] + "…"
				}
				resLine += " — " + strings.ReplaceAll(msg, "\n", " ")
			}
			line += resLine
		}
		line += suffix
		out = append(out, line)
	}
	return out
}
