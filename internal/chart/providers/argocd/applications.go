package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/platform"
	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// applicationGVR is the GroupVersionResource for ArgoCD Application CRDs.
// We access Applications via the dynamic client (unstructured) instead of the
// argo-cd Go module: argo-cd is not importable as a library (its go.mod uses a
// local `replace => ./gitops-engine`), and the dynamic client keeps us
// compatible with any deployed ArgoCD version.
var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

// Manager handles ArgoCD-specific operations
type Manager struct {
	executor    executor.CommandExecutor
	clusterName string // Optional cluster name for explicit context (e.g., "k3d-openframe")

	// Native Kubernetes clients for direct API access (reduces kubectl dependency)
	kubeConfig         *rest.Config
	kubeClient         kubernetes.Interface
	apiextClient       apiextensionsclientset.Interface
	dynamicClient      dynamic.Interface
	clientsInitialized bool

	// StabilizationChecks is the number of consecutive all-ready polls required
	// before declaring success. Defaults to 15 (~30s at 2s interval).
	// Tests can override this to a smaller value for speed.
	StabilizationChecks int

	// syncWait bounds how long RefreshAndSync waits for a hard refresh to be
	// processed and for any in-flight operation to clear. Zero means the default
	// (30s). Tests set a tiny value for speed.
	syncWait time.Duration

	// waitTimeout overrides how long WaitForApplications waits for every app to
	// become Healthy+Synced. Zero means the default (60m, sized for a fresh
	// install). The force-sync path sets a shorter value — re-syncing an existing
	// platform should not block for an hour.
	waitTimeout time.Duration
}

// WithWaitTimeout sets a custom WaitForApplications timeout and returns the
// Manager for chaining. Zero keeps the default (60m).
func (m *Manager) WithWaitTimeout(d time.Duration) *Manager {
	m.waitTimeout = d
	return m
}

// NewManager creates a new ArgoCD manager
func NewManager(exec executor.CommandExecutor) *Manager {
	return &Manager{
		executor: exec,
	}
}

// NewManagerWithCluster creates a new ArgoCD manager with explicit cluster context
func NewManagerWithCluster(exec executor.CommandExecutor, clusterName string) *Manager {
	return &Manager{
		executor:    exec,
		clusterName: clusterName,
	}
}

// NewManagerWithConfig creates a new ArgoCD manager with pre-configured Kubernetes clients
// This is the preferred constructor when you already have a *rest.Config (e.g., after k3d cluster creation)
func NewManagerWithConfig(exec executor.CommandExecutor, config *rest.Config) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("rest.Config cannot be nil")
	}

	// CRITICAL FIX: Bypass TLS Verification for local k3d clusters
	// Uses Insecure=true with CA data cleared, preserving client cert authentication.
	// Applied here as defense-in-depth in case the caller's config doesn't have it set.
	config = sharedconfig.ApplyInsecureTLSConfig(config)

	m := &Manager{
		executor:   exec,
		kubeConfig: config,
	}

	// Create core Kubernetes client
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	m.kubeClient = kubeClient

	// Create API extensions client (for CRD operations)
	apiextClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiextensions client: %w", err)
	}
	m.apiextClient = apiextClient

	// Create dynamic client (for ArgoCD Application CRDs)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	m.dynamicClient = dynamicClient

	m.clientsInitialized = true
	return m, nil
}

// initKubernetesClients initializes the native Kubernetes clients
// This is called lazily when native client operations are needed
func (m *Manager) initKubernetesClients() error {
	if m.clientsInitialized {
		return nil
	}

	// Build kubeconfig path
	kubeconfigPath := getKubeconfigPath()

	// Build config with explicit context if cluster name is set
	var kubeContext string
	if m.clusterName != "" {
		kubeContext = k8s.ResolveContextForCluster(kubeconfigPath, m.clusterName)
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// CRITICAL FIX: Bypass TLS Verification for local k3d clusters
	// Uses custom HTTP transport to bypass TLS at the deepest level.
	config = sharedconfig.ApplyInsecureTransport(config)

	// On Windows, normalize the host to 127.0.0.1 if needed
	if runtime.GOOS == "windows" && strings.Contains(config.Host, "host.docker.internal") {
		// Extract port and use 127.0.0.1
		parts := strings.Split(config.Host, ":")
		if len(parts) >= 3 {
			port := parts[len(parts)-1]
			config.Host = fmt.Sprintf("https://127.0.0.1:%s", port)
		}
	}

	m.kubeConfig = config

	// Create core Kubernetes client
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	m.kubeClient = kubeClient

	// Create API extensions client (for CRD operations)
	apiextClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}
	m.apiextClient = apiextClient

	// Create dynamic client (for ArgoCD Application CRDs)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	m.dynamicClient = dynamicClient

	m.clientsInitialized = true
	return nil
}

