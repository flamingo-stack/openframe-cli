package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// waitForArgoCDDeployments waits for ArgoCD workloads to be created in the cluster
// This addresses the race condition where Helm's --wait returns before Kubernetes
// has actually created the Deployment/StatefulSet objects (common in k3d/CI environments)
//
// NOTE: CRDs are installed by the Argo CD Helm chart itself (crds.install=true);
// this function only verifies the workloads exist.
//
// ArgoCD v3.x (Helm chart 10.x) deploys the application-controller as a StatefulSet,
// while server and repo-server remain as Deployments.
func (h *HelmManager) waitForArgoCDDeployments(ctx context.Context, verbose bool) error {
	if h.kubeClient == nil {
		return fmt.Errorf("kubernetes core client not initialized")
	}

	// Wait for API port to be available before making API calls
	// This prevents flooding a dead port with requests on Windows/WSL2
	if err := h.waitForAPIPort(ctx, 45*time.Second); err != nil {
		return fmt.Errorf("API port never opened: %w", err)
	}

	// List of expected Deployments (server and repo-server)
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
	}

	// List of expected StatefulSets (application-controller in ArgoCD v3.x)
	expectedStatefulSets := []string{
		"argocd-application-controller",
	}

	// CRITICAL: Use extended timeout since cluster operations can be slow
	// Native API calls are fast (~ms), so we use frequent polling with longer total timeout
	timeout := 90 * time.Second      // 90 seconds for slow CI/Windows environments
	retryInterval := 1 * time.Second // Fast polling interval (native API is ~ms per call)

	pterm.Info.Println("Waiting for ArgoCD workloads via NATIVE API...")

	// Use wait.PollUntilContextTimeout for resilient polling
	return wait.PollUntilContextTimeout(ctx, retryInterval, timeout, false, func(ctx context.Context) (bool, error) {

		missingWorkloads := []string{}

		// Check Deployments
		for _, name := range expectedDeployments {
			_, err := h.kubeClient.AppsV1().Deployments("argocd").Get(ctx, name, metav1.GetOptions{})

			if k8serrors.IsNotFound(err) {
				missingWorkloads = append(missingWorkloads, "deployment/"+name)
			} else if err != nil {
				// If it's a transient API error (not 'Not Found'), log and retry
				pterm.Warning.Printf("Transient API error checking deployment %s: %v\n", name, err)
				return false, nil
			}
		}

		// Check StatefulSets (application-controller in ArgoCD v3.x)
		for _, name := range expectedStatefulSets {
			_, err := h.kubeClient.AppsV1().StatefulSets("argocd").Get(ctx, name, metav1.GetOptions{})

			if k8serrors.IsNotFound(err) {
				missingWorkloads = append(missingWorkloads, "statefulset/"+name)
			} else if err != nil {
				// If it's a transient API error (not 'Not Found'), log and retry
				pterm.Warning.Printf("Transient API error checking statefulset %s: %v\n", name, err)
				return false, nil
			}
		}

		if len(missingWorkloads) == 0 {
			pterm.Success.Println("All ArgoCD workloads found.")
			return true, nil // Success: All workloads exist.
		}

		if verbose {
			pterm.Debug.Printf("Still missing workloads: %v\n", missingWorkloads)
		}

		return false, nil // Keep polling
	})
}

// ensureArgoCDNamespace creates the argocd namespace if it doesn't exist and waits for it to be active
// This addresses the race condition where Helm's --create-namespace may not complete before the command returns
// On Windows/WSL, uses kubectl since the native Go client can't reach the cluster running in WSL
func (h *HelmManager) ensureArgoCDNamespace(ctx context.Context, clusterName string, verbose bool) error {
	namespace := "argocd"

	// On Windows, use kubectl since the native Go client can't reach the cluster running in WSL
	if runtime.GOOS == "windows" || h.kubeClient == nil {
		return h.ensureArgoCDNamespaceKubectl(ctx, clusterName, verbose)
	}

	// Use native Go client for non-Windows platforms
	if verbose {
		pterm.Info.Println("Ensuring argocd namespace exists via native Go client...")
	}

	// Check if namespace already exists
	_, err := h.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		if verbose {
			pterm.Debug.Println("Namespace argocd already exists")
		}
		return nil
	}

	if !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to check namespace existence: %w", err)
	}

	// Create the namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = h.kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create argocd namespace: %w", err)
	}

	if verbose {
		pterm.Info.Println("Created argocd namespace, waiting for it to become Active...")
	}

	// Wait for namespace to become Active
	return wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		ns, err := h.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			return false, nil // Keep polling on transient errors
		}
		if ns.Status.Phase == corev1.NamespaceActive {
			if verbose {
				pterm.Success.Println("Namespace argocd is Active")
			}
			return true, nil
		}
		return false, nil
	})
}

