package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
)

// printArgoCDPodDiagnostics prints diagnostic information about ArgoCD pods when they fail to become ready
func (m *Manager) printArgoCDPodDiagnostics(ctx context.Context) {
	pterm.Warning.Println("ArgoCD pods failed to become ready. Collecting diagnostics...")

	// First check Helm release status to understand if ArgoCD was installed correctly
	helmStatusArgs := []string{"status", ArgoCDReleaseName, "-n", ArgoCDNamespace}
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
	deployArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "deployments", "-o", "wide")
	deployResult, _ := m.executor.Execute(ctx, "kubectl", deployArgs...)
	if deployResult != nil && deployResult.Stdout != "" {
		pterm.Info.Println("\nArgoCD deployments:")
		for _, line := range strings.Split(strings.TrimSpace(deployResult.Stdout), "\n") {
			pterm.Info.Printf("  %s\n", line)
		}
	}

	// Get all pods in argocd namespace with their status
	podArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods", "-o", "wide")
	podResult, _ := m.executor.Execute(ctx, "kubectl", podArgs...)
	if podResult != nil && podResult.Stdout != "" {
		pterm.Info.Println("ArgoCD pods status:")
		for _, line := range strings.Split(strings.TrimSpace(podResult.Stdout), "\n") {
			pterm.Info.Printf("  %s\n", line)
		}
	}

	// Get pods that are not ready and show their details
	// Use --field-selector instead of jsonpath to avoid Windows WSL escaping issues
	notReadyArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods", "--field-selector=status.phase!=Running", "-o", "name")
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
	runningPodsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods", "--field-selector=status.phase=Running", "-o", "json")
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
		podArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pod", podName, "-o", "json")
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
		eventsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "events",
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
		logsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "logs", podName, "--tail=10", "--all-containers=true")
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
		nsEventsArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "events",
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

// checkClusterConnectivity performs a quick health check on the Kubernetes cluster
func (m *Manager) checkClusterConnectivity(ctx context.Context, verbose bool) error {
	// Use kubectl cluster-info as a quick connectivity check
	clusterInfoArgs := m.getKubectlArgs("cluster-info")
	result, err := m.executor.Execute(ctx, "kubectl", clusterInfoArgs...)

	if err != nil {
		return fmt.Errorf("kubectl execution failed: %w", err)
	}

	if result.ExitCode != 0 {
		errMsg := result.Stderr
		if errMsg == "" {
			errMsg = result.Stdout
		}
		return fmt.Errorf("cluster unreachable (exit code %d): %s", result.ExitCode, errMsg)
	}

	if verbose {
		pterm.Debug.Println("Cluster connectivity check passed")
	}

	return nil
}

