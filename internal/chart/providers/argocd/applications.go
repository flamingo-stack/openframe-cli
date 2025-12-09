package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	argocdclientset "github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Manager handles ArgoCD-specific operations
type Manager struct {
	executor    executor.CommandExecutor
	clusterName string // Optional cluster name for explicit context (e.g., "k3d-openframe")

	// Native Kubernetes clients for direct API access (reduces kubectl dependency)
	kubeConfig       *rest.Config
	kubeClient       kubernetes.Interface
	apiextClient     apiextensionsclientset.Interface
	argocdClient     argocdclientset.Interface
	clientsInitialized bool
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

	// Create ArgoCD client
	argocdClient, err := argocdclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ArgoCD client: %w", err)
	}
	m.argocdClient = argocdClient

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
		kubeContext = "k3d-" + m.clusterName
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

	// Create ArgoCD client
	argocdClient, err := argocdclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD client: %w", err)
	}
	m.argocdClient = argocdClient

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

// getKubectlArgs returns kubectl args with explicit context if cluster name is set
func (m *Manager) getKubectlArgs(args ...string) []string {
	if m.clusterName != "" {
		contextName := "k3d-" + m.clusterName
		return append([]string{"--context", contextName}, args...)
	}
	return args
}

// Application represents an ArgoCD application status
type Application struct {
	Name   string
	Health string
	Sync   string
}

// argoAppList is used for JSON parsing of ArgoCD applications from kubectl output
type argoAppList struct {
	Items []argoApp `json:"items"`
}

// argoApp represents the minimal ArgoCD application structure for JSON parsing
type argoApp struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
		Sync struct {
			Status string `json:"status"`
		} `json:"sync"`
	} `json:"status"`
}

// getTotalExpectedApplications tries to determine the total number of applications that will be created
// This function prioritizes native Go client calls over kubectl shell commands for better performance
func (m *Manager) getTotalExpectedApplications(ctx context.Context, config config.ChartInstallConfig) int {
	// Set cluster name from config if available
	if config.ClusterName != "" && m.clusterName == "" {
		m.clusterName = config.ClusterName
	}

	// On Windows/WSL2, always use kubectl fallback because:
	// - The native Go client connects to 127.0.0.1:6550 from Windows
	// - But this port is only accessible from inside WSL where k3d runs
	// - kubectl works because it runs via WSL wrapper (wsl -d Ubuntu kubectl...)
	if runtime.GOOS == "windows" {
		return m.getTotalExpectedApplicationsViaKubectl(ctx, config)
	}

	// Initialize clients if needed
	if err := m.initKubernetesClients(); err != nil {
		if config.Verbose {
			pterm.Debug.Printf("Could not initialize native clients: %v\n", err)
		}
		return m.getTotalExpectedApplicationsViaKubectl(ctx, config)
	}

	if m.argocdClient == nil {
		return m.getTotalExpectedApplicationsViaKubectl(ctx, config)
	}

	// --- Primary Method: Native ArgoCD Client ---

	// Method 1: Get app-of-apps and count Application resources from its status
	app, err := m.argocdClient.ArgoprojV1alpha1().Applications("argocd").Get(ctx, "app-of-apps", metav1.GetOptions{})
	if err == nil {
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

	// Method 2: List all applications directly via native client
	appList, err := m.argocdClient.ArgoprojV1alpha1().Applications("argocd").List(ctx, metav1.ListOptions{})
	if err == nil && len(appList.Items) > 0 {
		// Count all apps except app-of-apps itself
		count := 0
		for _, a := range appList.Items {
			if a.Name != "app-of-apps" {
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

// getTotalExpectedApplicationsViaKubectl is the fallback method using kubectl commands
func (m *Manager) getTotalExpectedApplicationsViaKubectl(ctx context.Context, config config.ChartInstallConfig) int {
	// Fallback Method 1: Get all resources that app-of-apps will create from its status via kubectl
	manifestResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io", "app-of-apps",
		"-o", "jsonpath={.status.resources[?(@.kind=='Application')].name}")...)

	if err == nil && manifestResult.Stdout != "" {
		resources := strings.Fields(manifestResult.Stdout)
		if len(resources) > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Detected %d applications planned by app-of-apps (via kubectl)\n", len(resources))
			}
			return len(resources)
		}
	}

	// Fallback Method 2: Try to get all applications including those being created
	// Use -o json to avoid Windows WSL escaping issues with jsonpath
	allAppsResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io",
		"-o", "json")...)

	if err == nil && allAppsResult.Stdout != "" {
		var appList argoAppList
		if err := json.Unmarshal([]byte(allAppsResult.Stdout), &appList); err == nil && len(appList.Items) > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Found %d total ArgoCD applications (via kubectl)\n", len(appList.Items))
			}
			return len(appList.Items)
		}
	}

	// Default: return 0 to indicate unknown
	if config.Verbose {
		pterm.Debug.Println("Could not determine total expected applications upfront, will discover dynamically")
	}

	return 0
}

