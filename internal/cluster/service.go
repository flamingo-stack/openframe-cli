package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/provider"
	uiCluster "github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ApplicationCleaner removes the ArgoCD Application CRs that own the platform
// workloads, and strips the resources-finalizer from any left in Terminating.
//
// Cleanup needs both, in that order around the Helm uninstall: Applications
// must be deleted while the ArgoCD controller still runs (so it cascades the
// workload cleanup), and the finalizers must be stripped afterwards, once the
// controller — the only thing that could clear them — is gone. Otherwise the
// CRs sit in Terminating forever and pin the argocd namespace.
//
// It is an interface because internal/cluster must not import internal/chart:
// the ArgoCD-backed implementation is injected by the command layer, exactly
// like ClusterAccess in the app subsystem.
type ApplicationCleaner interface {
	DeleteApplications(ctx context.Context) (int, error)
	RemoveApplicationFinalizers(ctx context.Context) (int, error)
}

// ClusterService provides cluster configuration and management operations
// This handles cluster lifecycle operations and configuration management
type ClusterService struct {
	manager    provider.Provider
	executor   executor.CommandExecutor
	suppressUI bool // Suppress interactive UI elements for automation
	// appCleaner, when set, lets cleanup remove ArgoCD Applications before the
	// Helm uninstall and strip their finalizers afterwards. Optional: nil means
	// the Helm/namespace phases run as before (the CRs may then stay stuck).
	appCleaner ApplicationCleaner
}

// WithApplicationCleaner injects the ArgoCD-backed application cleaner used by
// the cleanup flow. Returns the service for chaining.
func (s *ClusterService) WithApplicationCleaner(c ApplicationCleaner) *ClusterService {
	s.appCleaner = c
	return s
}