// printClusterDiagnostics prints diagnostic information when the cluster becomes unreachable
func (m *Manager) printClusterDiagnostics(ctx context.Context) {
	pterm.Error.Println("Cluster became unreachable. Collecting diagnostics...")

	// Check if we're on Windows/WSL
	if runtime.GOOS == "windows" {
		pterm.Info.Println("\n=== WSL/Docker Diagnostics ===")

		// Check WSL status
		wslResult, _ := m.executor.Execute(ctx, "wsl", "--list", "--verbose")
		if wslResult != nil && wslResult.Stdout != "" {
			pterm.Info.Println("WSL distributions:")
			pterm.Println(wslResult.Stdout)
		}

		// === Resource Checks for Windows/WSL ===
		pterm.Info.Println("\n=== System Resources ===")

		// Check memory usage in WSL
		memResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"free -h 2>/dev/null || echo 'Could not get memory info'")
		if memResult != nil && memResult.Stdout != "" {
			pterm.Info.Println("Memory usage (WSL):")
			pterm.Println(memResult.Stdout)
		}

		// Check disk space in WSL
		diskResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"df -h / /tmp 2>/dev/null | head -5 || echo 'Could not get disk info'")
		if diskResult != nil && diskResult.Stdout != "" {
			pterm.Info.Println("Disk space (WSL):")
			pterm.Println(diskResult.Stdout)
		}

		// Check Docker system resources
		dockerStatsResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"sudo docker stats --no-stream --format 'table {{.Name}}\\t{{.CPUPerc}}\\t{{.MemUsage}}\\t{{.MemPerc}}' 2>/dev/null | head -10 || echo 'Could not get Docker stats'")
		if dockerStatsResult != nil && dockerStatsResult.Stdout != "" {
			pterm.Info.Println("Docker container resource usage:")
			pterm.Println(dockerStatsResult.Stdout)
		}

		// Check Docker system info (total resources)
		dockerInfoResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"sudo docker info 2>/dev/null | grep -E '(CPUs|Total Memory|Docker Root Dir)' || echo 'Could not get Docker info'")
		if dockerInfoResult != nil && dockerInfoResult.Stdout != "" {
			pterm.Info.Println("Docker system info:")
			pterm.Println(dockerInfoResult.Stdout)
		}

		// Check for OOM kills
		oomResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"dmesg 2>/dev/null | grep -i 'out of memory\\|oom\\|killed process' | tail -5 || echo 'No OOM events found (or dmesg not accessible)'")
		if oomResult != nil && oomResult.Stdout != "" {
			pterm.Info.Println("Recent OOM events:")
			pterm.Println(oomResult.Stdout)
		}

		pterm.Info.Println("\n=== Container Status ===")

		// Check if Docker is running inside WSL
		dockerResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", "sudo docker ps -a --format 'table {{.Names}}\\t{{.Status}}\\t{{.Size}}' 2>&1 || echo 'Docker not accessible'")
		if dockerResult != nil && dockerResult.Stdout != "" {
			pterm.Info.Println("Docker containers in WSL:")
			pterm.Println(dockerResult.Stdout)
		}

		// Check k3d cluster status
		k3dResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", "sudo -E k3d cluster list 2>&1 || echo 'k3d not accessible'")
		if k3dResult != nil && k3dResult.Stdout != "" {
			pterm.Info.Println("k3d clusters:")
			pterm.Println(k3dResult.Stdout)
		}

		// Check port connectivity
		portResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", "nc -zv 127.0.0.1 6550 2>&1 || echo 'Port 6550 not reachable'")
		if portResult != nil {
			pterm.Info.Println("Port 6550 connectivity:")
			output := portResult.Stdout
			if portResult.Stderr != "" {
				output = portResult.Stderr
			}
			pterm.Println(output)
		}

		// Check Docker logs for k3d server container
		dockerLogsResult, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c",
			"sudo docker logs --tail 50 k3d-"+m.clusterName+"-server-0 2>&1 || echo 'Could not get container logs'")
		if dockerLogsResult != nil && dockerLogsResult.Stdout != "" {
			pterm.Info.Println("k3d server container logs (last 50 lines):")
			pterm.Println(dockerLogsResult.Stdout)
		}
	} else {
		// Linux/macOS diagnostics
		pterm.Info.Println("\n=== Cluster Diagnostics ===")

		// === Resource Checks for Linux/macOS ===
		pterm.Info.Println("\n=== System Resources ===")

		// Check memory usage
		memResult, _ := m.executor.Execute(ctx, "bash", "-c", "free -h 2>/dev/null || vm_stat 2>/dev/null | head -10 || echo 'Could not get memory info'")
		if memResult != nil && memResult.Stdout != "" {
			pterm.Info.Println("Memory usage:")
			pterm.Println(memResult.Stdout)
		}

		// Check disk space
		diskResult, _ := m.executor.Execute(ctx, "df", "-h", "/", "/tmp", "/var")
		if diskResult != nil && diskResult.Stdout != "" {
			pterm.Info.Println("Disk space:")
			pterm.Println(diskResult.Stdout)
		}

		// Check Docker system resources
		dockerStatsResult, _ := m.executor.Execute(ctx, "docker", "stats", "--no-stream", "--format", "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}")
		if dockerStatsResult != nil && dockerStatsResult.Stdout != "" {
			pterm.Info.Println("Docker container resource usage:")
			pterm.Println(dockerStatsResult.Stdout)
		}

		// Check Docker system info
		dockerInfoResult, _ := m.executor.Execute(ctx, "bash", "-c", "docker info 2>/dev/null | grep -E '(CPUs|Total Memory|Docker Root Dir)' || echo 'Could not get Docker info'")
		if dockerInfoResult != nil && dockerInfoResult.Stdout != "" {
			pterm.Info.Println("Docker system info:")
			pterm.Println(dockerInfoResult.Stdout)
		}

		// Check for OOM kills (Linux only)
		oomResult, _ := m.executor.Execute(ctx, "bash", "-c", "dmesg 2>/dev/null | grep -i 'out of memory\\|oom\\|killed process' | tail -5 || echo 'No OOM events found'")
		if oomResult != nil && oomResult.Stdout != "" {
			pterm.Info.Println("Recent OOM events:")
			pterm.Println(oomResult.Stdout)
		}

		pterm.Info.Println("\n=== Container Status ===")

		// Check Docker status
		dockerResult, _ := m.executor.Execute(ctx, "docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Size}}")
		if dockerResult != nil && dockerResult.Stdout != "" {
			pterm.Info.Println("Docker containers:")
			pterm.Println(dockerResult.Stdout)
		}

		// Check k3d cluster status
		k3dResult, _ := m.executor.Execute(ctx, "k3d", "cluster", "list")
		if k3dResult != nil && k3dResult.Stdout != "" {
			pterm.Info.Println("k3d clusters:")
			pterm.Println(k3dResult.Stdout)
		}

		// Check Docker logs for k3d server container
		dockerLogsResult, _ := m.executor.Execute(ctx, "docker", "logs", "--tail", "50", "k3d-"+m.clusterName+"-server-0")
		if dockerLogsResult != nil && dockerLogsResult.Stdout != "" {
			pterm.Info.Println("k3d server container logs (last 50 lines):")
			pterm.Println(dockerLogsResult.Stdout)
		}
	}

	// Try to get Kubernetes node resource status if cluster is still partially accessible
	pterm.Info.Println("\n=== Kubernetes Node Resources ===")
	nodeArgs := m.getKubectlArgs("top", "nodes")
	nodeResult, _ := m.executor.Execute(ctx, "kubectl", nodeArgs...)
	if nodeResult != nil && nodeResult.Stdout != "" && nodeResult.ExitCode == 0 {
		pterm.Info.Println("Node resource usage:")
		pterm.Println(nodeResult.Stdout)
	} else {
		pterm.Warning.Println("Could not get node resource usage (cluster may be unreachable)")
	}

	// Try to get pod resource usage
	podTopArgs := m.getKubectlArgs("top", "pods", "-A", "--sort-by=memory")
	podTopResult, _ := m.executor.Execute(ctx, "kubectl", podTopArgs...)
	if podTopResult != nil && podTopResult.Stdout != "" && podTopResult.ExitCode == 0 {
		pterm.Info.Println("Top pods by memory usage:")
		// Only show top 15 pods
		lines := strings.Split(podTopResult.Stdout, "\n")
		if len(lines) > 16 {
			lines = lines[:16]
		}
		pterm.Println(strings.Join(lines, "\n"))
	}

	pterm.Info.Println("\n=== End Diagnostics ===")
}

