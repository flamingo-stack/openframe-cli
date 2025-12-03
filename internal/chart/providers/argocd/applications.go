package argocd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
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

// getTotalExpectedApplications tries to determine the total number of applications that will be created
func (m *Manager) getTotalExpectedApplications(ctx context.Context, config config.ChartInstallConfig) int {
	// Set cluster name from config if available
	if config.ClusterName != "" && m.clusterName == "" {
		m.clusterName = config.ClusterName
	}

	// Method 1: Use native ArgoCD client to get app-of-apps and count Application resources
	if err := m.initKubernetesClients(); err == nil && m.argocdClient != nil {
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
	}

	// Fallback Method 1: Get all resources that app-of-apps will create from its status via kubectl
	// This shows ALL planned applications across all sync waves
	manifestResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io", "app-of-apps",
		"-o", "jsonpath={.status.resources[?(@.kind=='Application')].name}")...)

	if err == nil && manifestResult.Stdout != "" {
		resources := strings.Fields(manifestResult.Stdout)
		if len(resources) > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Detected %d applications planned by app-of-apps\n", len(resources))
			}
			return len(resources)
		}
	}

	// Method 2: Get the source manifest from app-of-apps and count applications
	// This gives us the definitive count from the source repository
	sourceResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io", "app-of-apps",
		"-o", "jsonpath={.spec.source}")...)

	if err == nil && sourceResult.Stdout != "" && config.Verbose {
		pterm.Debug.Printf("App-of-apps source: %s\n", sourceResult.Stdout)
	}

	// Method 3: Try to get the complete resource list from app-of-apps status
	// This includes all resources that will be created, not just current ones
	allResourcesResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io", "app-of-apps",
		"-o", "jsonpath={range .status.resources[*]}{.kind}{\":\"}{.name}{\"\\n\"}{end}")...)

	if err == nil && allResourcesResult.Stdout != "" {
		lines := strings.Split(strings.TrimSpace(allResourcesResult.Stdout), "\n")
		appCount := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "Application:") {
				appCount++
			}
		}
		if appCount > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Found %d total Application resources in app-of-apps status\n", appCount)
			}
			return appCount
		}
	}

	// Method 4: Check ArgoCD server API for planned applications
	// Query the ArgoCD server pod directly for application information
	serverPod, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "pod",
		"-l", "app.kubernetes.io/name=argocd-server", "-o", "jsonpath={.items[0].metadata.name}")...)

	if err == nil && serverPod.Stdout != "" {
		podName := strings.TrimSpace(serverPod.Stdout)
		// Try to query ArgoCD's internal application list via kubectl exec
		appsResult, _ := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "exec", podName, "--",
			"argocd", "app", "list", "-o", "name")...)
		if appsResult != nil && appsResult.Stdout != "" {
			apps := strings.Split(strings.TrimSpace(appsResult.Stdout), "\n")
			count := 0
			for _, app := range apps {
				if strings.TrimSpace(app) != "" && app != "app-of-apps" {
					count++
				}
			}
			if count > 0 {
				if config.Verbose {
					pterm.Debug.Printf("Found %d applications via ArgoCD CLI\n", count)
				}
				return count
			}
		}
	}

	// Method 5: Try to get all applications including those being created
	// This includes applications in all states (even those not yet synced due to sync waves)
	allAppsResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io",
		"-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")...)

	if err == nil && allAppsResult.Stdout != "" {
		apps := strings.Split(strings.TrimSpace(allAppsResult.Stdout), "\n")
		// Filter out empty lines and count
		count := 0
		for _, app := range apps {
			if strings.TrimSpace(app) != "" {
				count++
			}
		}
		// If we found a reasonable number of apps, use it
		if count > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Found %d total ArgoCD applications\n", count)
			}
			return count
		}
	}

	// Method 6: Check helm values to count applications defined
	helmResult, err := m.executor.Execute(ctx, "helm", "get", "values", "app-of-apps", "-n", "argocd")
	if err == nil && helmResult.Stdout != "" {
		// Count application definitions in various formats
		// Look for patterns that indicate application definitions
		appPatterns := []string{
			"repoURL:",        // Each app typically has a repoURL
			"targetRevision:", // And a targetRevision
			"- name:",         // Applications might be in a list
		}

		maxCount := 0
		for _, pattern := range appPatterns {
			count := strings.Count(helmResult.Stdout, pattern)
			if count > maxCount {
				maxCount = count
			}
		}

		if maxCount > 0 {
			if config.Verbose {
				pterm.Debug.Printf("Estimated %d applications from helm values\n", maxCount)
			}
			return maxCount
		}
	}

	// Method 7: Check ApplicationSets which generate multiple applications
	appSetResult, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applicationsets.argoproj.io",
		"-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")...)

	if err == nil && appSetResult.Stdout != "" {
		appSets := strings.Split(strings.TrimSpace(appSetResult.Stdout), "\n")
		count := 0
		for _, appSet := range appSets {
			if strings.TrimSpace(appSet) != "" {
				count++
			}
		}
		// Each ApplicationSet typically generates 5-10 applications
		// Use a conservative estimate
		if count > 0 {
			estimated := count * 7
			if config.Verbose {
				pterm.Debug.Printf("Estimated %d applications from %d ApplicationSets\n", estimated, count)
			}
			return estimated
		}
	}

	// Default: return 0 to indicate unknown, will be discovered dynamically
	if config.Verbose {
		pterm.Debug.Println("Could not determine total expected applications upfront, will discover dynamically")
	}

	return 0
}

// parseApplications gets ArgoCD applications and their status using native ArgoCD client
// This reduces reliance on external kubectl binary
func (m *Manager) parseApplications(ctx context.Context, verbose bool) ([]Application, error) {
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
	result, err := m.executor.Execute(ctx, "kubectl", m.getKubectlArgs("-n", "argocd", "get", "applications.argoproj.io",
		"-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.health.status}{\"\\t\"}{.status.sync.status}{\"\\n\"}{end}")...)

	if err != nil {
		if verbose {
			pterm.Warning.Printf("kubectl jsonpath failed: %v\n", err)
		}
		return []Application{}, nil
	}

	apps := make([]Application, 0)
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			health := strings.TrimSpace(parts[1])
			sync := strings.TrimSpace(parts[2])

			if health == "" {
				health = "Unknown"
			}
			if sync == "" {
				sync = "Unknown"
			}

			app := Application{
				Name:   strings.TrimSpace(parts[0]),
				Health: health,
				Sync:   sync,
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}