// isTerminalEnvironment checks if we're running in a proper terminal
func isTerminalEnvironment() bool {
	// Check if stdout is a terminal
	if stat, err := os.Stdout.Stat(); err == nil {
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// NewClusterService creates a new cluster service with default configuration
func NewClusterService(exec executor.CommandExecutor) *ClusterService {
	manager, _ := provider.New(models.ClusterTypeK3d, exec) // k3d never fails to construct
	return &ClusterService{
		manager:    manager,
		executor:   exec,
		suppressUI: false,
	}
}

// NewClusterServiceSuppressed creates a cluster service with UI suppression
func NewClusterServiceSuppressed(exec executor.CommandExecutor) *ClusterService {
	manager, _ := provider.New(models.ClusterTypeK3d, exec) // k3d never fails to construct
	return &ClusterService{
		manager:    manager,
		executor:   exec,
		suppressUI: true,
	}
}

// providerFor resolves the backend for a cluster type. The k3d manager is the
// service default (list/status/cleanup are k3d-scoped until cloud backends
// land); anything else goes through the factory, which today yields
// ErrProviderNotFound for the recognized cloud types.
func (s *ClusterService) providerFor(clusterType models.ClusterType) (provider.Provider, error) {
	if clusterType == models.ClusterTypeK3d || clusterType == "" {
		return s.manager, nil
	}
	return provider.New(clusterType, s.executor)
}

// CreateCluster handles cluster creation operations
// Returns the *rest.Config for the created cluster that can be used to interact with it
func (s *ClusterService) CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error) {
	// Resolve the backend first: an unsupported type must fail here, before any
	// k3d-specific existence checks run.
	mgr, err := s.providerFor(config.Type)
	if err != nil {
		return nil, err
	}

	// Check if cluster already exists
	if existingInfo, err := mgr.GetClusterStatus(ctx, config.Name); err == nil {
		// Cluster already exists - show friendly message

		// Show warning for existing cluster
		pterm.Warning.Printf("Cluster '%s' already exists!\n", pterm.Cyan(config.Name))
		pterm.DefaultBasicText.Println()

		boxContent := fmt.Sprintf(
			"NAME:     %s\n"+
				"TYPE:     %s\n"+
				"STATUS:   %s\n"+
				"NODES:    %d\n"+
				"NETWORK:  k3d-%s",
			pterm.Bold.Sprint(existingInfo.Name),
			strings.ToUpper(string(existingInfo.Type)),
			pterm.Green("Running"),
			existingInfo.NodeCount,
			existingInfo.Name,
		)

		pterm.DefaultBox.
			WithTitle(" ⚠️  Cluster Already Running  ⚠️ ").
			WithTitleTopCenter().
			Println(boxContent)

		// Show what user can do (suppress for automation)
		if !s.suppressUI {
			pterm.DefaultBasicText.Println()
			pterm.Info.Printf("What would you like to do?\n")
			pterm.DefaultBasicText.Printf("  • Check status: openframe cluster status %s\n", config.Name)
			pterm.DefaultBasicText.Printf("  • Delete first: openframe cluster delete %s\n", config.Name)
			pterm.DefaultBasicText.Printf("  • Use different name: openframe cluster create my-new-cluster\n")
		}

		// Return the rest.Config for the existing cluster
		restConfig, err := mgr.GetRestConfig(ctx, config.Name)
		if err != nil {
			return nil, fmt.Errorf("cluster exists but failed to get REST config: %w", err)
		}
		return restConfig, nil // Exit gracefully without error
	}

	// Cluster doesn't exist, proceed with creation
	var sp *spinner.Spinner
	if !s.suppressUI {
		sp = spinner.New()
		sp.Start(fmt.Sprintf("Creating %s cluster '%s'...", config.Type, config.Name))
	} else {
		// In non-interactive mode, just show a simple info message
		pterm.Info.Printf("Creating %s cluster '%s'...\n", config.Type, config.Name)
	}

	restConfig, err := mgr.CreateCluster(ctx, config)
	if err != nil {
		if sp != nil {
			sp.Fail(fmt.Sprintf("Failed to create cluster '%s'", config.Name))
		}
		return nil, err
	}

	if sp != nil {
		sp.Success(fmt.Sprintf("Cluster '%s' created successfully", config.Name))
	} else {
		pterm.Success.Printf("Cluster '%s' created successfully\n", config.Name)
	}

	// Get and display cluster status
	if clusterInfo, statusErr := mgr.GetClusterStatus(ctx, config.Name); statusErr == nil {
		s.displayClusterCreationSummary(clusterInfo)
	}

	// Show next steps
	s.showNextSteps(config.Name)

	return restConfig, nil
}

// DeleteCluster handles cluster deletion business logic
func (s *ClusterService) DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error {
	mgr, err := s.providerFor(clusterType)
	if err != nil {
		return err
	}

	// Show deletion progress
	var sp *spinner.Spinner
	if !s.suppressUI {
		sp = spinner.New()
		sp.Start(fmt.Sprintf("Deleting %s cluster '%s'...", clusterType, name))
	} else {
		pterm.Info.Printf("Deleting %s cluster '%s'...\n", clusterType, name)
	}

	err = mgr.DeleteCluster(ctx, name, clusterType, force)
	if err != nil {
		if sp != nil {
			sp.Fail(fmt.Sprintf("Failed to delete cluster '%s'", name))
		}
		return err
	}

	if sp != nil {
		sp.Stop() // Stop spinner without message - UI layer will show success
	}

	// Don't show summary here - let the UI layer handle it

	return nil
}

// cloudProviders returns the cloud backends (EKS, GKE). Their registries are
// plain files under ~/.openframe, so construction practically never fails; a
// failed one is skipped, degrading the CLI to reduced visibility rather than
// blocking local operations.
func (s *ClusterService) cloudProviders() []provider.Provider {
	var providers []provider.Provider
	for _, t := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
		if p, err := s.providerFor(t); err == nil {
			providers = append(providers, p)
		}
	}
	return providers
}

// ListClusters merges the local k3d clusters with the cloud clusters recorded
// in the workspace registry.
func (s *ClusterService) ListClusters() ([]models.ClusterInfo, error) {
	ctx := context.Background()
	clusters, err := s.manager.ListAllClusters(ctx)
	if err != nil {
		return nil, err
	}
	for _, cloud := range s.cloudProviders() {
		cloudClusters, err := cloud.ListAllClusters(ctx)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cloudClusters...)
	}
	return clusters, nil
}

// GetClusterStatus handles cluster status business logic
func (s *ClusterService) GetClusterStatus(name string) (models.ClusterInfo, error) {
	ctx := context.Background()
	for _, cloud := range s.cloudProviders() {
		if info, err := cloud.GetClusterStatus(ctx, name); err == nil {
			return info, nil
		}
	}
	return s.manager.GetClusterStatus(ctx, name)
}

// GetRestConfig returns the rest.Config for an existing cluster
func (s *ClusterService) GetRestConfig(name string) (*rest.Config, error) {
	ctx := context.Background()
	for _, cloud := range s.cloudProviders() {
		if _, err := cloud.DetectClusterType(ctx, name); err == nil {
			return cloud.GetRestConfig(ctx, name)
		}
	}
	return s.manager.GetRestConfig(ctx, name)
}

