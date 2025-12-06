package kind

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/manager"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Constants for configuration
const (
	defaultKindImage   = "kindest/node:v1.31.4"
	defaultTimeout     = "300s"
	defaultAPIPort     = "6443"
	defaultHTTPPort    = "80"
	defaultHTTPSPort   = "443"
	timestampSuffixLen = 6
)

func init() {
	// Register the kind manager factory
	manager.RegisterManager(manager.ManagerTypeKind, func(exec executor.CommandExecutor, verbose bool) manager.ClusterManager {
		return NewKindManager(exec, verbose)
	})
}

// KindManager manages KIND cluster operations on Windows
type KindManager struct {
	executor executor.CommandExecutor
	verbose  bool
	timeout  string
}

// NewKindManager creates a new KIND cluster manager with default timeout
func NewKindManager(exec executor.CommandExecutor, verbose bool) *KindManager {
	return &KindManager{
		executor: exec,
		verbose:  verbose,
		timeout:  defaultTimeout,
	}
}

// NewKindManagerWithTimeout creates a new KIND cluster manager with custom timeout
func NewKindManagerWithTimeout(exec executor.CommandExecutor, verbose bool, timeout string) *KindManager {
	return &KindManager{
		executor: exec,
		verbose:  verbose,
		timeout:  timeout,
	}
}

// CreateCluster creates a new KIND cluster using config file approach
// Returns the *rest.Config for the created cluster that can be used to interact with it
func (m *KindManager) CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error) {
	if err := m.validateClusterConfig(config); err != nil {
		return nil, err
	}

	if config.Type != models.ClusterTypeKind {
		return nil, models.NewProviderNotFoundError(config.Type)
	}

	configFile, err := m.createKindConfigFile(config)
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("failed to create config file: %w", err))
	}
	defer os.Remove(configFile)

	if m.verbose {
		if configContent, err := os.ReadFile(configFile); err == nil {
			fmt.Printf("DEBUG: Config file content for %s:\n%s\n", config.Name, string(configContent))
		}
	}

	// Prepare kubeconfig directory
	if err := m.prepareKubeconfigDirectory(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not prepare kubeconfig directory: %v\n", err)
		}
	}

	// Create the cluster using kind
	args := []string{
		"create", "cluster",
		"--name", config.Name,
		"--config", configFile,
		"--wait", m.timeout,
	}
	if m.verbose {
		args = append(args, "-v", "1")
	}

	if _, err := m.executor.Execute(ctx, "kind", args...); err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("failed to create cluster %s: %w", config.Name, err))
	}

	// Verify the cluster is reachable and get the rest.Config
	restConfig, err := m.verifyClusterReachable(ctx, config.Name)
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("cluster created but not reachable: %w", err))
	}

	return restConfig, nil
}

// GetRestConfig returns the rest.Config for an existing cluster
func (m *KindManager) GetRestConfig(ctx context.Context, clusterName string) (*rest.Config, error) {
	return m.verifyClusterReachable(ctx, clusterName)
}

// DeleteCluster removes a KIND cluster
func (m *KindManager) DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error {
	if name == "" {
		return models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	if clusterType != models.ClusterTypeKind {
		return models.NewProviderNotFoundError(clusterType)
	}

	args := []string{"delete", "cluster", "--name", name}
	if m.verbose {
		args = append(args, "-v", "1")
	}

	if _, err := m.executor.Execute(ctx, "kind", args...); err != nil {
		return models.NewClusterOperationError("delete", name, fmt.Errorf("failed to delete cluster %s: %w", name, err))
	}

	return nil
}

// StartCluster starts a KIND cluster (KIND doesn't support stop/start, so this is a no-op)
func (m *KindManager) StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error {
	if name == "" {
		return models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	if clusterType != models.ClusterTypeKind {
		return models.NewProviderNotFoundError(clusterType)
	}

	// KIND doesn't support stop/start operations
	// Check if the cluster exists and the container is running
	if m.verbose {
		fmt.Printf("KIND clusters cannot be stopped/started - checking if cluster %s exists...\n", name)
	}

	_, err := m.GetClusterStatus(ctx, name)
	if err != nil {
		return fmt.Errorf("cluster %s not found or not running: %w", name, err)
	}

	return nil
}

// ListClusters returns all KIND clusters
func (m *KindManager) ListClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	result, err := m.executor.Execute(ctx, "kind", "get", "clusters")
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	var clusters []models.ClusterInfo
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}

		// Get more detailed info about the cluster
		info, err := m.GetClusterStatus(ctx, name)
		if err != nil {
			// If we can't get status, just add basic info
			clusters = append(clusters, models.ClusterInfo{
				Name:   name,
				Type:   models.ClusterTypeKind,
				Status: "Unknown",
			})
			continue
		}
		clusters = append(clusters, info)
	}

	return clusters, nil
}

// ListAllClusters is an alias for ListClusters for backward compatibility
func (m *KindManager) ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	return m.ListClusters(ctx)
}