// RepoServerIssue represents a detected issue with the ArgoCD repo-server
type RepoServerIssue struct {
	Type        string // "communication", "resource", "git", "timeout"
	Message     string
	Recoverable bool
}

// checkRepoServerHealth checks the health of the ArgoCD repo-server and returns any issues found
func (m *Manager) checkRepoServerHealth(ctx context.Context, verbose bool) *RepoServerIssue {
	// Get repo-server pod status
	podArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "get", "pods", "-l", "app.kubernetes.io/name=argocd-repo-server", "-o", "json")
	podResult, err := m.executor.Execute(ctx, "kubectl", podArgs...)
	if err != nil {
		return &RepoServerIssue{
			Type:        "communication",
			Message:     fmt.Sprintf("Failed to get repo-server pod status: %v", err),
			Recoverable: true,
		}
	}

	if podResult == nil || podResult.Stdout == "" {
		return &RepoServerIssue{
			Type:        "communication",
			Message:     "No repo-server pods found",
			Recoverable: false,
		}
	}

	var podList corev1.PodList
	if err := json.Unmarshal([]byte(podResult.Stdout), &podList); err != nil {
		return nil // Can't parse, assume OK
	}

	for _, pod := range podList.Items {
		// Check for restarts (indicates crash loops or OOM)
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > 0 {
				return &RepoServerIssue{
					Type:        "resource",
					Message:     fmt.Sprintf("repo-server container '%s' has restarted %d time(s) - may indicate OOM or crash", cs.Name, cs.RestartCount),
					Recoverable: true,
				}
			}

			// Check for waiting state with specific reasons
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "OOMKilled" {
					return &RepoServerIssue{
						Type:        "resource",
						Message:     fmt.Sprintf("repo-server container '%s' is in %s state", cs.Name, reason),
						Recoverable: reason != "OOMKilled", // OOM needs resource adjustment
					}
				}
			}
		}

		// Check pod phase
		if pod.Status.Phase != corev1.PodRunning {
			return &RepoServerIssue{
				Type:        "resource",
				Message:     fmt.Sprintf("repo-server pod is in %s phase (expected Running)", pod.Status.Phase),
				Recoverable: true,
			}
		}
	}

	return nil
}

// triggerRepoServerRecovery attempts to recover from repo-server issues
func (m *Manager) triggerRepoServerRecovery(ctx context.Context, appName string) bool {
	// Delete the repo-server pod to force a restart
	deleteArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "delete", "pod", "-l", "app.kubernetes.io/name=argocd-repo-server", "--wait=false")
	deleteResult, err := m.executor.Execute(ctx, "kubectl", deleteArgs...)
	if err != nil || (deleteResult != nil && deleteResult.ExitCode != 0) {
		return false
	}

	// Wait for repo-server to come back up (max 60 seconds)
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(3 * time.Second):
		}

		// Check if repo-server is running again
		issue := m.checkRepoServerHealth(ctx, false)
		if issue == nil {
			// Force a refresh of the application if specified
			if appName != "" {
				refreshArgs := m.getKubectlArgs("-n", ArgoCDNamespace, "patch", "application", appName,
					"--type", "merge", "-p", `{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"normal"}}}`)
				if _, err := m.executor.Execute(ctx, "kubectl", refreshArgs...); err != nil {
					pterm.Debug.Printf("best-effort refresh of application %s failed: %v\n", appName, err)
				}
			}
			return true
		}
	}

	return false
}

// logResourceStatus is a no-op placeholder for resource status logging
func (m *Manager) logResourceStatus(ctx context.Context, verbose bool) {
	// Resource logging disabled to reduce noise
}