// DetectClusterType consults the cloud registry first (a cheap local file
// read), then falls back to k3d discovery.
func (s *ClusterService) DetectClusterType(name string) (models.ClusterType, error) {
	ctx := context.Background()
	for _, cloud := range s.cloudProviders() {
		if t, err := cloud.DetectClusterType(ctx, name); err == nil {
			return t, nil
		}
	}
	return s.manager.DetectClusterType(ctx, name)
}

// CleanupCluster handles cluster cleanup business logic. The returned
// CleanupResult reports what was actually removed and which phases failed; a
// nil error with a non-empty Failures list is a partial cleanup.
func (s *ClusterService) CleanupCluster(ctx context.Context, name string, clusterType models.ClusterType, verbose bool, force bool) (models.CleanupResult, error) {
	switch clusterType {
	case models.ClusterTypeK3d:
		return s.cleanupK3dCluster(ctx, name, verbose, force)
	case models.ClusterTypeEKS, models.ClusterTypeGKE:
		return models.CleanupResult{}, fmt.Errorf("cleanup is not supported for cloud clusters; use 'openframe cluster delete %s' to tear the cluster down", name)
	default:
		return models.CleanupResult{}, fmt.Errorf("cleanup not supported for cluster type: %s", clusterType)
	}
}

// cleanupK3dCluster handles K3d-specific cleanup.
//
// Every phase is best-effort: a failure is recorded and the next phase still
// runs, because a partly-installed cluster must remain tearable-down. Failures
// are surfaced (not just under --verbose) so "cleanup completed" never hides a
// phase that did nothing.
func (s *ClusterService) cleanupK3dCluster(ctx context.Context, clusterName string, verbose bool, force bool) (models.CleanupResult, error) {
	if verbose {
		pterm.Info.Printf("Starting cleanup of cluster: %s\n", clusterName)
	}
	var result models.CleanupResult

	// 1. Delete the ArgoCD Applications WHILE the ArgoCD controller is still
	// running, so it cascades the workload cleanup itself. Best-effort: a
	// cluster without OpenFrame installed simply has none.
	if s.appCleaner != nil {
		deleted, err := s.appCleaner.DeleteApplications(ctx)
		switch {
		case err != nil:
			result.AddFailure("ArgoCD applications", err)
		default:
			result.ApplicationsDeleted = deleted
			if deleted > 0 && verbose {
				pterm.Info.Printf("Deleted %d ArgoCD application(s)\n", deleted)
			}
		}
	}

	// 2. Clean up Helm releases (including ArgoCD) — pinned to this cluster's
	// kube-context. Without the pin helm operates on the kubeconfig's CURRENT
	// context, which may be a different (even production) cluster.
	kubeContext := k8s.ResolveContextForCluster(k8s.DefaultKubeconfigPath(), clusterName)
	removed, err := s.cleanupHelmReleases(ctx, kubeContext, verbose, force)
	result.ReleasesRemoved = removed
	if err != nil {
		result.AddFailure("Helm releases", err)
	}

	// 3. ArgoCD is gone now, so nothing is left to clear its resources-finalizer.
	// Strip it from any Application still in Terminating — otherwise those CRs
	// (and the argocd namespace deleted in the next phase) never get reaped.
	if s.appCleaner != nil {
		cleared, err := s.appCleaner.RemoveApplicationFinalizers(ctx)
		switch {
		case err != nil:
			result.AddFailure("application finalizers", err)
		default:
			result.FinalizersCleared = cleared
			if cleared > 0 && verbose {
				pterm.Info.Printf("Cleared finalizers on %d stuck application(s)\n", cleared)
			}
		}
	}

	// 4. Clean up Kubernetes resources in common namespaces
	deletedNS, err := s.cleanupKubernetesResources(ctx, clusterName, verbose, force)
	result.NamespacesDeleted = deletedNS
	if err != nil {
		result.AddFailure("Kubernetes namespaces", err)
	}

	// 5. Reclaim disk by pruning unused container images inside each node.
	// Not gated on force: removing images no container references is safe, and
	// reclaiming disk is the whole point of `cluster cleanup`.
	pruned, err := s.cleanupNodeImages(ctx, clusterName, verbose)
	result.NodesPruned = pruned
	if err != nil {
		result.AddFailure("Container images", err)
	}

	if verbose {
		pterm.Success.Printf("Cleanup completed for cluster: %s\n", clusterName)
	}

	return result, nil
}

