package argocd

import (
	"context"
	"fmt"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// repoServerSelector matches the ArgoCD repo-server pods.
const repoServerSelector = "app.kubernetes.io/name=argocd-repo-server"

// RepoServerIssue describes a detected problem with the ArgoCD repo-server.
type RepoServerIssue struct {
	Type        string // "communication", "resource", "git", "timeout"
	Message     string
	Recoverable bool
}

// printArgoCDPodDiagnostics prints a concise, best-effort summary of the ArgoCD
// workloads (via the native client) when pods fail to become ready. The verbose
// kubectl event/log dumps were dropped in favour of a compact native summary.
func (m *Manager) printArgoCDPodDiagnostics(ctx context.Context) {
	pterm.Warning.Println("ArgoCD pods failed to become ready. Collecting diagnostics...")
	if m.kubeClient == nil {
		pterm.Warning.Println("Native Kubernetes client unavailable; skipping pod diagnostics.")
		return
	}

	if deps, err := m.kubeClient.AppsV1().Deployments(ArgoCDNamespace).List(ctx, metav1.ListOptions{}); err == nil {
		pterm.Info.Println("ArgoCD deployments:")
		for i := range deps.Items {
			d := deps.Items[i]
			pterm.Info.Printf("  %s: %d/%d ready\n", d.Name, d.Status.ReadyReplicas, d.Status.Replicas)
		}
	}

	pods, err := m.kubeClient.CoreV1().Pods(ArgoCDNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		pterm.Warning.Printf("Could not list ArgoCD pods: %v\n", err)
		return
	}
	pterm.Info.Println("ArgoCD pods:")
	for i := range pods.Items {
		p := pods.Items[i]
		ready, total := containerReadiness(p)
		pterm.Info.Printf("  %s: %s, %d/%d containers ready, %d restart(s)\n",
			p.Name, p.Status.Phase, ready, total, totalRestarts(p))
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				pterm.Warning.Printf("    %s waiting: %s\n", cs.Name, cs.State.Waiting.Reason)
			}
		}
	}
}

// checkClusterConnectivity performs a lightweight cluster-reachability check via
// the native client (replaces `kubectl cluster-info`). It is load-bearing: the
// wait loop uses it to detect the cluster going away mid-install.
func (m *Manager) checkClusterConnectivity(ctx context.Context, verbose bool) error {
	if err := platform.WSLClusterHint("check cluster connectivity"); err != nil {
		return err
	}
	if err := m.initKubernetesClients(); err != nil {
		return fmt.Errorf("failed to initialize the Kubernetes client: %w", err)
	}
	if m.kubeClient == nil {
		return fmt.Errorf("kubernetes client unavailable")
	}
	// A namespace GET is a cheap call that requires a reachable API server.
	if _, err := m.kubeClient.CoreV1().Namespaces().Get(ctx, ArgoCDNamespace, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil // API reachable; namespace simply not created yet
		}
		return fmt.Errorf("cluster unreachable: %w", err)
	}
	if verbose {
		pterm.Debug.Println("Cluster connectivity check passed")
	}
	return nil
}

// printClusterDiagnostics prints a concise node summary when the cluster becomes
// unreachable. Best-effort via the native client (the previous WSL/docker/top
// shell-out dump was dropped).
func (m *Manager) printClusterDiagnostics(ctx context.Context) {
	pterm.Warning.Println("Collecting cluster diagnostics...")
	if m.kubeClient == nil {
		pterm.Warning.Println("Native Kubernetes client unavailable; skipping cluster diagnostics.")
		return
	}
	nodes, err := m.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		pterm.Warning.Printf("Could not list nodes: %v\n", err)
		return
	}
	pterm.Info.Printf("Nodes: %d\n", len(nodes.Items))
	for i := range nodes.Items {
		pterm.Info.Printf("  %s: %s\n", nodes.Items[i].Name, nodeReady(nodes.Items[i]))
	}
}

// checkRepoServerHealth inspects the ArgoCD repo-server pods via the native
// client and reports the first issue found (restarts, crash/OOM, non-Running).
func (m *Manager) checkRepoServerHealth(ctx context.Context, _ bool) *RepoServerIssue {
	if m.kubeClient == nil {
		return nil // best-effort; can't check without a client
	}
	pods, err := m.kubeClient.CoreV1().Pods(ArgoCDNamespace).List(ctx, metav1.ListOptions{LabelSelector: repoServerSelector})
	if err != nil {
		return &RepoServerIssue{Type: "communication", Message: fmt.Sprintf("Failed to get repo-server pod status: %v", err), Recoverable: true}
	}
	if len(pods.Items) == 0 {
		return &RepoServerIssue{Type: "communication", Message: "No repo-server pods found", Recoverable: false}
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > 0 {
				return &RepoServerIssue{Type: "resource", Message: fmt.Sprintf("repo-server container '%s' has restarted %d time(s) - may indicate OOM or crash", cs.Name, cs.RestartCount), Recoverable: true}
			}
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "OOMKilled" {
					return &RepoServerIssue{Type: "resource", Message: fmt.Sprintf("repo-server container '%s' is in %s state", cs.Name, reason), Recoverable: reason != "OOMKilled"}
				}
			}
		}
		if pod.Status.Phase != corev1.PodRunning {
			return &RepoServerIssue{Type: "resource", Message: fmt.Sprintf("repo-server pod is in %s phase (expected Running)", pod.Status.Phase), Recoverable: true}
		}
	}
	return nil
}

// triggerRepoServerRecovery restarts the repo-server (delete its pods → the
// controller recreates them) and optionally forces an application refresh. Uses
// the native client for both the delete and the ArgoCD Application patch.
func (m *Manager) triggerRepoServerRecovery(ctx context.Context, appName string) bool {
	if m.kubeClient == nil {
		return false
	}
	if err := m.kubeClient.CoreV1().Pods(ArgoCDNamespace).DeleteCollection(ctx,
		metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: repoServerSelector}); err != nil {
		return false
	}

	// Wait for repo-server to come back up (max ~60s).
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(3 * time.Second):
		}
		if m.checkRepoServerHealth(ctx, false) != nil {
			continue
		}
		// Recovered — force a refresh of the application if specified.
		if appName != "" && m.dynamicClient != nil {
			patch := []byte(`{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"normal"}}}`)
			if _, err := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace).
				Patch(ctx, appName, types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
				pterm.Debug.Printf("best-effort refresh of application %s failed: %v\n", appName, err)
			}
		}
		return true
	}
	return false
}

// logResourceStatus is a no-op placeholder (resource logging disabled to reduce noise).
func (m *Manager) logResourceStatus(_ context.Context, _ bool) {}

// containerReadiness returns the ready and total container counts for a pod.
func containerReadiness(p corev1.Pod) (ready, total int) {
	for _, cs := range p.Status.ContainerStatuses {
		total++
		if cs.Ready {
			ready++
		}
	}
	return ready, total
}

// totalRestarts sums container restart counts for a pod.
func totalRestarts(p corev1.Pod) int32 {
	var r int32
	for _, cs := range p.Status.ContainerStatuses {
		r += cs.RestartCount
	}
	return r
}

// nodeReady returns a human-readable readiness for a node.
func nodeReady(n corev1.Node) string {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			if c.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}