// getKubeconfigPath returns the kubeconfig file path
func getKubeconfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return clientcmd.RecommendedHomeFile
	}

	return filepath.Join(homeDir, ".kube", "config")
}

// SetClusterName sets the cluster name for explicit context usage
func (m *Manager) SetClusterName(name string) {
	m.clusterName = name
}

// Application represents an ArgoCD application status
type Application struct {
	Name             string
	Health           string
	HealthMessage    string // Detailed health message
	Sync             string
	SyncRevision     string // Git revision being synced
	Condition        string // Status condition message (e.g., error messages from repo-server)
	ConditionType    string // Type of condition (e.g., "ComparisonError", "InvalidSpecError")
	OperationPhase   string // Operation phase (e.g., "Running", "Failed", "Succeeded")
	OperationMessage string // Operation error message
	RepoURL          string // Source repository URL
	Path             string // Path in repository
	TargetRevision   string // Target revision (branch/tag)
	ReconciledAt     string // Last reconciliation time
}

// argoApp represents the minimal ArgoCD application structure for JSON parsing.
// It is populated either from `kubectl get applications -o json` or from the
// dynamic client's unstructured objects (via a JSON round-trip).
type argoApp struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Health struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"health"`
		Sync struct {
			Status   string `json:"status"`
			Revision string `json:"revision"`
		} `json:"sync"`
		Conditions []struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"conditions"`
		OperationState struct {
			Phase   string `json:"phase"`
			Message string `json:"message"`
		} `json:"operationState"`
		// Resources are the child resources planned/managed by an app (used to
		// count Applications created by the app-of-apps).
		Resources []struct {
			Kind string `json:"kind"`
		} `json:"resources"`
		ReconciledAt string `json:"reconciledAt"`
	} `json:"status"`
	Spec struct {
		Source struct {
			RepoURL        string `json:"repoURL"`
			Path           string `json:"path"`
			TargetRevision string `json:"targetRevision"`
		} `json:"source"`
		Destination struct {
			Namespace string `json:"namespace"`
		} `json:"destination"`
	} `json:"spec"`
}

// argoAppFromObject converts a dynamic-client unstructured object into an
// argoApp via a JSON round-trip, so the native and kubectl paths share the
// exact same field extraction.
func argoAppFromObject(obj map[string]interface{}) (argoApp, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return argoApp{}, err
	}
	var a argoApp
	if err := json.Unmarshal(data, &a); err != nil {
		return argoApp{}, err
	}
	return a, nil
}

// applicationFromArgoApp maps the parsed JSON structure to the public
// Application type (shared by the dynamic-client and kubectl paths).
func applicationFromArgoApp(item argoApp) Application {
	health := item.Status.Health.Status
	sync := item.Status.Sync.Status
	if health == "" {
		health = "Unknown"
	}
	if sync == "" {
		sync = "Unknown"
	}

	condition := ""
	conditionType := ""
	for _, cond := range item.Status.Conditions {
		if cond.Message != "" {
			condition = cond.Message
			conditionType = cond.Type
			break // Take the first condition message
		}
	}

	return Application{
		Name:             item.Metadata.Name,
		Health:           health,
		HealthMessage:    item.Status.Health.Message,
		Sync:             sync,
		SyncRevision:     item.Status.Sync.Revision,
		Condition:        condition,
		ConditionType:    conditionType,
		OperationPhase:   item.Status.OperationState.Phase,
		OperationMessage: item.Status.OperationState.Message,
		RepoURL:          item.Spec.Source.RepoURL,
		Path:             item.Spec.Source.Path,
		TargetRevision:   item.Spec.Source.TargetRevision,
		ReconciledAt:     item.Status.ReconciledAt,
	}
}

// getTotalExpectedApplications tries to determine the total number of applications that will be created
// This function prioritizes native Go client calls over kubectl shell commands for better performance
func (m *Manager) getTotalExpectedApplications(ctx context.Context, config config.ChartInstallConfig) int {
	// Set cluster name from config if available
	if config.ClusterName != "" && m.clusterName == "" {
		m.clusterName = config.ClusterName
	}

	// Best-effort upfront count via the native dynamic client; 0 means "unknown"
	// and the caller discovers the count dynamically while polling.
	if err := m.initKubernetesClients(); err != nil || m.dynamicClient == nil {
		if config.Verbose {
			pterm.Debug.Printf("Native client unavailable for upfront app count: %v\n", err)
		}
		return 0
	}

	apps := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)

	// --- Primary Method: Native dynamic client ---

	// Method 1: Get the root app-of-apps and count Application resources from its status
	if obj, err := apps.Get(ctx, AppOfAppsName, metav1.GetOptions{}); err == nil {
		if app, cerr := argoAppFromObject(obj.Object); cerr == nil {
			appCount := 0
			for _, res := range app.Status.Resources {
				if res.Kind == "Application" {
					appCount++
				}
			}
			if appCount > 0 {
				if config.Verbose {
					pterm.Debug.Printf("Detected %d applications planned by app-of-apps (via native client)\n", appCount)
				}
				return appCount
			}
		}
	}

	// Method 2: List all applications directly via native client
	if list, err := apps.List(ctx, metav1.ListOptions{}); err == nil && len(list.Items) > 0 {
		// Count all apps except the root app-of-apps itself
		count := 0
		for _, item := range list.Items {
			if item.GetName() != AppOfAppsName {
				count++
			}
		}
		if count > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Found %d ArgoCD applications (via native client)\n", count)
			}
			return count
		}
	}

	// Default: return 0 to indicate unknown, will be discovered dynamically
	if config.Verbose {
		pterm.Debug.Println("Could not determine total expected applications upfront, will discover dynamically")
	}

	return 0
}