// ensureArgoCDNamespaceKubectl creates the argocd namespace using kubectl (for Windows/WSL)
func (h *HelmManager) ensureArgoCDNamespaceKubectl(ctx context.Context, clusterName string, verbose bool) error {
	namespace := "argocd"

	if verbose {
		pterm.Info.Println("Ensuring argocd namespace exists via kubectl...")
	}

	// Build kubectl args with explicit context if cluster name is provided
	baseArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}

	// Check if namespace exists
	checkArgs := append(baseArgs, "get", "namespace", namespace, "-o", "name")
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    checkArgs,
	})

	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		if verbose {
			pterm.Debug.Println("Namespace argocd already exists")
		}
		return nil
	}

	// Create the namespace
	createArgs := append(baseArgs, "create", "namespace", namespace)
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    createArgs,
	})

	if err != nil {
		// Check if it's an "already exists" error (race condition)
		if result != nil && strings.Contains(result.Stderr, "already exists") {
			if verbose {
				pterm.Debug.Println("Namespace argocd already exists (created by concurrent process)")
			}
			return nil
		}
		return fmt.Errorf("failed to create argocd namespace: %w", err)
	}

	if verbose {
		pterm.Info.Println("Created argocd namespace, waiting for it to become Active...")
	}

	// Wait for namespace to become Active
	maxRetries := 60 // 60 * 500ms = 30 seconds
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Use -o json and parse in Go to avoid Windows WSL escaping issues with jsonpath.
		statusArgs := append(baseArgs, "get", "namespace", namespace, "-o", "json")
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    statusArgs,
		})

		if err == nil && result != nil && result.Stdout != "" {
			var ns corev1.Namespace
			if jerr := json.Unmarshal([]byte(result.Stdout), &ns); jerr == nil && ns.Status.Phase == corev1.NamespaceActive {
				if verbose {
					pterm.Success.Println("Namespace argocd is Active")
				}
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for argocd namespace to become Active")
}

// waitForArgoCDDeploymentsKubectl waits for ArgoCD workloads using kubectl
// This is a fallback for when the native Go client is unavailable (e.g., Windows/WSL)
//
// ArgoCD v3.x (Helm chart 8.x) deploys the application-controller as a StatefulSet,
// while server and repo-server remain as Deployments.
func (h *HelmManager) waitForArgoCDDeploymentsKubectl(ctx context.Context, clusterName string, verbose bool) error {
	// List of expected Deployments (server and repo-server)
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
	}

	// List of expected StatefulSets (application-controller in ArgoCD v3.x)
	expectedStatefulSets := []string{
		"argocd-application-controller",
	}

	// Wait settings - increased for slow CI environments
	maxRetries := 40 // 40 retries * 3 seconds = 120 seconds max (2 minutes)
	retryInterval := 3 * time.Second
	initialDelay := 5 * time.Second // Give Kubernetes time to create resources after Helm completes

	pterm.Info.Println("Waiting for ArgoCD workloads to appear via kubectl...")

	// Initial delay: Helm's --wait returns before Kubernetes controllers fully create resources
	// This delay allows the controllers to process the Helm release and create resources
	if verbose {
		pterm.Debug.Printf("Initial delay of %v to allow Kubernetes controllers to create resources...\n", initialDelay)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(initialDelay):
	}

	// Build kubectl base args with explicit context if cluster name is provided
	contextArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		contextArgs = append(contextArgs, "--context", contextName)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var missingWorkloads []string

		// Check Deployments
		deployArgs := append(contextArgs, "-n", "argocd", "get", "deployments", "-o", "json")
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    deployArgs,
		})

		if err == nil && result != nil && result.Stdout != "" {
			var deploymentList appsv1.DeploymentList
			if jsonErr := json.Unmarshal([]byte(result.Stdout), &deploymentList); jsonErr != nil {
				lastErr = fmt.Errorf("failed to parse deployments JSON: %w", jsonErr)
				if verbose {
					pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): JSON parse error\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Build set of found deployments
			foundDeployments := make(map[string]bool)
			for _, d := range deploymentList.Items {
				foundDeployments[d.Name] = true
			}

			// Check if all expected deployments are present
			for _, expected := range expectedDeployments {
				if !foundDeployments[expected] {
					missingWorkloads = append(missingWorkloads, "deployment/"+expected)
				}
			}
		} else {
			lastErr = fmt.Errorf("kubectl deployments error: %w", err)
			if verbose {
				pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): kubectl deployments error\n", i+1, maxRetries)
			}
			time.Sleep(retryInterval)
			continue
		}

		// Check StatefulSets
		stsArgs := append(contextArgs, "-n", "argocd", "get", "statefulsets", "-o", "json")
		result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    stsArgs,
		})

		if err == nil && result != nil && result.Stdout != "" {
			var statefulSetList appsv1.StatefulSetList
			if jsonErr := json.Unmarshal([]byte(result.Stdout), &statefulSetList); jsonErr != nil {
				lastErr = fmt.Errorf("failed to parse statefulsets JSON: %w", jsonErr)
				if verbose {
					pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): StatefulSet JSON parse error\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Build set of found statefulsets
			foundStatefulSets := make(map[string]bool)
			for _, s := range statefulSetList.Items {
				foundStatefulSets[s.Name] = true
			}

			// Check if all expected statefulsets are present
			for _, expected := range expectedStatefulSets {
				if !foundStatefulSets[expected] {
					missingWorkloads = append(missingWorkloads, "statefulset/"+expected)
				}
			}
		} else {
			lastErr = fmt.Errorf("kubectl statefulsets error: %w", err)
			if verbose {
				pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): kubectl statefulsets error\n", i+1, maxRetries)
			}
			time.Sleep(retryInterval)
			continue
		}

		// Check if all workloads are present
		if len(missingWorkloads) == 0 {
			pterm.Success.Println("All ArgoCD workloads found.")
			return nil
		}

		lastErr = fmt.Errorf("missing workloads: %v", missingWorkloads)
		if verbose {
			pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): missing %v\n", i+1, maxRetries, missingWorkloads)
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("workloads not found after %d retries: %w", maxRetries, lastErr)
}