// helmRelease is the subset of `helm list --output json` we consume.
type helmRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// cleanupHelmReleases removes all Helm releases from the cluster identified by
// kubeContext. The explicit --kube-context on every helm call is what keeps
// cleanup scoped to that cluster (T0-1): helm otherwise acts on the
// kubeconfig's current context, whatever the user last switched to.
// It returns the number of releases actually uninstalled. A release that fails
// to uninstall is counted as a failure, not as removed.
func (s *ClusterService) cleanupHelmReleases(ctx context.Context, kubeContext string, verbose bool, force bool) (int, error) {
	if kubeContext == "" {
		return 0, fmt.Errorf("refusing to cleanup Helm releases without an explicit kube-context")
	}

	result, err := s.executor.Execute(ctx, "helm", "list", "--all-namespaces", "--output", "json", "--kube-context", kubeContext)
	if err != nil {
		return 0, fmt.Errorf("failed to list Helm releases: %w", err)
	}

	var releases []helmRelease
	if out := strings.TrimSpace(result.Stdout); out != "" {
		if err := json.Unmarshal([]byte(out), &releases); err != nil {
			return 0, fmt.Errorf("failed to parse helm list output: %w", err)
		}
	}
	if len(releases) == 0 {
		if verbose {
			pterm.Info.Println("No Helm releases found to cleanup")
		}
		return 0, nil
	}

	var removed int
	var failed []string
	for _, release := range releases {
		if release.Name == "" || release.Namespace == "" {
			continue
		}

		if verbose {
			pterm.Info.Printf("Uninstalling Helm release: %s (namespace %s)\n", release.Name, release.Namespace)
		}

		// Aggressive uninstall, deliberately WITHOUT --wait: the releases here
		// include argo-cd and app-of-apps, whose Application CRs carry ArgoCD's
		// resources-finalizer. Once the ArgoCD controller is being removed it
		// can no longer clear that finalizer, so --wait would block for helm's
		// default 5m PER RELEASE (see UninstallRelease in
		// internal/chart/providers/helm for the same rationale).
		//
		// The Application CRs left in Terminating are reaped by the
		// finalizer-stripping phase that runs right after this one (see
		// cleanupK3dCluster step 3), mirroring `app uninstall`.
		args := []string{"uninstall", release.Name, "--namespace", release.Namespace, "--kube-context", kubeContext, "--no-hooks"}
		if force {
			// Add even more aggressive flags when force is enabled
			args = append(args, "--ignore-not-found")
		}
		if _, err := s.executor.Execute(ctx, "helm", args...); err != nil {
			failed = append(failed, release.Name)
			if verbose {
				pterm.Warning.Printf("Failed to uninstall release %s: %v\n", release.Name, err)
			}
		} else {
			removed++
			if verbose {
				pterm.Success.Printf("Uninstalled Helm release: %s\n", release.Name)
			}
		}
	}

	if len(failed) > 0 {
		return removed, fmt.Errorf("%d of %d release(s) could not be uninstalled: %s",
			len(failed), len(releases), strings.Join(failed, ", "))
	}
	return removed, nil
}

// protectedNamespaces must never be deleted by cleanup, regardless of --force.
// Deleting any of these can render the cluster unrecoverable or destroy
// unrelated workloads (audit I7/M3).
var protectedNamespaces = map[string]struct{}{
	"kube-system":     {},
	"kube-public":     {},
	"kube-node-lease": {},
	"default":         {},
}

// isProtectedNamespace reports whether ns must never be deleted.
func isProtectedNamespace(ns string) bool {
	_, ok := protectedNamespaces[ns]
	return ok
}

// filterProtectedNamespaces returns raw with every protected/system namespace
// removed. It is the I7 defense-in-depth guard: even if a protected namespace is
// added to a cleanup list by mistake, it can never be deleted.
func filterProtectedNamespaces(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, ns := range raw {
		if !isProtectedNamespace(ns) {
			out = append(out, ns)
		}
	}
	return out
}