// ListApplications returns the ArgoCD applications currently in the cluster
// along with their health and sync status.
func (m *Manager) ListApplications(ctx context.Context, verbose bool) ([]Application, error) {
	return m.parseApplications(ctx, verbose)
}

// AdminPassword returns the initial ArgoCD admin password read from the
// argocd-initial-admin-secret. It errors if the secret is absent (ArgoCD not
// installed, or the secret was rotated/removed).
func (m *Manager) AdminPassword(ctx context.Context) (string, error) {
	if m.kubeClient == nil {
		if err := m.initKubernetesClients(); err != nil {
			return "", err
		}
	}
	if m.kubeClient == nil {
		return "", fmt.Errorf("kubernetes client not available")
	}
	secret, err := m.kubeClient.CoreV1().Secrets(ArgoCDNamespace).Get(ctx, "argocd-initial-admin-secret", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("reading argocd-initial-admin-secret: %w", err)
	}
	pw, ok := secret.Data["password"]
	if !ok || len(pw) == 0 {
		return "", fmt.Errorf("argocd-initial-admin-secret has no password field")
	}
	return string(pw), nil
}

// DeleteApplications deletes every ArgoCD Application in the argocd namespace
// and returns the count deleted. Deleting the Application CRs (with ArgoCD's
// resources finalizer) is what cascades removal of the workloads they manage, so
// this must run while ArgoCD is still installed. It is a no-op when there are no
// applications.
func (m *Manager) DeleteApplications(ctx context.Context) (int, error) {
	if m.dynamicClient == nil {
		if err := m.initKubernetesClients(); err != nil {
			return 0, err
		}
	}
	if m.dynamicClient == nil {
		return 0, fmt.Errorf("dynamic client not available")
	}

	res := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)
	list, err := res.List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return 0, nil // the CRD/namespace is already gone
		}
		return 0, fmt.Errorf("listing applications: %w", err)
	}

	deleted := 0
	for i := range list.Items {
		name := list.Items[i].GetName()
		if derr := res.Delete(ctx, name, metav1.DeleteOptions{}); derr != nil && !apierrors.IsNotFound(derr) {
			return deleted, fmt.Errorf("deleting application %q: %w", name, derr)
		}
		deleted++
	}
	return deleted, nil
}

// DeleteNamespace deletes a namespace, treating "not found" as success.
func (m *Manager) DeleteNamespace(ctx context.Context, name string) error {
	if m.kubeClient == nil {
		if err := m.initKubernetesClients(); err != nil {
			return err
		}
	}
	if m.kubeClient == nil {
		return fmt.Errorf("kubernetes client not available")
	}
	if err := m.kubeClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting namespace %q: %w", name, err)
	}
	return nil
}

// parseApplications gets ArgoCD applications and their status using the native
// dynamic client. This reduces reliance on the external kubectl binary.
func (m *Manager) parseApplications(ctx context.Context, verbose bool) ([]Application, error) {
	// On Windows the cluster lives in WSL2 and must be reached from inside WSL.
	if err := platform.WSLClusterHint("list ArgoCD applications"); err != nil {
		return nil, err
	}

	// Initialize clients if needed
	if err := m.initKubernetesClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize the Kubernetes client: %w", err)
	}
	if m.dynamicClient == nil {
		return nil, fmt.Errorf("kubernetes dynamic client unavailable: cannot reach the cluster to list ArgoCD applications")
	}

	// Use the dynamic client to list Application CRDs
	list, err := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to list Argo CD applications via native client: %v\n", err)
		}
		return []Application{}, fmt.Errorf("native ArgoCD client list failed: %w", err)
	}

	apps := make([]Application, 0, len(list.Items))
	for i := range list.Items {
		item, cerr := argoAppFromObject(list.Items[i].Object)
		if cerr != nil {
			if verbose {
				pterm.Warning.Printf("Failed to parse application %q: %v\n", list.Items[i].GetName(), cerr)
			}
			continue
		}
		apps = append(apps, applicationFromArgoApp(item))
	}

	return apps, nil
}