// waitForAPIPort waits for the Kubernetes API port to be open before making API calls
// This prevents flooding a dead port with requests on Windows/WSL2 where the port
// might not be immediately available after k3d reports success
func (h *HelmManager) waitForAPIPort(ctx context.Context, timeout time.Duration) error {
	if h.kubeConfig == nil {
		return nil // Skip if no kubeConfig available
	}

	// Extract host:port from kubeConfig.Host
	apiAddress := strings.TrimPrefix(strings.TrimPrefix(h.kubeConfig.Host, "https://"), "http://")
	if apiAddress == "" {
		return nil // Skip if we can't determine the address
	}

	dialer := net.Dialer{Timeout: 2 * time.Second}
	pterm.Info.Printf("Waiting for API port %s to open...\n", apiAddress)

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		conn, err := dialer.DialContext(ctx, "tcp", apiAddress)
		if err == nil {
			_ = conn.Close()
			pterm.Success.Printf("API port %s is open\n", apiAddress)
			return true, nil // Port is open!
		}
		return false, nil // Keep polling
	})
}

// verifyHelmRelease checks if a Helm release was actually created by running helm list
// This helps diagnose issues where Helm reports success but doesn't create resources
func (h *HelmManager) verifyHelmRelease(ctx context.Context, releaseName, namespace, clusterName string, verbose bool) error {
	if verbose {
		pterm.Info.Printf("Verifying Helm release '%s' in namespace '%s'...\n", releaseName, namespace)
	}

	// Build helm list args
	args := []string{"list", "-n", namespace, "--filter", releaseName, "-o", "json"}

	// Add explicit kube-context if cluster name is provided
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		args = append(args, "--kube-context", contextName)
	}

	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("failed to run helm list: %w", err)
	}

	// Log the helm list output
	if verbose {
		pterm.Info.Println("Helm list output:")
		pterm.Println(result.Stdout)
	}

	// Check if the release exists in the output
	// The JSON output will be an empty array "[]" if no releases found
	output := strings.TrimSpace(result.Stdout)
	if output == "" || output == "[]" {
		return fmt.Errorf("helm release '%s' not found in namespace '%s' - helm list returned empty", releaseName, namespace)
	}

	// Also run helm status for more details
	statusArgs := []string{"status", releaseName, "-n", namespace}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		statusArgs = append(statusArgs, "--kube-context", contextName)
	}

	statusResult, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    statusArgs,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("helm release exists but status check failed: %w", err)
	}

	if verbose {
		pterm.Info.Println("Helm status output:")
		pterm.Println(statusResult.Stdout)
	}

	pterm.Success.Printf("Helm release '%s' verified successfully\n", releaseName)
	return nil
}

