package helm

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
	"github.com/pterm/pterm"
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
// This addresses the race condition where Helm's --create-namespace may not complete before the command returns.
// Uses the native Go client (client-go); on Windows the cluster lives in WSL and must be reached from inside WSL.
func (h *HelmManager) ensureArgoCDNamespace(ctx context.Context, clusterName string, verbose bool) error {
	namespace := argocd.ArgoCDNamespace

	if err := platform.WSLClusterHint("create the argocd namespace"); err != nil {
		return err
	}
	if h.kubeClient == nil {
		return fmt.Errorf("kubernetes client unavailable: cannot reach the cluster to create the argocd namespace")
	}

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
		contextName := k8s.ResolveContextForCluster(k8s.DefaultKubeconfigPath(), clusterName)
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

	// Log the helm list output. Redact at the print site (the struct value is
	// parsed below): helm output may carry values echoed back from the release.
	if verbose {
		pterm.Info.Println("Helm list output:")
		pterm.Println(redact.Redact(result.Stdout))
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
		contextName := k8s.ResolveContextForCluster(k8s.DefaultKubeconfigPath(), clusterName)
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
		pterm.Println(redact.Redact(statusResult.Stdout))
	}

	pterm.Success.Printf("Helm release '%s' verified successfully\n", releaseName)
	return nil
}

// showArgoCDDiagnostics prints a concise, best-effort summary of ArgoCD
// deployments and pods via the native client when installation fails. The
// verbose kubectl event/log/describe dumps were dropped for a compact summary.
func (h *HelmManager) showArgoCDDiagnostics(ctx context.Context, _ string) {
	pterm.Warning.Println("=== ArgoCD Installation Diagnostics ===")
	if h.kubeClient == nil {
		pterm.Warning.Println("Native Kubernetes client unavailable; skipping diagnostics.")
		return
	}

	if deps, err := h.kubeClient.AppsV1().Deployments(argocd.ArgoCDNamespace).List(ctx, metav1.ListOptions{}); err == nil {
		pterm.Info.Println("Deployments:")
		for i := range deps.Items {
			d := deps.Items[i]
			pterm.Info.Printf("  %s: %d/%d ready\n", d.Name, d.Status.ReadyReplicas, d.Status.Replicas)
		}
	}

	pods, err := h.kubeClient.CoreV1().Pods(argocd.ArgoCDNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		pterm.Warning.Printf("Could not list ArgoCD pods: %v\n", err)
		return
	}
	pterm.Info.Println("Pods:")
	for i := range pods.Items {
		p := pods.Items[i]
		ready, total := 0, 0
		var restarts int32
		for _, cs := range p.Status.ContainerStatuses {
			total++
			if cs.Ready {
				ready++
			}
			restarts += cs.RestartCount
		}
		pterm.Info.Printf("  %s: %s, %d/%d ready, %d restart(s)\n", p.Name, p.Status.Phase, ready, total, restarts)
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				pterm.Warning.Printf("    %s waiting: %s\n", cs.Name, cs.State.Waiting.Reason)
			}
		}
	}
	pterm.Warning.Println("=== End of Diagnostics ===")
}

// verifyClusterConnectivity verifies the cluster is reachable before app-of-apps
// installation, via the native client (retried, since the API may need a moment
// after an idle period).
func (h *HelmManager) verifyClusterConnectivity(ctx context.Context, config config.ChartInstallConfig) error {
	if err := platform.WSLClusterHint("verify cluster connectivity"); err != nil {
		return err
	}
	if h.kubeClient == nil {
		return fmt.Errorf("kubernetes client unavailable: cannot reach the cluster")
	}

	pterm.Info.Println("Verifying cluster connectivity before app-of-apps installation...")
	var lastErr error
	for i := 0; i < 5; i++ {
		_, err := h.kubeClient.CoreV1().Namespaces().Get(ctx, argocd.ArgoCDNamespace, metav1.GetOptions{})
		if err == nil || k8serrors.IsNotFound(err) {
			pterm.Success.Println("Cluster is reachable")
			return nil
		}
		lastErr = err
		if config.Verbose {
			pterm.Warning.Printf("Cluster connectivity check attempt %d/5 failed: %v\n", i+1, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("cluster not reachable after retries: %w", lastErr)
}