// parseApplications gets ArgoCD applications and their status using native ArgoCD client
// This reduces reliance on external kubectl binary
func (m *Manager) parseApplications(ctx context.Context, verbose bool) ([]Application, error) {
	// On Windows/WSL2, always use kubectl fallback because:
	// - The native Go client connects to 127.0.0.1:6550 from Windows
	// - But this port is only accessible from inside WSL where k3d runs
	// - kubectl works because it runs via WSL wrapper (wsl -d Ubuntu kubectl...)
	if runtime.GOOS == "windows" {
		return m.parseApplicationsViaKubectl(ctx, verbose)
	}

	// Initialize clients if needed
	if err := m.initKubernetesClients(); err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to initialize native clients, falling back to kubectl: %v\n", err)
		}
		return m.parseApplicationsViaKubectl(ctx, verbose)
	}

	if m.argocdClient == nil {
		return m.parseApplicationsViaKubectl(ctx, verbose)
	}

	// Use native ArgoCD client to list applications
	appList, err := m.argocdClient.ArgoprojV1alpha1().Applications("argocd").List(ctx, metav1.ListOptions{})
	if err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to list Argo CD applications via native client: %v\n", err)
		}
		// Return empty list on failure, allowing the wait loop to continue trying
		return []Application{}, nil
	}

	apps := make([]Application, 0, len(appList.Items))

	for _, argoApp := range appList.Items {
		health := "Unknown"
		sync := "Unknown"

		// Safely extract Health and Sync status from the Go struct
		if argoApp.Status.Health.Status != "" {
			health = string(argoApp.Status.Health.Status)
		}
		if argoApp.Status.Sync.Status != "" {
			sync = string(argoApp.Status.Sync.Status)
		}

		app := Application{
			Name:   argoApp.Name,
			Health: health,
			Sync:   sync,
		}
		apps = append(apps, app)
	}

	return apps, nil
}

// parseApplicationsViaKubectl is the fallback method using kubectl
func (m *Manager) parseApplicationsViaKubectl(ctx context.Context, verbose bool) ([]Application, error) {
	// Use -o json to avoid Windows WSL escaping issues with jsonpath
	result, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io",
		"-o", "json")...)

	if err != nil {
		if verbose {
			pterm.Warning.Printf("kubectl get applications failed: %v\n", err)
		}
		// Return the error so the caller can detect cluster connectivity issues
		return []Application{}, fmt.Errorf("kubectl execution failed: %w", err)
	}

	// Check for connection refused or other cluster connectivity errors in output
	combinedOutput := result.Stdout + result.Stderr
	if strings.Contains(combinedOutput, "connection refused") ||
		strings.Contains(combinedOutput, "Unable to connect to the server") ||
		strings.Contains(combinedOutput, "was refused") ||
		strings.Contains(combinedOutput, "no such host") {
		errMsg := "cluster connectivity error"
		if result.Stderr != "" {
			errMsg = result.Stderr
		}
		return []Application{}, fmt.Errorf("cluster unreachable: %s", errMsg)
	}

	var appList argoAppList
	if err := json.Unmarshal([]byte(result.Stdout), &appList); err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to parse applications JSON: %v\n", err)
		}
		return []Application{}, nil
	}

	apps := make([]Application, 0, len(appList.Items))
	for _, item := range appList.Items {
		health := item.Status.Health.Status
		sync := item.Status.Sync.Status

		if health == "" {
			health = "Unknown"
		}
		if sync == "" {
			sync = "Unknown"
		}

		apps = append(apps, Application{
			Name:   item.Metadata.Name,
			Health: health,
			Sync:   sync,
		})
	}

	return apps, nil
}