// cleanupKubernetesResources removes namespaces created by OpenFrame components
// via the native Kubernetes client (client-go). It never touches
// protected/system namespaces.
// It returns the number of namespaces whose deletion was accepted by the API
// server.
func (s *ClusterService) cleanupKubernetesResources(ctx context.Context, clusterName string, verbose bool, _ bool) (int, error) {
	// On Windows the cluster lives in WSL and must be reached from inside WSL.
	if err := platform.WSLClusterHint("clean up OpenFrame namespaces"); err != nil {
		return 0, err
	}

	// TLS policy is the provider's mint-time decision: k3d marks its local
	// rest.Config insecure itself (verify.go), and a future cloud provider's
	// config must NOT be downgraded here.
	restConfig, err := s.manager.GetRestConfig(ctx, clusterName)
	if err != nil {
		return 0, fmt.Errorf("failed to get cluster config for cleanup: %w", err)
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Namespaces created by OpenFrame component installs. System namespaces are
	// intentionally absent and are additionally filtered (I7 defense-in-depth).
	var deleted int
	var failed []string
	for _, namespace := range filterProtectedNamespaces([]string{"argocd", "openframe"}) {
		if _, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
			continue // doesn't exist (or unreachable) — skip
		}

		if verbose {
			pterm.Info.Printf("Cleaning up namespace: %s\n", namespace)
		}

		if err := client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{}); err != nil && !k8serrors.IsNotFound(err) {
			failed = append(failed, namespace)
			if verbose {
				pterm.Warning.Printf("Failed to delete namespace %s: %v\n", namespace, err)
			}
		} else {
			deleted++
			if verbose {
				pterm.Success.Printf("Deleted namespace: %s\n", namespace)
			}
		}
	}

	if len(failed) > 0 {
		return deleted, fmt.Errorf("could not delete namespace(s): %s", strings.Join(failed, ", "))
	}
	return deleted, nil
}

// cleanupNodeImages reclaims disk by removing unused container images inside
// each k3d node. It returns the number of nodes pruned without error.
//
// The nodes run CONTAINERD, not Docker. They are `rancher/k3s` containers, and
// that image ships no docker binary at all — so the previous implementation,
// which ran `docker exec <node> docker image|container|volume|network|system
// prune`, failed with exit 127 ("docker": executable file not found in $PATH)
// on every node, for every prune, on every cleanup. Every error was swallowed,
// so `cluster cleanup` reported success and reclaimed nothing, ever.
//
// crictl is the CRI client k3s provides: /bin/crictl is a symlink to the k3s
// multi-call binary, so it is always present in a k3d node (verified against
// rancher/k3s:v1.35.5-k3s1). `crictl rmi --prune` removes every image not
// referenced by a container.
//
// There is deliberately no container/volume/network/builder prune: those are
// Docker concepts with no containerd equivalent inside the node. Stopped pods
// are the kubelet's business, and the node's volumes and networks belong to the
// Docker daemon on the host, not to anything running inside the node.
func (s *ClusterService) cleanupNodeImages(ctx context.Context, clusterName string, verbose bool) (int, error) {
	if verbose {
		pterm.Info.Printf("Reclaiming unused container images for cluster: %s\n", clusterName)
	}

	// Dynamically discover all k3d nodes for this cluster
	nodeNames, err := s.getK3dClusterNodes(ctx, clusterName)
	if err != nil {
		// Not fatal: a cluster whose nodes are already gone still cleans up.
		return 0, fmt.Errorf("could not discover cluster nodes: %w", err)
	}

	if len(nodeNames) == 0 {
		if verbose {
			pterm.Info.Printf("No k3d nodes found for cluster: %s\n", clusterName)
		}
		return 0, nil
	}

	if verbose {
		pterm.Info.Printf("Found %d k3d nodes for cluster %s\n", len(nodeNames), clusterName)
	}

	var pruned int
	var failed []string
	for _, nodeName := range nodeNames {
		if verbose {
			pterm.Info.Printf("Pruning unused images in node: %s\n", nodeName)
		}

		if _, err := s.executor.Execute(ctx, "docker", "exec", nodeName, "crictl", "rmi", "--prune"); err != nil {
			failed = append(failed, nodeName)
			if verbose {
				pterm.Warning.Printf("Failed to prune images in node %s: %v\n", nodeName, err)
			}
			continue
		}
		pruned++
	}

	if verbose {
		pterm.Success.Printf("Image cleanup completed for cluster: %s\n", clusterName)
	}

	if len(failed) > 0 {
		return pruned, fmt.Errorf("image prune failed on node(s): %s", strings.Join(failed, ", "))
	}
	return pruned, nil
}