// showArgoCDDiagnostics outputs diagnostic information about ArgoCD pods when installation fails
// This helps identify why the Helm install timed out (e.g., image pull issues, crashloops, pending pods)
func (h *HelmManager) showArgoCDDiagnostics(ctx context.Context, clusterName string) {
	pterm.Warning.Println("=== ArgoCD Installation Diagnostics ===")

	// Build kubectl args with explicit context if cluster name is provided
	baseArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}

	// Get pod status
	pterm.Info.Println("Pod Status in argocd namespace:")
	podArgs := append(baseArgs, "get", "pods", "-n", "argocd", "-o", "wide")
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    podArgs,
	})
	if err == nil && result != nil {
		pterm.Println(result.Stdout)
		if result.Stderr != "" {
			pterm.Println(result.Stderr)
		}
	} else {
		pterm.Error.Printf("Failed to get pods: %v\n", err)
	}

	// Get events in the namespace (sorted by timestamp)
	pterm.Info.Println("\nRecent Events in argocd namespace:")
	eventArgs := append(baseArgs, "get", "events", "-n", "argocd", "--sort-by=.lastTimestamp")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    eventArgs,
	})
	if err == nil && result != nil {
		pterm.Println(result.Stdout)
		if result.Stderr != "" {
			pterm.Println(result.Stderr)
		}
	} else {
		pterm.Error.Printf("Failed to get events: %v\n", err)
	}

	// Get logs from all ArgoCD pods (helpful for CrashLoopBackOff debugging)
	// Use -o json to avoid Windows WSL escaping issues with jsonpath
	pterm.Info.Println("\nPod Logs (last 50 lines from each container):")
	podListArgs := append(baseArgs, "get", "pods", "-n", "argocd", "-o", "json")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    podListArgs,
	})
	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		var podList corev1.PodList
		pods := []string{}
		if jsonErr := json.Unmarshal([]byte(result.Stdout), &podList); jsonErr == nil {
			for _, p := range podList.Items {
				pods = append(pods, p.Name)
			}
		}
		for _, pod := range pods {
			if pod == "" {
				continue
			}
			// Get current logs
			pterm.Info.Printf("--- Logs from pod: %s ---\n", pod)
			logArgs := append(baseArgs, "logs", pod, "-n", "argocd", "--all-containers=true", "--tail=50")
			logResult, logErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    logArgs,
			})
			if logErr == nil && logResult != nil && strings.TrimSpace(logResult.Stdout) != "" {
				pterm.Println(logResult.Stdout)
			} else if logErr != nil {
				pterm.Debug.Printf("No current logs available for %s\n", pod)
			}

			// Get previous logs (from crashed containers)
			prevLogArgs := append(baseArgs, "logs", pod, "-n", "argocd", "--all-containers=true", "--tail=50", "--previous")
			prevLogResult, prevLogErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    prevLogArgs,
			})
			if prevLogErr == nil && prevLogResult != nil && strings.TrimSpace(prevLogResult.Stdout) != "" {
				pterm.Info.Printf("--- Previous logs from pod: %s (crashed container) ---\n", pod)
				pterm.Println(prevLogResult.Stdout)
			}
		}
	}

	// Describe pods that are not Running
	pterm.Info.Println("\nDescribing non-running pods:")
	describeArgs := append(baseArgs, "get", "pods", "-n", "argocd", "--field-selector=status.phase!=Running", "-o", "name")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    describeArgs,
	})
	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		pods := strings.Split(strings.TrimSpace(result.Stdout), "\n")
		for _, pod := range pods {
			if pod == "" {
				continue
			}
			pterm.Info.Printf("--- Describing pod: %s ---\n", pod)
			describePodArgs := append(baseArgs, "describe", pod, "-n", "argocd")
			descResult, descErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    describePodArgs,
			})
			if descErr == nil && descResult != nil {
				pterm.Println(descResult.Stdout)
			}
		}
	}

	pterm.Warning.Println("=== End of Diagnostics ===")
}