// GetClusterStatus returns detailed status for a specific KIND cluster
func (m *KindManager) GetClusterStatus(ctx context.Context, name string) (models.ClusterInfo, error) {
	if name == "" {
		return models.ClusterInfo{}, models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	// Check if cluster exists
	result, err := m.executor.Execute(ctx, "kind", "get", "clusters")
	if err != nil {
		return models.ClusterInfo{}, models.NewClusterOperationError("status", name, err)
	}

	found := false
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == name {
			found = true
			break
		}
	}

	if !found {
		return models.ClusterInfo{}, models.NewClusterOperationError("status", name, fmt.Errorf("cluster %s not found", name))
	}

	// Get node info using kubectl
	nodeCount := 0
	status := "Unknown"

	kubeconfigPath := m.getKubeconfigPath()
	contextName := fmt.Sprintf("kind-%s", name)

	nodeResult, err := m.executor.Execute(ctx, "kubectl", "--kubeconfig", kubeconfigPath, "--context", contextName, "get", "nodes", "-o", "json")
	if err == nil {
		var nodesResponse struct {
			Items []struct {
				Status struct {
					Conditions []struct {
						Type   string `json:"type"`
						Status string `json:"status"`
					} `json:"conditions"`
				} `json:"status"`
			} `json:"items"`
		}

		if json.Unmarshal([]byte(nodeResult.Stdout), &nodesResponse) == nil {
			nodeCount = len(nodesResponse.Items)
			readyCount := 0
			for _, node := range nodesResponse.Items {
				for _, condition := range node.Status.Conditions {
					if condition.Type == "Ready" && condition.Status == "True" {
						readyCount++
						break
					}
				}
			}
			status = fmt.Sprintf("%d/%d", readyCount, nodeCount)
		}
	}

	return models.ClusterInfo{
		Name:      name,
		Type:      models.ClusterTypeKind,
		Status:    status,
		NodeCount: nodeCount,
		Nodes:     []models.NodeInfo{},
	}, nil
}