// getK3dClusterNodes discovers all Docker containers that are part of a k3d cluster
// It returns only server and agent nodes (excludes load balancer and tools containers)
func (s *ClusterService) getK3dClusterNodes(ctx context.Context, clusterName string) ([]string, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("cluster name cannot be empty")
	}

	// Use docker ps to find all containers with the k3d cluster label
	// Only include running containers for cleanup operations
	result, err := s.executor.Execute(ctx, "docker", "ps",
		"--filter", fmt.Sprintf("label=k3d.cluster=%s", clusterName),
		"--filter", "status=running",
		"--format", "{{.Names}}")
	if err != nil {
		return nil, fmt.Errorf("failed to list k3d cluster nodes for cluster %s: %w", clusterName, err)
	}

	if result.Stdout == "" {
		return []string{}, nil
	}

	return s.filterK3dNodes(result.Stdout, clusterName), nil
}

// filterK3dNodes filters and validates k3d node names, excluding non-node containers
func (s *ClusterService) filterK3dNodes(output, clusterName string) []string {
	// Always return an empty slice instead of nil for consistent behavior
	validNodes := make([]string, 0)

	if strings.TrimSpace(output) == "" {
		return validNodes
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		nodeName := strings.TrimSpace(line)
		if nodeName == "" {
			continue
		}

		// Only include server and agent nodes, exclude load balancer and tools containers
		// k3d nodes follow the pattern: k3d-{cluster}-{server|agent}-{number}
		if s.isK3dWorkerNode(nodeName, clusterName) {
			validNodes = append(validNodes, nodeName)
		}
	}

	return validNodes
}

// isK3dWorkerNode checks if a container name represents a k3d worker node (server or agent)
func (s *ClusterService) isK3dWorkerNode(nodeName, clusterName string) bool {
	prefix := fmt.Sprintf("k3d-%s-", clusterName)

	// Must start with the correct cluster prefix
	if !strings.HasPrefix(nodeName, prefix) {
		return false
	}

	suffix := strings.TrimPrefix(nodeName, prefix)

	// Check if it's a server or agent node (exclude serverlb, tools, etc.)
	return strings.HasPrefix(suffix, "server-") || strings.HasPrefix(suffix, "agent-")
}

// apiServerEndpoint returns the cluster's API server URL as the kubeconfig
// records it, or "" when it cannot be determined.
//
// It replaces a hardcoded "https://0.0.0.0:6550" printed in three places. 6550
// is only the *preferred* API port: findPort falls back to 6551, then 6552,
// when the port is taken (see providers/k3d/ports.go). So on any machine with
// a second cluster the box pointed the user at a different cluster's API
// server. The kubeconfig records the port that was actually bound.
func (s *ClusterService) apiServerEndpoint(ctx context.Context, name string) string {
	cfg, err := s.manager.GetRestConfig(ctx, name)
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.Host
}

// apiServerLine renders the API row for the summary boxes, degrading to an
// honest "unknown" rather than inventing an address.
func apiServerLine(endpoint string) string {
	if endpoint == "" {
		return "API:      (unknown — kubeconfig not readable)"
	}
	return "API:      " + endpoint
}

// displayClusterCreationSummary displays a summary after cluster creation
func (s *ClusterService) displayClusterCreationSummary(info models.ClusterInfo) {
	pterm.DefaultBasicText.Println()

	// Create a clean box for the summary
	boxContent := fmt.Sprintf(
		"NAME:     %s\n"+
			"TYPE:     %s\n"+
			"STATUS:   %s\n"+
			"NODES:    %d\n"+
			"NETWORK:  k3d-%s\n"+
			"%s",
		pterm.Bold.Sprint(info.Name),
		strings.ToUpper(string(info.Type)),
		pterm.Green("Ready"),
		info.NodeCount,
		info.Name,
		apiServerLine(s.apiServerEndpoint(context.Background(), info.Name)),
	)

	pterm.DefaultBox.
		WithTitle(" ✅ Cluster Created ").
		WithTitleTopCenter().
		Println(boxContent)
}

// showNextSteps displays clean next steps after cluster creation
func (s *ClusterService) showNextSteps(clusterName string) {
	// Skip showing next steps if UI is suppressed (e.g., during bootstrap)
	if s.suppressUI {
		return
	}

	pterm.DefaultBasicText.Println()
	pterm.Info.Printf("🚀 Next Steps:\n")
	pterm.DefaultBasicText.Printf("  1. Bootstrap platform:   openframe bootstrap\n")
	pterm.DefaultBasicText.Printf("  2. Check cluster nodes:  kubectl get nodes\n")
	pterm.DefaultBasicText.Printf("  3. View cluster status:  openframe cluster status %s\n", clusterName)
	pterm.DefaultBasicText.Printf("  4. View running pods:    kubectl get pods -A\n")

	pterm.DefaultBasicText.Println()
}