// verifyClusterConnectivity verifies that the cluster is reachable before running helm commands
// This is important after idle periods where WSL networking may have gone stale
// It also logs kubeconfig details to help diagnose connectivity issues
func (h *HelmManager) verifyClusterConnectivity(ctx context.Context, config config.ChartInstallConfig) error {
	pterm.Info.Println("Verifying cluster connectivity before app-of-apps installation...")

	// Build kubectl args with explicit context if cluster name is provided
	kubectlArgs := []string{"cluster-info"}
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		kubectlArgs = []string{"--context", contextName, "cluster-info"}
	}

	// On Windows/WSL, also dump kubeconfig details for debugging
	if runtime.GOOS == "windows" {
		h.debugWSLKubeconfig(ctx, config.Verbose)
	}

	// Retry kubectl cluster-info a few times (cluster may need a moment after idle)
	maxRetries := 5
	retryDelay := 2 * time.Second
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    kubectlArgs,
		})

		if err == nil && result.ExitCode == 0 {
			pterm.Success.Println("Cluster is reachable")
			if config.Verbose {
				pterm.Info.Println("kubectl cluster-info output:")
				pterm.Println(result.Stdout)
			}
			return nil
		}

		lastErr = err
		if config.Verbose {
			pterm.Warning.Printf("Cluster connectivity check attempt %d/%d failed: %v\n", i+1, maxRetries, err)
			if result != nil {
				if result.Stdout != "" {
					pterm.Println("stdout:", result.Stdout)
				}
				if result.Stderr != "" {
					pterm.Println("stderr:", result.Stderr)
				}
			}
		}

		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("cluster not reachable after %d attempts: %w", maxRetries, lastErr)
}

// debugWSLKubeconfig logs kubeconfig details from WSL for debugging connectivity issues
func (h *HelmManager) debugWSLKubeconfig(ctx context.Context, verbose bool) {
	if !verbose {
		return
	}

	pterm.Info.Println("=== WSL Kubeconfig Debug Info ===")

	// Get the WSL user (same logic as executor)
	wslUser := os.Getenv("WSL_USER")
	if wslUser == "" {
		wslUser = "runner"
	}

	// Check if kubeconfig exists
	checkCmd := "ls -la ~/.kube/config 2>&1 || echo 'Kubeconfig not found'"
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", checkCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Kubeconfig file status:")
		pterm.Println(result.Stdout)
	}

	// Show the server addresses in kubeconfig (without showing secrets)
	serverCmd := "grep -A2 'cluster:' ~/.kube/config 2>/dev/null | grep 'server:' || echo 'No server entries found'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", serverCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Server addresses in kubeconfig:")
		pterm.Println(result.Stdout)
	}

	// Show current context
	contextCmd := "kubectl config current-context 2>&1 || echo 'No current context'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", contextCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Current kubectl context:")
		pterm.Println(result.Stdout)
	}

	// Check if the API server port is reachable
	portCheckCmd := "nc -zv 127.0.0.1 6550 2>&1 || echo 'Port 6550 not reachable'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", portCheckCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Port 6550 connectivity check:")
		pterm.Println(result.Stdout)
	}

	// Show Docker containers (k3d nodes)
	dockerCmd := "docker ps --filter 'name=k3d' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>&1 || echo 'Docker not available or no k3d containers'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", dockerCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("k3d Docker containers:")
		pterm.Println(result.Stdout)
	}

	pterm.Info.Println("=== End WSL Kubeconfig Debug ===")
}