// DetectClusterType determines if a cluster is KIND
func (m *KindManager) DetectClusterType(ctx context.Context, name string) (models.ClusterType, error) {
	if name == "" {
		return "", models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	result, err := m.executor.Execute(ctx, "kind", "get", "clusters")
	if err != nil {
		return "", models.NewClusterNotFoundError(name)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == name {
			return models.ClusterTypeKind, nil
		}
	}

	return "", models.NewClusterNotFoundError(name)
}

// GetKubeconfig gets the kubeconfig for a specific KIND cluster
func (m *KindManager) GetKubeconfig(ctx context.Context, name string, clusterType models.ClusterType) (string, error) {
	if clusterType != models.ClusterTypeKind {
		return "", models.NewProviderNotFoundError(clusterType)
	}

	args := []string{"get", "kubeconfig", "--name", name}
	result, err := m.executor.Execute(ctx, "kind", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig for cluster %s: %w", name, err)
	}

	return result.Stdout, nil
}

// validateClusterConfig validates the cluster configuration
func (m *KindManager) validateClusterConfig(config models.ClusterConfig) error {
	if config.Name == "" {
		return models.NewInvalidConfigError("name", config.Name, "cluster name cannot be empty")
	}
	if config.Type == "" {
		return models.NewInvalidConfigError("type", config.Type, "cluster type cannot be empty")
	}
	if config.NodeCount < 1 {
		return models.NewInvalidConfigError("nodeCount", config.NodeCount, "node count must be at least 1")
	}
	return nil
}

// createKindConfigFile creates a KIND config file
func (m *KindManager) createKindConfigFile(config models.ClusterConfig) (string, error) {
	image := defaultKindImage
	if config.K8sVersion != "" {
		image = "kindest/node:" + config.K8sVersion
	}

	// KIND config with control-plane and workers
	workers := config.NodeCount - 1
	if workers < 0 {
		workers = 0
	}

	configContent := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: %s
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: %s
    protocol: TCP
  - containerPort: 443
    hostPort: %s
    protocol: TCP
  - containerPort: 6443
    hostPort: %s
    protocol: TCP
`, image, defaultHTTPPort, defaultHTTPSPort, defaultAPIPort)

	// Add worker nodes
	for i := 0; i < workers; i++ {
		configContent += fmt.Sprintf(`- role: worker
  image: %s
`, image)
	}

	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// prepareKubeconfigDirectory ensures ~/.kube directory exists with proper permissions
func (m *KindManager) prepareKubeconfigDirectory(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .kube directory: %w", err)
	}

	if m.verbose {
		fmt.Println("Prepared kubeconfig directory")
	}

	return nil
}

// verifyClusterReachable checks if the cluster is reachable using native Go client
func (m *KindManager) verifyClusterReachable(ctx context.Context, clusterName string) (*rest.Config, error) {
	contextName := fmt.Sprintf("kind-%s", clusterName)
	kubeconfigPath := m.getKubeconfigPath()

	// Load the Kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig file from %s: %w", kubeconfigPath, err)
	}

	// Check if the context exists
	if _, exists := config.Contexts[contextName]; !exists {
		return nil, fmt.Errorf("kubectl context %s not found in kubeconfig", contextName)
	}

	// Switch the current context
	config.CurrentContext = contextName
	if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
		return nil, fmt.Errorf("failed to switch and write kubectl context: %w", err)
	}

	if m.verbose {
		fmt.Printf("Switched kubectl context to %s\n", contextName)
	}

	// Build rest.Config from the loaded Kubeconfig
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build REST config: %w", err)
	}

	// For local clusters, we may need to apply insecure TLS config
	// KIND uses 127.0.0.1 by default which should work without issues
	// But apply it for consistency with k3d
	restConfig = sharedconfig.ApplyInsecureTLSConfig(restConfig)

	if m.verbose {
		fmt.Println("TLS verification bypassed for local KIND cluster")
	}

	// Extract host and port from restConfig.Host for TCP check
	host, port, err := extractHostPort(restConfig.Host)
	if err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not extract host:port from %s: %v\n", restConfig.Host, err)
		}
		host = "127.0.0.1"
		port = defaultAPIPort
	}

	// Wait for TCP port to be available
	tcpRetries := 10
	tcpRetryDelay := 1 * time.Second
	if err := m.waitForTCPPort(ctx, host, port, tcpRetries, tcpRetryDelay); err != nil {
		return nil, fmt.Errorf("API server port not available: %w", err)
	}

	// Create Kubernetes client with the restConfig
	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Verify cluster reachability and node readiness with polling
	maxRetries := 15
	retryDelay := 2 * time.Second
	var lastErr error

	if m.verbose {
		fmt.Println("Waiting for cluster API and nodes to be reachable...")
	}

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		nodes, err := coreClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			if isTemporaryError(err) {
				lastErr = err
				if m.verbose {
					fmt.Printf("  Cluster not ready yet (attempt %d/%d): %v\n", i+1, maxRetries, err)
				}
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to connect to cluster API: %w", err)
		}

		if len(nodes.Items) == 0 {
			lastErr = fmt.Errorf("no nodes found in cluster")
			if m.verbose {
				fmt.Printf("  No nodes found yet (attempt %d/%d), waiting...\n", i+1, maxRetries)
			}
			time.Sleep(retryDelay)
			continue
		}

		readyCount := 0
		for _, node := range nodes.Items {
			for _, condition := range node.Status.Conditions {
				if string(condition.Type) == "Ready" && string(condition.Status) == "True" {
					readyCount++
					break
				}
			}
		}

		if readyCount > 0 {
			if m.verbose {
				fmt.Printf("  Found %d ready node(s) out of %d total\n", readyCount, len(nodes.Items))
				fmt.Println("Cluster API and nodes are ready.")
			}
			return restConfig, nil
		}

		lastErr = fmt.Errorf("no nodes in Ready state (found %d nodes, 0 ready)", len(nodes.Items))
		if m.verbose {
			fmt.Printf("  Nodes exist but none are Ready yet (attempt %d/%d), waiting...\n", i+1, maxRetries)
		}
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("cluster not reachable after %d retries (last error: %w)", maxRetries, lastErr)
}

// waitForTCPPort performs a TCP connectivity check
func (m *KindManager) waitForTCPPort(ctx context.Context, host string, port string, maxRetries int, retryDelay time.Duration) error {
	address := net.JoinHostPort(host, port)

	if m.verbose {
		fmt.Printf("Waiting for TCP port %s to be available...\n", address)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			conn.Close()
			if m.verbose {
				fmt.Printf("TCP port %s is open\n", address)
			}
			return nil
		}

		lastErr = err
		if m.verbose {
			fmt.Printf("  TCP port not ready yet (attempt %d/%d): %v\n", i+1, maxRetries, err)
		}
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("TCP port %s not available after %d retries: %w", address, maxRetries, lastErr)
}

// getKubeconfigPath returns the kubeconfig file path
func (m *KindManager) getKubeconfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return clientcmd.RecommendedHomeFile
	}

	return filepath.Join(homeDir, ".kube", "config")
}

// isTemporaryError checks if an error is temporary and should be retried
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "Service Unavailable") ||
		strings.Contains(errStr, "server is currently unable")
}

// extractHostPort extracts host and port from a URL string
func extractHostPort(urlStr string) (string, string, error) {
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Handle IPv6 addresses in brackets
	if strings.HasPrefix(urlStr, "[") {
		// IPv6 format: [::1]:port
		re := regexp.MustCompile(`\[([^\]]+)\]:(\d+)`)
		match := re.FindStringSubmatch(urlStr)
		if len(match) == 3 {
			return match[1], match[2], nil
		}
	}

	host, port, err := net.SplitHostPort(urlStr)
	if err != nil {
		return urlStr, "", fmt.Errorf("could not split host:port from %s: %w", urlStr, err)
	}

	return host, port, nil
}

// Factory functions for backward compatibility

// CreateClusterManagerWithExecutor creates a KIND cluster manager with a specific command executor
func CreateClusterManagerWithExecutor(exec executor.CommandExecutor) *KindManager {
	if exec == nil {
		panic("Executor cannot be nil")
	}
	return NewKindManager(exec, false)
}