// ShowClusterStatus handles cluster status display logic
func (s *ClusterService) ShowClusterStatus(name string, detailed bool, skipApps bool, verbose bool) error {
	ctx := context.Background()

	// Get cluster status
	status, err := s.manager.GetClusterStatus(ctx, name)
	if err != nil {
		// Check if it's a "cluster not found" error and handle it friendly
		if strings.Contains(err.Error(), "not found") {
			// Show friendly "cluster not found" message only in interactive terminals
			if isTerminalEnvironment() {
				pterm.DefaultBasicText.Println()

				// Get list of available clusters to show user their options
				clusters, listErr := s.manager.ListClusters(ctx)

				var boxContent string
				if listErr == nil && len(clusters) > 0 {
					// Show available clusters
					boxContent = fmt.Sprintf(
						"Cluster '%s' not found\n\n"+
							"Available clusters:",
						name,
					)
					for _, cluster := range clusters {
						boxContent += fmt.Sprintf("\n  %s", cluster.Name)
					}
				} else {
					// No clusters available
					boxContent = fmt.Sprintf(
						"Cluster '%s' not found\n\n"+
							"No clusters available\n\n"+
							"Create one: openframe cluster create",
						name,
					)
				}

				pterm.DefaultBox.
					WithTitle(" ❓ Cluster Not Found ").
					WithTitleTopCenter().
					Println(boxContent)
			}

			// Always return error for programmatic use and automation
			return fmt.Errorf("cluster '%s' not found", name)
		}

		// For other errors, return the original error
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	// Display comprehensive cluster status
	s.displayDetailedClusterStatus(status, detailed, verbose)

	return nil
}

// displayDetailedClusterStatus shows comprehensive cluster information
func (s *ClusterService) displayDetailedClusterStatus(status models.ClusterInfo, detailed bool, verbose bool) {
	pterm.DefaultBasicText.Println()

	// Main cluster information box
	statusDisplay := fmt.Sprintf("Ready (%s)", status.Status)
	if status.Status != "1/1" {
		statusDisplay = fmt.Sprintf("Partial (%s)", status.Status)
	}

	// Calculate age
	ageStr := "Unknown"
	if !status.CreatedAt.IsZero() {
		duration := time.Since(status.CreatedAt)
		if duration.Hours() < 1 {
			ageStr = fmt.Sprintf("%.0f minutes ago", duration.Minutes())
		} else if duration.Hours() < 24 {
			ageStr = fmt.Sprintf("%.1f hours ago", duration.Hours())
		} else {
			days := int(duration.Hours() / 24)
			ageStr = fmt.Sprintf("%d days ago", days)
		}
	}

	endpoint := s.apiServerEndpoint(context.Background(), status.Name)
	boxContent := fmt.Sprintf(
		"NAME:     %s\n"+
			"TYPE:     %s\n"+
			"STATUS:   %s\n"+
			"NODES:    %d\n"+
			"NETWORK:  k3d-%s\n"+
			"%s\n"+
			"AGE:      %s",
		pterm.Bold.Sprint(status.Name),
		strings.ToUpper(string(status.Type)),
		statusDisplay,
		status.NodeCount,
		status.Name,
		apiServerLine(endpoint),
		ageStr,
	)

	pterm.DefaultBox.
		WithTitle(" 📊 Cluster Status ").
		WithTitleTopCenter().
		Println(boxContent)

	// Network information
	pterm.DefaultBasicText.Println()
	pterm.Info.Printf("🌐 Network Information:\n")
	pterm.DefaultBasicText.Printf("  Network:    k3d-%s\n", status.Name)
	if endpoint != "" {
		pterm.DefaultBasicText.Printf("  API Server: %s\n", endpoint)
	}
	pterm.DefaultBasicText.Printf("  Kubeconfig: %s\n", k8s.DefaultKubeconfigPath())

	// --detailed lists the nodes the provider actually reported. It used to
	// print fixed CPU/Memory/Storage figures ("0.2 cores (10%)", "512MB (5%)",
	// "2.1GB (local)") that were never measured — identical for every cluster,
	// on every machine, at every point in time. The CLI does not collect
	// metrics, so it says so and points at the tool that does.
	if detailed {
		pterm.DefaultBasicText.Println()
		pterm.Info.Printf("🖥️ Nodes:\n")
		if len(status.Nodes) == 0 {
			pterm.DefaultBasicText.Printf("  (none reported)\n")
		}
		for _, node := range status.Nodes {
			pterm.DefaultBasicText.Printf("  %-28s %-8s %s\n", node.Name, node.Role, node.Status)
		}

		pterm.DefaultBasicText.Println()
		pterm.Info.Printf("💾 Resource Usage:\n")
		pterm.DefaultBasicText.Printf("  Not collected by the CLI. With metrics-server installed:\n")
		pterm.DefaultBasicText.Printf("    kubectl top nodes\n")
		pterm.DefaultBasicText.Printf("    kubectl top pods -A\n")
	}

	// Management commands
	pterm.DefaultBasicText.Println()
	pterm.Info.Printf("⚙️ Management Commands:\n")
	pterm.DefaultBasicText.Printf("  Delete cluster:      openframe cluster delete %s\n", status.Name)
	pterm.DefaultBasicText.Printf("  Access with kubectl: kubectl get nodes\n")
	pterm.DefaultBasicText.Printf("  View pods:           kubectl get pods -A\n")
	pterm.DefaultBasicText.Printf("  Get cluster info:    kubectl cluster-info\n")
}

// DisplayClusterList handles cluster list display logic
func (s *ClusterService) DisplayClusterList(clusters []models.ClusterInfo, quiet bool, verbose bool) error {
	if len(clusters) == 0 {
		if quiet {
			// In quiet mode, just exit silently if no clusters
			return nil
		}
		// Use the OperationsUI for consistent messaging
		operationsUI := uiCluster.NewOperationsUI()
		operationsUI.ShowNoResourcesMessage("clusters", "list")
		return nil
	}

	if quiet {
		// In quiet mode, only show cluster names
		for _, cluster := range clusters {
			fmt.Println(cluster.Name)
		}
		return nil
	}

	// Convert to UI display format
	displayClusters := make([]uiCluster.ClusterDisplayInfo, len(clusters))
	for i, cluster := range clusters {
		displayClusters[i] = uiCluster.ClusterDisplayInfo{
			Name:      cluster.Name,
			Type:      string(cluster.Type),
			Status:    cluster.Status,
			NodeCount: cluster.NodeCount,
			CreatedAt: cluster.CreatedAt,
		}
	}

	// Use UI service to display the list
	displayService := uiCluster.NewDisplayService()
	displayService.ShowClusterList(displayClusters, os.Stdout)

	// Show additional info if verbose
	if verbose {
		pterm.DefaultBasicText.Println()
		pterm.Info.Println("Use 'openframe cluster status <name>' for detailed cluster information")
	}

	return nil
}

// CreateClusterWithPrerequisitesNonInteractive creates a cluster with non-interactive support
// Returns the *rest.Config for the created cluster
func CreateClusterWithPrerequisitesNonInteractive(ctx context.Context, clusterName string, verbose bool, nonInteractive bool) (*rest.Config, error) {
	// Show logo first, then check prerequisites (consistent with individual commands)
	ui.ShowLogo()

	// Check prerequisites using the installer directly. OR the flag with
	// environment detection (CI / piped stdin) so a forgotten --non-interactive
	// in CI cannot reach an interactive confirm that hangs the job — same rule
	// as the chart-side gate (chart_service.go).
	installer := prerequisites.NewInstaller()
	if err := installer.CheckAndInstallNonInteractive(nonInteractive || ui.IsNonInteractive()); err != nil {
		return nil, err
	}

	// Create service directly without using utils to avoid circular import
	exec := executor.NewRealCommandExecutor(false, verbose) // dryRun = false
	// Use regular service (with spinner) for interactive mode, suppressed for non-interactive
	var service *ClusterService
	if nonInteractive {
		service = NewClusterServiceSuppressed(exec)
	} else {
		service = NewClusterService(exec)
	}

	// Build cluster configuration
	config := models.ClusterConfig{
		Name:       clusterName,
		Type:       models.ClusterTypeK3d,
		K8sVersion: "",
		NodeCount:  4,
	}
	if clusterName == "" {
		config.Name = "openframe-dev" // default name
	}

	// Create the cluster and return the rest.Config
	return service.CreateCluster(ctx, config)
}
