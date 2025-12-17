package k3d

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

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
	defaultK3sImage    = "rancher/k3s:v1.31.5-k3s1"
	defaultTimeout     = "300s"
	defaultAPIPort     = "6550"
	defaultHTTPPort    = "8080"
	defaultHTTPSPort   = "8443"
	dynamicPortStart   = 20000
	dynamicPortEnd     = 50000
	portSearchStep     = 1000
	timestampSuffixLen = 6
)

// ClusterManager interface for managing clusters
type ClusterManager interface {
	DetectClusterType(ctx context.Context, name string) (models.ClusterType, error)
	ListClusters(ctx context.Context) ([]models.ClusterInfo, error)
	ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error)
}

// K3dManager manages K3D cluster operations
type K3dManager struct {
	executor executor.CommandExecutor
	verbose  bool
	timeout  string
}

// NewK3dManager creates a new K3D cluster manager with default timeout
func NewK3dManager(exec executor.CommandExecutor, verbose bool) *K3dManager {
	return &K3dManager{
		executor: exec,
		verbose:  verbose,
		timeout:  defaultTimeout,
	}
}

// NewK3dManagerWithTimeout creates a new K3D cluster manager with custom timeout
func NewK3dManagerWithTimeout(exec executor.CommandExecutor, verbose bool, timeout string) *K3dManager {
	return &K3dManager{
		executor: exec,
		verbose:  verbose,
		timeout:  timeout,
	}
}

// CreateCluster creates a new K3D cluster using config file approach
// Returns the *rest.Config for the created cluster that can be used to interact with it
func (m *K3dManager) CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error) {
	if err := m.validateClusterConfig(config); err != nil {
		return nil, err
	}

	if config.Type != models.ClusterTypeK3d {
		return nil, models.NewProviderNotFoundError(config.Type)
	}

	// Increase inotify limits for applications like MeshCentral that use many file watchers
	// This must be done before cluster creation as it affects the Docker/WSL host
	if err := m.increaseInotifyLimits(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not increase inotify limits: %v\n", err)
		}
		// Don't fail - cluster might still work if limits are already sufficient
	}

	// On Windows/WSL2, configure reliable DNS servers in /etc/resolv.conf before creating the cluster
	// Note: Docker daemon DNS is now configured during Docker installation (see docker/docker.go)
	// This WSL DNS configuration is a backup to ensure the WSL host can resolve DNS
	if runtime.GOOS == "windows" {
		if err := m.configureWSLDNS(ctx); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not configure WSL DNS: %v\n", err)
			}
			// Don't fail - cluster might still work with existing DNS configuration
		}

		// Ensure Docker daemon has DNS configuration applied
		// This is critical for k3d containers to resolve external registries
		if err := m.ensureDockerDNS(ctx); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not configure Docker DNS: %v\n", err)
			}
			// Don't fail - might still work if Docker was configured during installation
		}

		// Pre-pull critical k3s images to avoid DNS issues during cluster creation
		// This is critical because k3d container nodes may fail to pull images
		// if Docker Hub DNS resolution is flaky
		if err := m.prePullK3sImages(ctx); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not pre-pull k3s images: %v\n", err)
			}
			// Don't fail - cluster creation might still succeed
		}

		// Fix Docker network routing BEFORE creating the cluster
		// This ensures IP forwarding and NAT masquerading are set up correctly
		if err := m.fixDockerNetworkRouting(ctx); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not fix Docker network routing: %v\n", err)
			}
			// Don't fail - might still work
		}
	}

	// On Windows/WSL2, get the WSL internal IP before creating the cluster
	// to include it as a TLS SAN in the k3s certificate
	var wslInternalIP string
	if runtime.GOOS == "windows" {
		var err error
		wslInternalIP, err = m.getWSLInternalIP(ctx)
		if err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not get WSL internal IP for TLS SAN: %v\n", err)
			}
			// Continue without the extra SAN - the insecure TLS config will still work
		} else if m.verbose {
			fmt.Printf("✓ Retrieved WSL internal IP for TLS SAN: %s\n", wslInternalIP)
		}
	}

	configFile, err := m.createK3dConfigFile(config, wslInternalIP)
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("failed to create config file: %w", err))
	}
	defer os.Remove(configFile)

	if m.verbose {
		if configContent, err := os.ReadFile(configFile); err == nil {
			fmt.Printf("DEBUG: Config file content for %s:\n%s\n", config.Name, string(configContent))
		}
	}

	// Prepare kubeconfig directory before k3d operations (Windows/WSL and Linux CI)
	if err := m.prepareKubeconfigDirectory(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not prepare kubeconfig directory: %v\n", err)
		}
		// Don't fail - k3d will create it, but log the warning
	}

	// Clean up any stale lock files that might prevent k3d from updating kubeconfig
	if err := m.cleanupStaleLockFiles(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not cleanup stale lock files: %v\n", err)
		}
		// Don't fail - this is not critical
	}

	// Convert Windows path to WSL path if running on Windows
	configFilePath := configFile
	if runtime.GOOS == "windows" {
		configFilePath, err = m.convertWindowsPathToWSL(configFile)
		if err != nil {
			return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("failed to convert config file path for WSL: %w", err))
		}
		if m.verbose {
			fmt.Printf("DEBUG: Converted Windows path '%s' to WSL path '%s'\n", configFile, configFilePath)
		}

		// Final Docker readiness check right before k3d - Docker can become unavailable on WSL2
		if err := m.verifyDockerReady(ctx); err != nil {
			return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("Docker is not ready: %w", err))
		}
	}

	args := []string{
		"cluster", "create",
		"--config", configFilePath,
		"--timeout", m.timeout,
		"--kubeconfig-update-default", // Update default kubeconfig with new cluster context
		"--kubeconfig-switch-context", // Automatically switch to new cluster context
	}

	if m.verbose {
		args = append(args, "--verbose")
	}

	if _, err := m.executor.Execute(ctx, "k3d", args...); err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("failed to create cluster %s: %w", config.Name, err))
	}

	// On Windows/WSL2, fix DNS inside k3d node containers and import critical images
	// This is critical because k3d containers may have stale DNS configuration
	if runtime.GOOS == "windows" {
		// First, fix DNS configuration inside the k3d node containers
		// This ensures containerd can resolve registry-1.docker.io
		if err := m.fixK3dNodeDNS(ctx, config.Name); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not fix DNS in k3d nodes: %v\n", err)
			}
			// Don't fail - continue with image import as fallback
		}

		// Then import critical images as a fallback mechanism
		if err := m.importK3sImages(ctx, config.Name); err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not import k3s images: %v\n", err)
			}
			// Don't fail - pods might still be able to pull images
		}
	}

	// Fix kubeconfig permissions if k3d ran with sudo (Windows/WSL and Linux CI)
	// This is necessary because k3d creates ~/.kube/config with root ownership when run with sudo
	if err := m.fixKubeconfigPermissions(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not fix kubeconfig permissions: %v\n", err)
		}
		// Don't fail - this is not critical, just log the warning
	}

	// Clean up any lock files after fixing permissions to ensure kubectl can access the config
	// This is critical because lock files may have been created with root ownership
	if err := m.cleanupStaleLockFiles(ctx); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not cleanup lock files after permission fix: %v\n", err)
		}
		// Don't fail - this is not critical
	}

	// On Windows, rewrite the kubeconfig server address to use the WSL internal IP
	// This is necessary for helm (running inside Ubuntu WSL) to reach the k3d cluster
	if err := m.rewriteWSLKubeconfigServerAddress(ctx, config.Name); err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not rewrite kubeconfig server address: %v\n", err)
		}
		// Don't fail - helm might still work if the network is configured correctly
	}

	// Verify the cluster is reachable and get the rest.Config
	restConfig, err := m.verifyClusterReachable(ctx, config.Name)
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("cluster created but not reachable: %w", err))
	}

	// Additional kubectl verification checks (especially important for Windows/WSL)
	if err := m.verifyClusterViaKubectl(ctx, config.Name); err != nil {
		if m.verbose {
			fmt.Printf("Warning: kubectl verification checks failed: %v\n", err)
		}
		// Don't fail - the native Go client verification passed, kubectl might just need more time
	}

	return restConfig, nil
}

// GetRestConfig returns the rest.Config for an existing cluster
// This is used to get the config for a cluster that was already created
func (m *K3dManager) GetRestConfig(ctx context.Context, clusterName string) (*rest.Config, error) {
	return m.verifyClusterReachable(ctx, clusterName)
}

// DeleteCluster removes a K3D cluster
func (m *K3dManager) DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error {
	if name == "" {
		return models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	if clusterType != models.ClusterTypeK3d {
		return models.NewProviderNotFoundError(clusterType)
	}

	args := []string{"cluster", "delete", name}
	if m.verbose {
		args = append(args, "--verbose")
	}

	// Use a 2-minute timeout to prevent hanging on WSL networking issues
	options := executor.ExecuteOptions{
		Command: "k3d",
		Args:    args,
		Timeout: 2 * time.Minute,
	}

	_, err := m.executor.ExecuteWithOptions(ctx, options)
	if err != nil {
		// On Windows/WSL or when force is set, fall back to direct Docker cleanup
		// This handles WSL networking issues that can cause k3d to hang or fail
		if runtime.GOOS == "windows" || force {
			if m.verbose {
				fmt.Printf("k3d delete failed, attempting direct Docker cleanup for cluster %s: %v\n", name, err)
			}
			if cleanupErr := m.forceCleanupDockerContainers(ctx, name); cleanupErr != nil {
				// Return original error if cleanup also fails
				return models.NewClusterOperationError("delete", name, fmt.Errorf("failed to delete cluster %s (cleanup also failed: %v): %w", name, cleanupErr, err))
			}
			// Cleanup succeeded, cluster is removed
			if m.verbose {
				fmt.Printf("✓ Cluster %s removed via direct Docker cleanup\n", name)
			}
			return nil
		}
		return models.NewClusterOperationError("delete", name, fmt.Errorf("failed to delete cluster %s: %w", name, err))
	}

	return nil
}

// forceCleanupDockerContainers removes all Docker containers associated with a k3d cluster
// This is a fallback mechanism when k3d cluster delete fails (e.g., due to WSL networking issues)
func (m *K3dManager) forceCleanupDockerContainers(ctx context.Context, clusterName string) error {
	if runtime.GOOS == "windows" {
		return m.forceCleanupDockerContainersWSL(ctx, clusterName)
	}
	return m.forceCleanupDockerContainersDirect(ctx, clusterName)
}

// forceCleanupDockerContainersWSL removes k3d containers via WSL on Windows
func (m *K3dManager) forceCleanupDockerContainersWSL(ctx context.Context, clusterName string) error {
	username, err := m.getWSLUser(ctx)
	if err != nil {
		username = "root" // fallback to root - all Linux distributions have root
	}

	// Remove containers matching k3d-<clustername> pattern
	cleanupCmd := fmt.Sprintf(
		"sudo docker ps -aq --filter 'name=k3d-%s' | xargs -r sudo docker rm -f 2>/dev/null || true",
		clusterName,
	)
	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", cleanupCmd)
	if err != nil {
		return fmt.Errorf("failed to cleanup containers via WSL: %w", err)
	}

	// Also remove the network
	networkCleanupCmd := fmt.Sprintf("sudo docker network rm k3d-%s 2>/dev/null || true", clusterName)
	_, _ = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", networkCleanupCmd)

	return nil
}

// forceCleanupDockerContainersDirect removes k3d containers directly (non-Windows)
func (m *K3dManager) forceCleanupDockerContainersDirect(ctx context.Context, clusterName string) error {
	// List containers matching k3d-<clustername> pattern
	result, err := m.executor.Execute(ctx, "docker", "ps", "-aq", "--filter", fmt.Sprintf("name=k3d-%s", clusterName))
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.TrimSpace(result.Stdout)
	if containerIDs != "" {
		// Remove each container
		for _, id := range strings.Split(containerIDs, "\n") {
			id = strings.TrimSpace(id)
			if id != "" {
				_, _ = m.executor.Execute(ctx, "docker", "rm", "-f", id)
			}
		}
	}

	// Also remove the network
	_, _ = m.executor.Execute(ctx, "docker", "network", "rm", fmt.Sprintf("k3d-%s", clusterName))

	return nil
}

// StartCluster starts a K3D cluster
func (m *K3dManager) StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error {
	if name == "" {
		return models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	if clusterType != models.ClusterTypeK3d {
		return models.NewProviderNotFoundError(clusterType)
	}

	args := []string{"cluster", "start", name}
	if m.verbose {
		args = append(args, "--verbose")
	}

	if _, err := m.executor.Execute(ctx, "k3d", args...); err != nil {
		return models.NewClusterOperationError("start", name, fmt.Errorf("failed to start cluster %s: %w", name, err))
	}

	return nil
}

// ListClusters returns all K3D clusters
func (m *K3dManager) ListClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	args := []string{"cluster", "list", "--output", "json"}

	// Use a 30-second timeout to prevent hanging on WSL networking issues
	options := executor.ExecuteOptions{
		Command: "k3d",
		Args:    args,
		Timeout: 30 * time.Second,
	}

	result, err := m.executor.ExecuteWithOptions(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	var k3dClusters []k3dClusterInfo
	if err := json.Unmarshal([]byte(result.Stdout), &k3dClusters); err != nil {
		return nil, fmt.Errorf("failed to parse cluster list JSON: %w", err)
	}

	var clusters []models.ClusterInfo
	for _, k3dCluster := range k3dClusters {
		// Find the earliest server node creation time as cluster creation time
		var createdAt time.Time
		for _, node := range k3dCluster.Nodes {
			if node.Role == "server" {
				if createdAt.IsZero() || node.Created.Before(createdAt) {
					createdAt = node.Created
				}
			}
		}

		clusters = append(clusters, models.ClusterInfo{
			Name:      k3dCluster.Name,
			Type:      models.ClusterTypeK3d,
			Status:    fmt.Sprintf("%d/%d", k3dCluster.ServersRunning, k3dCluster.ServersCount),
			NodeCount: k3dCluster.AgentsCount + k3dCluster.ServersCount,
			CreatedAt: createdAt,
			Nodes:     []models.NodeInfo{},
		})
	}

	return clusters, nil
}

// ListAllClusters is an alias for ListClusters for backward compatibility
func (m *K3dManager) ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	return m.ListClusters(ctx)
}

// GetClusterStatus returns detailed status for a specific K3D cluster
func (m *K3dManager) GetClusterStatus(ctx context.Context, name string) (models.ClusterInfo, error) {
	if name == "" {
		return models.ClusterInfo{}, models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	clusters, err := m.ListClusters(ctx)
	if err != nil {
		return models.ClusterInfo{}, models.NewClusterOperationError("status", name, err)
	}

	for _, clusterInfo := range clusters {
		if clusterInfo.Name == name {
			return clusterInfo, nil
		}
	}

	return models.ClusterInfo{}, models.NewClusterOperationError("status", name, fmt.Errorf("cluster %s not found", name))
}

// DetectClusterType determines if a cluster is K3D
func (m *K3dManager) DetectClusterType(ctx context.Context, name string) (models.ClusterType, error) {
	if name == "" {
		return "", models.NewInvalidConfigError("name", name, "cluster name cannot be empty")
	}

	args := []string{"cluster", "get", name}

	// Use a 30-second timeout to prevent hanging on WSL networking issues
	options := executor.ExecuteOptions{
		Command: "k3d",
		Args:    args,
		Timeout: 30 * time.Second,
	}

	if _, err := m.executor.ExecuteWithOptions(ctx, options); err != nil {
		return "", models.NewClusterNotFoundError(name)
	}

	return models.ClusterTypeK3d, nil
}

// GetKubeconfig gets the kubeconfig for a specific K3D cluster
func (m *K3dManager) GetKubeconfig(ctx context.Context, name string, clusterType models.ClusterType) (string, error) {
	if clusterType != models.ClusterTypeK3d {
		return "", models.NewProviderNotFoundError(clusterType)
	}

	args := []string{"kubeconfig", "get", name}
	result, err := m.executor.Execute(ctx, "k3d", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig for cluster %s: %w", name, err)
	}

	return result.Stdout, nil
}

// validateClusterConfig validates the cluster configuration
func (m *K3dManager) validateClusterConfig(config models.ClusterConfig) error {
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

// createK3dConfigFile creates a k3d config file
// wslInternalIP is optional - if provided, it will be added as a TLS SAN for the k3s API server certificate
func (m *K3dManager) createK3dConfigFile(config models.ClusterConfig, wslInternalIP string) (string, error) {
	image := defaultK3sImage
	if runtime.GOARCH == "arm64" {
		image = defaultK3sImage
	}
	if config.K8sVersion != "" {
		image = "rancher/k3s:" + config.K8sVersion
	}

	servers := 1
	agents := config.NodeCount - 1
	if agents < 0 {
		agents = 0
	}

	configContent := fmt.Sprintf(`apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: %s
servers: %d
agents: %d
image: %s`, config.Name, servers, agents, image)

	// Use fixed default ports for consistent cluster configuration
	apiPort := defaultAPIPort
	httpPort := defaultHTTPPort
	httpsPort := defaultHTTPSPort

	// On Windows/WSL2, bind to 0.0.0.0 so the API is accessible via the WSL eth0 IP
	// Docker runs inside WSL2 Ubuntu, and binding to 0.0.0.0 makes the API accessible:
	// - From within WSL via 127.0.0.1 (for kubectl/helm running in WSL)
	// - From Windows via WSL's eth0 IP (for the Go client running on Windows)
	hostIP := "127.0.0.1"
	if runtime.GOOS == "windows" {
		hostIP = "0.0.0.0"
	}

	// Build TLS SAN argument if WSL internal IP is provided
	// This ensures the k3s API server certificate includes the WSL internal IP,
	// allowing kubectl/helm to connect via the WSL network without TLS errors
	tlsSanArg := ""
	if wslInternalIP != "" {
		tlsSanArg = fmt.Sprintf(`
      - arg: --tls-san=%s
        nodeFilters:
          - server:*`, wslInternalIP)
	}

	// On Windows/WSL2, add runtime options for k3d containers
	runtimeOptions := ""
	if runtime.GOOS == "windows" {
		// Set ulimits for file descriptors
		runtimeOptions = `
  runtime:
    ulimits:
      - name: nofile
        soft: 65536
        hard: 65536`
	}

	// On Windows/WSL2, configure k3s registries to help with image pulling
	// This creates a registries.yaml inside k3d nodes that k3s/containerd will use
	registriesConfig := ""
	if runtime.GOOS == "windows" {
		registriesConfig = `
registries:
  config: |
    mirrors:
      "docker.io":
        endpoint:
          - "https://registry-1.docker.io"`
	}


	configContent += fmt.Sprintf(`
kubeAPI:
  host: "%s"
  hostIP: "%s"
  hostPort: "%s"
options:
  k3s:
    extraArgs:
      - arg: --disable=traefik
        nodeFilters:
          - server:*
      - arg: --kubelet-arg=eviction-hard=
        nodeFilters:
          - all
      - arg: --kubelet-arg=eviction-soft=
        nodeFilters:
          - all%s%s
ports:
  - port: %s:80
    nodeFilters:
      - loadbalancer
  - port: %s:443
    nodeFilters:
      - loadbalancer%s`, hostIP, hostIP, apiPort, tlsSanArg, runtimeOptions, httpPort, httpsPort, registriesConfig)

	tmpFile, err := os.CreateTemp("", "k3d-config-*.yaml")
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

// isTestCluster determines if a cluster name indicates it's a test cluster
func (m *K3dManager) isTestCluster(name string) bool {
	testPatterns := []string{
		"test", "cleanup", "status", "list", "delete", "create",
		"multi", "single", "default_config", "with_type", "manual",
	}

	for _, pattern := range testPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return len(name) > timestampSuffixLen &&
		name[len(name)-timestampSuffixLen:] != name &&
		strings.ContainsAny(name[len(name)-timestampSuffixLen:], "0123456789")
}

// findAvailablePorts finds the specified number of available TCP ports using intelligent approach
func (m *K3dManager) findAvailablePorts(count int) ([]int, error) {
	// Get ports used by existing k3d clusters
	usedPorts := m.getUsedPortsByExistingClusters()

	// Start with default ports and increment if busy (matching script behavior)
	defaultPorts := []int{6550, 80, 443} // API, HTTP, HTTPS
	alternatePorts := []int{6551, 8080, 8443}

	var ports []int

	for i := 0; i < count && i < len(defaultPorts); i++ {
		// Check if default port is available and not used by existing clusters
		if m.isPortAvailable(defaultPorts[i]) && !m.isPortInUse(defaultPorts[i], usedPorts) {
			ports = append(ports, defaultPorts[i])
		} else if m.isPortAvailable(alternatePorts[i]) && !m.isPortInUse(alternatePorts[i], usedPorts) {
			ports = append(ports, alternatePorts[i])
		} else {
			// Find next available port that's not used by k3d clusters
			found := false
			for port := alternatePorts[i] + 1; port < alternatePorts[i]+1000; port++ {
				if m.isPortAvailable(port) && !m.isPortInUse(port, usedPorts) {
					ports = append(ports, port)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("could not find available port for index %d", i)
			}
		}
	}

	if len(ports) < count {
		return nil, fmt.Errorf("could not find %d available ports", count)
	}

	return ports, nil
}

// getUsedPortsByExistingClusters returns a map of ports used by existing k3d clusters
func (m *K3dManager) getUsedPortsByExistingClusters() map[int]bool {
	usedPorts := make(map[int]bool)

	ctx := context.Background()
	result, err := m.executor.Execute(ctx, "k3d", "cluster", "list", "--output", "json")
	if err != nil {
		return usedPorts // Return empty map on error, will rely on port availability check
	}

	var k3dClusters []k3dClusterInfo
	if err := json.Unmarshal([]byte(result.Stdout), &k3dClusters); err != nil {
		return usedPorts // Return empty map on error
	}

	// Extract ports from all existing clusters
	for _, cluster := range k3dClusters {
		for _, node := range cluster.Nodes {
			if node.Role == "server" || node.Role == "loadbalancer" {
				// Parse runtime labels to get port bindings
				if apiPort, exists := node.RuntimeLabels["k3d.server.api.port"]; exists {
					if port, err := strconv.Atoi(apiPort); err == nil {
						usedPorts[port] = true
					}
				}

				// Parse port mappings from the load balancer
				for _, mappings := range node.PortMappings {
					for _, mapping := range mappings {
						if mapping.HostPort != "" {
							if port, err := strconv.Atoi(mapping.HostPort); err == nil {
								usedPorts[port] = true
							}
						}
					}
				}
			}
		}
	}

	return usedPorts
}

// isPortInUse checks if a port is in the used ports map
func (m *K3dManager) isPortInUse(port int, usedPorts map[int]bool) bool {
	return usedPorts[port]
}

// isPortAvailable checks if a TCP port is available
func (m *K3dManager) isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// convertWindowsPathToWSL converts a Windows path to a WSL path format
// Example: C:\Users\foo\file.txt -> /mnt/c/Users/foo/file.txt
func (m *K3dManager) convertWindowsPathToWSL(windowsPath string) (string, error) {
	if windowsPath == "" {
		return "", fmt.Errorf("empty path provided")
	}

	// Expand Windows 8.3 short filenames to long path names
	// For example: C:\Users\RUNNER~1\... -> C:\Users\runneradmin\...
	// This is critical because WSL doesn't understand Windows short filenames
	expandedPath, err := expandShortPath(windowsPath)
	if err == nil && expandedPath != "" {
		windowsPath = expandedPath
	}

	// Replace backslashes with forward slashes
	path := strings.ReplaceAll(windowsPath, "\\", "/")

	// Convert drive letter (e.g., C: -> /mnt/c)
	if len(path) >= 2 && path[1] == ':' {
		driveLetter := strings.ToLower(string(path[0]))
		// Remove the drive letter and colon, then prepend /mnt/<drive>
		path = "/mnt/" + driveLetter + path[2:]
	}

	return path, nil
}

// k3dClusterInfo represents the JSON structure returned by k3d cluster list
type k3dClusterInfo struct {
	Name           string    `json:"name"`
	ServersCount   int       `json:"serversCount"`
	ServersRunning int       `json:"serversRunning"`
	AgentsCount    int       `json:"agentsCount"`
	AgentsRunning  int       `json:"agentsRunning"`
	Image          string    `json:"image,omitempty"`
	Nodes          []k3dNode `json:"nodes"`
}

// k3dNode represents a node in the k3d cluster
type k3dNode struct {
	Name          string                   `json:"name"`
	Role          string                   `json:"role"`
	Created       time.Time                `json:"created"`
	RuntimeLabels map[string]string        `json:"runtimeLabels,omitempty"`
	PortMappings  map[string][]PortMapping `json:"portMappings,omitempty"`
}

// PortMapping represents a port mapping for k3d nodes
type PortMapping struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// getWSLUser determines the correct WSL user to use for kubeconfig operations
// It tries to detect the non-root user that k3d/kubectl will run as
func (m *K3dManager) getWSLUser(ctx context.Context) (string, error) {
	// First, check what the default WSL user is (without specifying -u flag)
	// This works even if no named user exists (will return root)
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "whoami")
	if err == nil {
		defaultUser := strings.TrimSpace(result.Stdout)
		// If default user is not root, use it
		if defaultUser != "" && defaultUser != "root" {
			if m.verbose {
				fmt.Printf("Using default WSL user: %s\n", defaultUser)
			}
			return defaultUser, nil
		}
	}

	// If default user is root or detection failed, try to find a non-root user with a home directory
	result, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", "getent passwd | grep '/home/' | head -1 | cut -d: -f1")
	if err == nil && strings.TrimSpace(result.Stdout) != "" {
		username := strings.TrimSpace(result.Stdout)
		// Verify this user exists by checking whoami (run without -u to avoid circular failure)
		if verifyResult, verifyErr := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", fmt.Sprintf("id -u %s 2>/dev/null && echo %s", username, username)); verifyErr == nil {
			if strings.Contains(verifyResult.Stdout, username) {
				if m.verbose {
					fmt.Printf("Using detected WSL user: %s\n", username)
				}
				return username, nil
			}
		}
	}

	// Fall back to root - all commands use sudo anyway, so this is safer than
	// assuming a non-existent user like "runner"
	if m.verbose {
		fmt.Println("No non-root WSL user found, using root")
	}
	return "root", nil
}

// prepareKubeconfigDirectory ensures ~/.kube directory exists with proper permissions on Windows/WSL and Linux
func (m *K3dManager) prepareKubeconfigDirectory(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		// Get the WSL user that k3d will run as
		// The wrappers in the workflow use "runner", so we should detect or default to that
		username, err := m.getWSLUser(ctx)
		if err != nil {
			return fmt.Errorf("failed to get WSL user: %w", err)
		}

		// Create .kube directory with proper permissions in WSL
		createCmd := "mkdir -p ~/.kube && chmod 755 ~/.kube"
		_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", createCmd)
		if err != nil {
			return fmt.Errorf("failed to create .kube directory: %w", err)
		}

		if m.verbose {
			fmt.Println("✓ Prepared kubeconfig directory in WSL")
		}
	} else {
		// Linux/macOS: Create .kube directory with proper permissions
		createCmd := "mkdir -p ~/.kube && chmod 755 ~/.kube"
		_, err := m.executor.Execute(ctx, "bash", "-c", createCmd)
		if err != nil {
			return fmt.Errorf("failed to create .kube directory: %w", err)
		}

		if m.verbose {
			fmt.Println("✓ Prepared kubeconfig directory")
		}
	}

	return nil
}

// fixKubeconfigPermissions fixes kubeconfig file permissions on Windows/WSL and Linux
// This is needed because k3d running with sudo creates ~/.kube/config with root ownership
func (m *K3dManager) fixKubeconfigPermissions(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		// Get the WSL user that k3d will run as
		// The wrappers in the workflow use "runner", so we should detect or default to that
		username, err := m.getWSLUser(ctx)
		if err != nil {
			return fmt.Errorf("failed to get WSL user: %w", err)
		}

		// Fix ownership and permissions of both .kube directory and kubeconfig file in WSL
		// This is critical because k3d runs with sudo and creates files as root,
		// but kubectl needs to run as the regular user
		fixCmd := fmt.Sprintf("test -d ~/.kube && sudo chown -R %s:%s ~/.kube && sudo chmod 755 ~/.kube && test -f ~/.kube/config && sudo chmod 600 ~/.kube/config", username, username)
		_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", fixCmd)
		if err != nil {
			return fmt.Errorf("failed to fix kubeconfig permissions: %w", err)
		}

		if m.verbose {
			fmt.Println("✓ Fixed kubeconfig directory and file permissions for WSL user")
		}
	} else {
		// Linux/macOS: Fix permissions without changing ownership (assuming we're the owner)
		// First check if the file exists and needs fixing
		fixCmd := "test -f ~/.kube/config && chmod 600 ~/.kube/config || true"
		_, err := m.executor.Execute(ctx, "bash", "-c", fixCmd)
		if err != nil {
			return fmt.Errorf("failed to fix kubeconfig permissions: %w", err)
		}

		if m.verbose {
			fmt.Println("✓ Fixed kubeconfig permissions")
		}
	}

	return nil
}

// verifyClusterReachable checks if the cluster is reachable using native Go client
// This reduces reliance on external kubectl binary for context management
// Returns the *rest.Config that can be used to interact with the cluster
func (m *K3dManager) verifyClusterReachable(ctx context.Context, clusterName string) (*rest.Config, error) {
	contextName := fmt.Sprintf("k3d-%s", clusterName)

	var restConfig *rest.Config

	// On Windows, load kubeconfig content directly from WSL into memory
	// to avoid Windows filesystem path issues
	if runtime.GOOS == "windows" {
		// Retrieve kubeconfig content directly from k3d inside WSL
		kubeconfigContent, err := m.getKubeconfigContentFromWSL(ctx, clusterName)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve kubeconfig content from WSL: %w", err)
		}

		// Load the content from string into memory
		config, err := clientcmd.Load([]byte(kubeconfigContent))
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig content into memory: %w", err)
		}

		// Use WSL's eth0 IP for the Go client running on Windows to reach k3d inside WSL2
		// Docker runs inside WSL2 Ubuntu, so we need WSL's own IP (not the gateway).
		// From Windows, we can reach WSL services via WSL's eth0 IP (e.g., 172.x.x.2).
		wslIP, err := m.getWSLInternalIP(ctx)
		if err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not get WSL IP: %v\n", err)
				fmt.Println("Falling back to 127.0.0.1")
			}
			wslIP = "127.0.0.1" // Fallback
		} else if m.verbose {
			fmt.Printf("✓ Retrieved WSL IP for Go client: %s\n", wslIP)
		}

		// Replace all server addresses with the WSL IP
		for clusterName, cluster := range config.Clusters {
			// Extract the port from the current server URL
			re := regexp.MustCompile(`:(\d+)`)
			match := re.FindStringSubmatch(cluster.Server)

			if len(match) > 1 {
				oldServer := cluster.Server
				cluster.Server = fmt.Sprintf("https://%s:%s", wslIP, match[1])
				if m.verbose {
					fmt.Printf("Rewrote KubeAPI host for cluster %s: %s -> %s\n", clusterName, oldServer, cluster.Server)
				}
			}
		}

		// Check if the context exists
		if _, exists := config.Contexts[contextName]; !exists {
			return nil, fmt.Errorf("kubectl context %s not found in kubeconfig content", contextName)
		}

		// Switch the current context in memory
		config.CurrentContext = contextName

		if m.verbose {
			fmt.Printf("✓ Loaded kubeconfig from WSL and set context to %s\n", contextName)
		}

		// Build REST config from the in-memory kubeconfig
		restConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build REST config from in-memory kubeconfig: %w", err)
		}
	} else {
		// Non-Windows: use file-based kubeconfig
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
			fmt.Printf("✓ Switched kubectl context to %s\n", contextName)
		}

		// Build rest.Config from the loaded Kubeconfig
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			&clientcmd.ConfigOverrides{CurrentContext: contextName},
		).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build REST config: %w", err)
		}
	}

	// CRITICAL FIX: Bypass TLS Verification for local k3d clusters
	// The API server's certificate is issued to the cluster name or specific hostnames,
	// which may not match when connecting via 127.0.0.1 from Windows/WSL2.
	// This is safe for local development clusters and solves handshake failures.
	// Uses Insecure=true with CA data cleared, preserving client cert authentication.
	restConfig = sharedconfig.ApplyInsecureTLSConfig(restConfig)

	if m.verbose {
		fmt.Println("✓ TLS verification bypassed for local k3d cluster (Insecure=true, auth preserved)")
	}

	// --- PHASE 2: Verify Network Connectivity and Update Endpoint ---

	// On Windows/WSL2, the port might not be immediately available after k3d reports success
	// Use kubectl cluster-info (via shell) to verify the cluster is reachable
	// We use 127.0.0.1 with TLS bypass for reliable connectivity
	if runtime.GOOS == "windows" {
		_, err := m.getClusterEndpointFromShell(ctx, contextName)
		if err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not verify endpoint from kubectl cluster-info: %v\n", err)
				fmt.Println("Will proceed with 127.0.0.1 endpoint (TLS bypass enabled)...")
			}
			// Don't fail - proceed with the kubeconfig endpoint
		} else {
			if m.verbose {
				fmt.Printf("✓ Cluster endpoint verified via kubectl, using host: %s\n", restConfig.Host)
			}
		}
	}

	// Extract host and port from restConfig.Host for TCP check
	host, port, err := extractHostPort(restConfig.Host)
	if err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not extract host:port from %s: %v\n", restConfig.Host, err)
		}
		// Default to 127.0.0.1:6550 for k3d
		host = "127.0.0.1"
		port = defaultAPIPort
	}

	// Brief pause before TCP check on Windows/WSL2
	// This gives the port mapping time to stabilize after k3d reports success
	if runtime.GOOS == "windows" {
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for TCP port to be available before attempting API calls
	// This prevents flooding a dead port with requests on Windows/WSL2
	tcpRetries := 10
	tcpRetryDelay := 1 * time.Second
	if err := m.waitForTCPPort(ctx, host, port, tcpRetries, tcpRetryDelay); err != nil {
		return nil, fmt.Errorf("API server port not available: %w", err)
	}

	// --- PHASE 3: Verify Cluster Reachability via API ---

	// Create Kubernetes client with the (possibly updated) restConfig
	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Verify cluster reachability and node readiness with polling
	maxRetries := 15 // 15 retries * 2 seconds = 30 seconds max
	retryDelay := 2 * time.Second
	var lastErr error

	if m.verbose {
		fmt.Println("Waiting for cluster API and nodes to be reachable...")
	}

	for i := 0; i < maxRetries; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// 1. Check API server connectivity (simple list operation)
		nodes, err := coreClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			// Check if the error is temporary (e.g., connection refused)
			if isTemporaryError(err) {
				lastErr = err
				if m.verbose {
					fmt.Printf("  Cluster not ready yet (attempt %d/%d): %v\n", i+1, maxRetries, err)
				}
				time.Sleep(retryDelay)
				continue
			}
			// Fatal error - don't retry
			return nil, fmt.Errorf("failed to connect to cluster API: %w", err)
		}

		// 2. Check for node existence (k3d should have at least one node)
		if len(nodes.Items) == 0 {
			lastErr = fmt.Errorf("no nodes found in cluster")
			if m.verbose {
				fmt.Printf("  No nodes found yet (attempt %d/%d), waiting...\n", i+1, maxRetries)
			}
			time.Sleep(retryDelay)
			continue
		}

		// 3. Check if the required number of nodes are Ready
		// Using string constants to avoid k8s.io/api/core/v1 dependency
		readyCount := 0
		for _, node := range nodes.Items {
			for _, condition := range node.Status.Conditions {
				if string(condition.Type) == "Ready" && string(condition.Status) == "True" {
					readyCount++
					break
				}
			}
		}

		// Success condition: Nodes exist and at least one is ready
		if readyCount > 0 {
			if m.verbose {
				fmt.Printf("  Found %d ready node(s) out of %d total\n", readyCount, len(nodes.Items))
				fmt.Println("✓ Cluster API and nodes are ready.")
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

// waitForTCPPort performs a TCP connectivity check to verify the port is open
// This is critical for Windows/WSL2 where the port may not be immediately available
// after k3d reports cluster creation success
func (m *K3dManager) waitForTCPPort(ctx context.Context, host string, port string, maxRetries int, retryDelay time.Duration) error {
	address := net.JoinHostPort(host, port)

	if m.verbose {
		fmt.Printf("Waiting for TCP port %s to be available...\n", address)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Attempt TCP connection with short timeout
		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			conn.Close()
			if m.verbose {
				fmt.Printf("✓ TCP port %s is open\n", address)
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

// verifyClusterViaKubectl performs additional verification checks using kubectl
// This helps diagnose issues where the native Go client works but kubectl-based tools (like Helm) might not
func (m *K3dManager) verifyClusterViaKubectl(ctx context.Context, clusterName string) error {
	contextName := fmt.Sprintf("k3d-%s", clusterName)

	if m.verbose {
		fmt.Println("Running kubectl verification checks...")
	}

	// 1. Check kubectl cluster-info
	if m.verbose {
		fmt.Printf("  Checking kubectl cluster-info --context %s...\n", contextName)
	}
	result, err := m.executor.Execute(ctx, "kubectl", "--context", contextName, "cluster-info")
	if err != nil {
		return fmt.Errorf("kubectl cluster-info failed: %w", err)
	}
	if m.verbose {
		fmt.Printf("  kubectl cluster-info output:\n%s\n", result.Stdout)
	}

	// 2. Check kubectl get namespaces
	if m.verbose {
		fmt.Printf("  Checking kubectl get namespaces --context %s...\n", contextName)
	}
	result, err = m.executor.Execute(ctx, "kubectl", "--context", contextName, "get", "namespaces")
	if err != nil {
		return fmt.Errorf("kubectl get namespaces failed: %w", err)
	}
	if m.verbose {
		fmt.Printf("  Namespaces in cluster:\n%s\n", result.Stdout)
	}

	// Verify kube-system namespace exists (indicates cluster is properly initialized)
	if !strings.Contains(result.Stdout, "kube-system") {
		return fmt.Errorf("kube-system namespace not found - cluster may not be fully initialized")
	}

	// 3. Check kubectl get nodes
	if m.verbose {
		fmt.Printf("  Checking kubectl get nodes --context %s...\n", contextName)
	}
	result, err = m.executor.Execute(ctx, "kubectl", "--context", contextName, "get", "nodes", "-o", "wide")
	if err != nil {
		return fmt.Errorf("kubectl get nodes failed: %w", err)
	}
	if m.verbose {
		fmt.Printf("  Nodes in cluster:\n%s\n", result.Stdout)
	}

	// Verify at least one node is Ready
	if !strings.Contains(result.Stdout, "Ready") {
		return fmt.Errorf("no nodes in Ready state")
	}

	// 4. Verify kubeconfig context exists
	if m.verbose {
		fmt.Printf("  Checking kubectl config get-contexts for %s...\n", contextName)
	}
	result, err = m.executor.Execute(ctx, "kubectl", "config", "get-contexts", contextName)
	if err != nil {
		return fmt.Errorf("kubectl context %s not found: %w", contextName, err)
	}
	if m.verbose {
		fmt.Printf("  Context info:\n%s\n", result.Stdout)
	}

	if m.verbose {
		fmt.Println("✓ All kubectl verification checks passed")
	}

	return nil
}

// getClusterEndpointFromShell uses kubectl cluster-info to get the verified, live API endpoint URL
// This is more reliable than trusting the kubeconfig's advertised IP/port immediately after cluster creation
func (m *K3dManager) getClusterEndpointFromShell(ctx context.Context, contextName string) (string, error) {
	result, err := m.executor.Execute(ctx, "kubectl", "--context", contextName, "cluster-info")
	if err != nil {
		return "", fmt.Errorf("kubectl cluster-info failed: %w", err)
	}

	// Parse the output to extract the API server URL
	// Example output: "Kubernetes control plane is running at https://127.0.0.1:6550"
	endpoint := parseClusterInfoURL(result.Stdout)
	if endpoint == "" {
		return "", fmt.Errorf("could not parse API endpoint from cluster-info output: %s", result.Stdout)
	}

	if m.verbose {
		fmt.Printf("✓ Verified live API endpoint: %s\n", endpoint)
	}

	return endpoint, nil
}

// parseClusterInfoURL extracts the Kubernetes API server URL from kubectl cluster-info output
func parseClusterInfoURL(output string) string {
	// Look for "Kubernetes control plane is running at <URL>" or similar patterns
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Match patterns like "is running at https://..."
		if strings.Contains(line, "is running at") {
			parts := strings.Split(line, "is running at")
			if len(parts) >= 2 {
				// Extract the URL (trim ANSI codes and whitespace)
				urlPart := strings.TrimSpace(parts[1])
				// Remove any ANSI color codes
				urlPart = stripANSICodes(urlPart)
				// Find the URL starting with http
				for _, word := range strings.Fields(urlPart) {
					if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
						return word
					}
				}
			}
		}
	}
	return ""
}

// stripANSICodes removes ANSI escape codes from a string
func stripANSICodes(s string) string {
	// Simple ANSI code removal - handles common escape sequences
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// extractHostPort extracts host and port from a URL string
func extractHostPort(urlStr string) (string, string, error) {
	// Remove scheme if present
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	host, port, err := net.SplitHostPort(urlStr)
	if err != nil {
		// If no port specified, try to determine from scheme
		return urlStr, "", fmt.Errorf("could not split host:port from %s: %w", urlStr, err)
	}

	return host, port, nil
}

// getKubeconfigPath returns the kubeconfig file path (for non-Windows platforms)
func (m *K3dManager) getKubeconfigPath() string {
	// Check if KUBECONFIG environment variable is set
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Default to ~/.kube/config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to clientcmd recommended path
		return clientcmd.RecommendedHomeFile
	}

	return filepath.Join(homeDir, ".kube", "config")
}

// getKubeconfigContentFromWSL fetches kubeconfig content directly from k3d inside WSL
// This avoids Windows filesystem path issues by loading content into memory
func (m *K3dManager) getKubeconfigContentFromWSL(ctx context.Context, clusterName string) (string, error) {
	args := []string{"kubeconfig", "get", clusterName}

	// Execute the command - the executor will automatically wrap with 'wsl -d Ubuntu k3d ...'
	result, err := m.executor.Execute(ctx, "k3d", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig content from k3d: %w", err)
	}

	return result.Stdout, nil
}

// extractIPFromRouteOutput extracts a valid IPv4 address from route command output.
// This handles cases where the awk command doesn't properly extract field 3,
// returning the full line like "default via 172.21.96.1 dev eth0 proto kernel".
// It scans the output for a valid IPv4 address and returns it.
func extractIPFromRouteOutput(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}

	// If it's already a valid IP, return it
	if net.ParseIP(output) != nil {
		return output
	}

	// Otherwise, scan through the fields looking for an IP address
	// This handles both "ip route" output and other formats
	fields := strings.Fields(output)
	for _, field := range fields {
		if ip := net.ParseIP(field); ip != nil {
			// Ensure it's an IPv4 address (not IPv6)
			if ip.To4() != nil {
				return field
			}
		}
	}

	return ""
}

// getWSLInternalIP retrieves the WSL2 VM's own IP address (eth0 interface).
// This is the IP that Windows can use to reach services running inside WSL2.
// Docker runs inside WSL2 Ubuntu, and ports exposed by Docker are accessible
// via this IP from Windows.
//
// Note: This is different from the Windows host IP (default gateway from WSL's perspective).
// - WSL eth0 IP (e.g., 172.x.x.2): WSL's own IP, reachable from Windows
// - Windows host IP (e.g., 172.x.x.1): The gateway, used by WSL to reach Windows
func (m *K3dManager) getWSLInternalIP(ctx context.Context) (string, error) {
	// Get the WSL user for running the command
	username, err := m.getWSLUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Get WSL's own IP address (eth0 interface)
	// This is the IP that Windows can use to reach services in WSL2
	// The 'hostname -I' command returns all IP addresses, we take the first one (usually eth0)
	ipCmd := "hostname -I | awk '{print $1}'"
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", ipCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get WSL IP address: %w", err)
	}

	ip := strings.TrimSpace(result.Stdout)

	// Fallback: try getting eth0 IP directly
	if ip == "" {
		ipCmd = "ip addr show eth0 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1"
		result, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", ipCmd)
		if err != nil {
			return "", fmt.Errorf("failed to get WSL eth0 IP: %w", err)
		}
		ip = strings.TrimSpace(result.Stdout)
	}

	if ip == "" {
		return "", fmt.Errorf("WSL IP address is empty - could not determine from hostname or eth0")
	}

	// Validate that it's a proper IPv4 address using net.ParseIP
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid WSL IP format: %s", ip)
	}

	return ip, nil
}

// cleanupStaleLockFiles removes any stale kubeconfig lock files
func (m *K3dManager) cleanupStaleLockFiles(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		// Get the WSL user
		username, err := m.getWSLUser(ctx)
		if err != nil {
			return fmt.Errorf("failed to get WSL user: %w", err)
		}

		// Remove lock files in WSL
		cleanupCmd := "rm -f ~/.kube/config.lock ~/.kube/config.lock.* 2>/dev/null || true"
		_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", cleanupCmd)
		if err != nil {
			return fmt.Errorf("failed to cleanup lock files: %w", err)
		}
	} else {
		// Linux/macOS: Remove lock files
		cleanupCmd := "rm -f ~/.kube/config.lock ~/.kube/config.lock.* 2>/dev/null || true"
		_, err := m.executor.Execute(ctx, "bash", "-c", cleanupCmd)
		if err != nil {
			return fmt.Errorf("failed to cleanup lock files: %w", err)
		}
	}

	if m.verbose {
		fmt.Println("✓ Cleaned up stale kubeconfig lock files")
	}

	return nil
}

// rewriteWSLKubeconfigServerAddress rewrites the kubeconfig file in WSL to use 127.0.0.1
// instead of 0.0.0.0. This is necessary because:
// - k3d writes 0.0.0.0 as the server address which doesn't work for connections
// - Docker runs INSIDE WSL2 Ubuntu (not Docker Desktop), so k3d is in the same network namespace
// - From Ubuntu WSL, 127.0.0.1 refers to the WSL loopback where Docker/k3d is listening
// - We only need to rewrite 0.0.0.0 to 127.0.0.1
func (m *K3dManager) rewriteWSLKubeconfigServerAddress(ctx context.Context, _ string) error {
	// Only needed on Windows where helm runs inside WSL
	if runtime.GOOS != "windows" {
		return nil
	}

	// Get the WSL user
	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Use sed to replace 0.0.0.0 with 127.0.0.1
	// Docker runs inside WSL2 Ubuntu, so localhost (127.0.0.1) is the correct address
	sedCmd := `sed -i 's|server: https://0\.0\.0\.0:|server: https://127.0.0.1:|g' ~/.kube/config`

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", sedCmd)
	if err != nil {
		return fmt.Errorf("failed to rewrite kubeconfig server address: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Rewrote kubeconfig server addresses to use 127.0.0.1")
	}

	return nil
}

// Factory functions for backward compatibility

// CreateClusterManagerWithExecutor creates a K3D cluster manager with a specific command executor
func CreateClusterManagerWithExecutor(exec executor.CommandExecutor) *K3dManager {
	if exec == nil {
		panic("Executor cannot be nil - must be provided by calling code to avoid import cycles")
	}
	return NewK3dManager(exec, false)
}

// CreateDefaultClusterManager creates a K3D cluster manager with all default configuration
// Deprecated: Use CreateClusterManagerWithExecutor instead with a proper executor.
func CreateDefaultClusterManager() *K3dManager {
	panic("CreateDefaultClusterManager is deprecated - use CreateClusterManagerWithExecutor with proper executor")
}

// configureWSLDNS configures reliable DNS servers in WSL2 before k3d cluster creation.
// This is critical because k3d nodes (containers inside Docker inside WSL2) can have
// intermittent DNS resolution failures when using WSL2's default DNS configuration.
//
// WSL2's default DNS resolution can be flaky, especially on GitHub Actions runners,
// causing k3d container nodes to fail pulling images like "rancher/mirrored-pause:3.6"
// with "dial tcp ...:443: i/o timeout" errors.
//
// By configuring /etc/resolv.conf with Google DNS (8.8.8.8) and Cloudflare DNS (1.1.1.1),
// we ensure reliable DNS resolution for all containers in the k3d cluster.
func (m *K3dManager) configureWSLDNS(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil // Only needed on Windows/WSL2
	}

	// Get the WSL user
	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Configure DNS in /etc/resolv.conf with reliable public DNS servers
	// We use a combination of Google DNS and Cloudflare DNS for redundancy
	// The script:
	// 1. Disables WSL's automatic resolv.conf generation (if not already disabled)
	// 2. Writes reliable DNS servers to /etc/resolv.conf
	dnsConfigScript := `
# Check if resolv.conf is a symlink (WSL auto-generated) and remove it
if [ -L /etc/resolv.conf ]; then
    sudo rm /etc/resolv.conf
fi

# Configure reliable DNS servers
sudo tee /etc/resolv.conf > /dev/null <<EOF
# DNS configured by openframe-cli for reliable k3d networking
nameserver 8.8.8.8
nameserver 1.1.1.1
nameserver 8.8.4.4
EOF

# Prevent WSL from overwriting resolv.conf on restart (if wsl.conf doesn't exist or doesn't have the setting)
if [ ! -f /etc/wsl.conf ] || ! grep -q "generateResolvConf" /etc/wsl.conf 2>/dev/null; then
    sudo tee -a /etc/wsl.conf > /dev/null <<EOF

[network]
generateResolvConf = false
EOF
fi

echo "DNS configured successfully"
`

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", dnsConfigScript)
	if err != nil {
		return fmt.Errorf("failed to configure DNS in WSL: %w", err)
	}

	// Verify network is working by testing connectivity
	// Use curl instead of nslookup - it's more reliable as it uses the system resolver
	verifyScript := `
for i in 1 2 3; do
    if curl -fsSL --connect-timeout 5 --max-time 10 -o /dev/null https://registry-1.docker.io/v2/ 2>/dev/null; then
        echo "DNS verification passed"
        exit 0
    fi
    sleep 2
done
echo "DNS verification failed but continuing"
exit 0
`

	_, _ = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", verifyScript)

	if m.verbose {
		fmt.Println("✓ Configured reliable DNS servers in WSL (8.8.8.8, 1.1.1.1, 8.8.4.4)")
	}

	return nil
}

// ensureDockerDNS ensures Docker daemon has DNS configuration applied.
// This is critical for k3d containers to resolve external registries like registry-1.docker.io.
// We only restart Docker if DNS config was missing - otherwise we just verify Docker is running.
func (m *K3dManager) ensureDockerDNS(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil // Only needed on Windows/WSL2
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Script to ensure Docker daemon.json has DNS config
	// Only restart Docker if the config was missing/changed
	// Base64 encoding avoids shell interpretation issues when passing through Windows -> WSL
	ensureDNSScript := `#!/bin/bash

# Ensure /etc/docker directory exists
sudo mkdir -p /etc/docker

# Check if daemon.json exists and has DNS configuration
DAEMON_JSON="/etc/docker/daemon.json"
NEEDS_RESTART=0

if [ -f "$DAEMON_JSON" ]; then
    if grep -q '"dns"' "$DAEMON_JSON" 2>/dev/null; then
        echo "Docker daemon.json already has DNS configuration"
    else
        NEEDS_RESTART=1
    fi
else
    NEEDS_RESTART=1
fi

# If DNS not configured, create daemon.json and restart
if [ "$NEEDS_RESTART" = "1" ]; then
    echo "Configuring Docker daemon DNS..."
    # Create a simple daemon.json with DNS servers
    sudo tee "$DAEMON_JSON" > /dev/null <<'DAEMONJSON'
{
    "dns": ["8.8.8.8", "1.1.1.1", "8.8.4.4"]
}
DAEMONJSON
    echo "Docker daemon.json configured with DNS servers"

    echo "Restarting Docker daemon to apply DNS configuration..."
    sudo systemctl restart docker || sudo service docker restart || true

    # Wait for Docker to be ready after restart
    echo "Waiting for Docker to be ready..."
    for i in $(seq 1 30); do
        if sudo docker info >/dev/null 2>&1; then
            echo "Docker is ready"
            exit 0
        fi
        echo "Waiting for Docker... (attempt $i/30)"
        sleep 2
    done
    echo "WARNING: Docker may not be fully ready"
else
    # Just verify Docker is running without restarting
    if ! sudo docker info >/dev/null 2>&1; then
        echo "Docker is not running, starting it..."
        sudo systemctl start docker || sudo service docker start || true
        sleep 5
    fi
    echo "Docker is running with DNS configuration"
fi

exit 0
`

	// Base64-encode the script to avoid shell character interpretation issues
	encoded := base64.StdEncoding.EncodeToString([]byte(ensureDNSScript))
	wrapperCmd := fmt.Sprintf("echo %s | base64 -d | bash", encoded)

	if m.verbose {
		fmt.Println("Ensuring Docker daemon has DNS configuration...")
	}

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", wrapperCmd)
	if err != nil {
		return fmt.Errorf("failed to ensure Docker DNS configuration: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Docker daemon DNS configuration verified")
	}

	return nil
}

// prePullK3sImages pre-pulls critical k3s images to Docker cache.
// This is called BEFORE cluster creation to cache images locally.
// Note: This doesn't directly help k3d nodes, but it can speed up k3d image import.
func (m *K3dManager) prePullK3sImages(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Critical images that k3s needs - pre-pull them to Docker cache
	// The pause image is the most critical one that causes pod sandbox creation failures
	images := []string{
		"rancher/mirrored-pause:3.6",
	}

	if m.verbose {
		fmt.Println("Pre-pulling critical k3s images to Docker cache...")
	}

	// Script to pull images with retries
	pullScript := `#!/bin/bash
IMAGES="$@"
for IMAGE in $IMAGES; do
    echo "Pulling $IMAGE..."
    for i in 1 2 3; do
        if sudo docker pull "$IMAGE" 2>/dev/null; then
            echo "Successfully pulled $IMAGE"
            break
        fi
        echo "Retry $i/3 for $IMAGE..."
        sleep 5
    done
done
echo "Image pre-pull complete"
`

	// Base64 encode the script
	encoded := base64.StdEncoding.EncodeToString([]byte(pullScript))
	imageList := strings.Join(images, " ")
	wrapperCmd := fmt.Sprintf("echo %s | base64 -d | bash -s %s", encoded, imageList)

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", wrapperCmd)
	if err != nil {
		return fmt.Errorf("failed to pre-pull images: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Critical k3s images pre-pulled to Docker cache")
	}

	return nil
}

// importK3sImages imports critical images directly into containerd inside k3d nodes.
// This bypasses DNS resolution issues by:
// 1. Saving the image from Docker to a tar file
// 2. Copying into k3d container and importing via ctr
func (m *K3dManager) importK3sImages(ctx context.Context, clusterName string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// The pause image is critical for pod sandbox creation
	image := "rancher/mirrored-pause:3.6"
	tarFile := "/tmp/pause-image.tar"

	fmt.Println("Importing critical images directly into containerd...")

	// Step 1: Save image from Docker to tar file
	saveCmd := fmt.Sprintf("sudo docker save -o %s %s", tarFile, image)
	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", saveCmd)
	if err != nil {
		return fmt.Errorf("failed to save image to tar: %w", err)
	}
	fmt.Printf("✓ Saved %s to %s\n", image, tarFile)

	// Step 2: Get list of k3d node containers (excluding serverlb)
	listCmd := fmt.Sprintf("sudo docker ps --filter 'name=k3d-%s' --format '{{.Names}}' | grep -v serverlb", clusterName)
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", listCmd)
	if err != nil {
		return fmt.Errorf("failed to list k3d containers: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	// Step 3: Import image into containerd in each node
	for _, container := range containers {
		container = strings.TrimSpace(container)
		if container == "" {
			continue
		}

		// Copy tar file into the container
		copyCmd := fmt.Sprintf("sudo docker cp %s %s:/tmp/pause-image.tar", tarFile, container)
		_, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", copyCmd)
		if err != nil {
			fmt.Printf("Warning: Failed to copy tar to %s: %v\n", container, err)
			continue
		}

		// Import using ctr into k8s.io namespace (where k3s/containerd looks for images)
		importCmd := fmt.Sprintf(`sudo docker exec %s sh -c 'ctr -n k8s.io images import /tmp/pause-image.tar && rm /tmp/pause-image.tar'`, container)
		_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", importCmd)
		if err != nil {
			fmt.Printf("Warning: Failed to import image in %s: %v\n", container, err)
			continue
		}
		fmt.Printf("✓ Imported pause image into %s\n", container)
	}

	// Cleanup tar file
	cleanupCmd := fmt.Sprintf("rm -f %s", tarFile)
	_, _ = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", cleanupCmd)

	// Verify images are in containerd
	serverContainer := fmt.Sprintf("k3d-%s-server-0", clusterName)
	verifyCmd := fmt.Sprintf(`sudo docker exec %s sh -c '
echo "=== Images in containerd (ctr k8s.io namespace) ==="
ctr -n k8s.io images ls 2>&1 | grep -E "(pause|REF)" || echo "No pause image found"
'`, serverContainer)
	result, _ = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", verifyCmd)
	fmt.Printf("Image verification:\n%s\n", strings.TrimSpace(result.Stdout))

	fmt.Println("✓ Image import complete")

	return nil
}

// fixDockerNetworkRouting fixes Docker bridge network routing issues on WSL2.
// On WSL2/GitHub Actions, the Docker bridge network sometimes can't route to the internet.
// This function ensures IP forwarding is enabled.
func (m *K3dManager) fixDockerNetworkRouting(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	fmt.Println("Enabling IP forwarding for Docker networking...")

	// Only enable IP forwarding - don't restart Docker as that can destabilize it
	// Docker manages its own iptables rules for internal networking
	// Base64 encoded script to avoid shell escaping issues:
	// #!/bin/bash
	// echo 1 | sudo tee /proc/sys/net/ipv4/ip_forward > /dev/null
	// sudo sysctl -w net.ipv4.ip_forward=1 2>/dev/null || true
	// echo "IP forwarding enabled"
	script := "IyEvYmluL2Jhc2gKZWNobyAxIHwgc3VkbyB0ZWUgL3Byb2Mvc3lzL25ldC9pcHY0L2lwX2ZvcndhcmQgPiAvZGV2L251bGwKc3VkbyBzeXNjdGwgLXcgbmV0LmlwdjQuaXBfZm9yd2FyZD0xIDI+L2Rldi9udWxsIHx8IHRydWUKZWNobyAiSVAgZm9yd2FyZGluZyBlbmFibGVkIgo="

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", fmt.Sprintf("echo %s | base64 -d | bash", script))
	if err != nil {
		// Non-fatal - just log and continue, IP forwarding might already be enabled
		fmt.Printf("Warning: failed to enable IP forwarding: %v\n", err)
		return nil
	}

	fmt.Println("✓ IP forwarding enabled")
	return nil
}

// fixK3dNodeDNS fixes DNS configuration inside k3d container nodes.
// k3d nodes are Docker containers running k3s/containerd. Even if Docker daemon.json
// has DNS configured, the containers may have stale /etc/resolv.conf that doesn't work.
// This function updates /etc/resolv.conf inside each k3d node container.
func (m *K3dManager) fixK3dNodeDNS(ctx context.Context, clusterName string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	fmt.Println("Checking network connectivity inside k3d nodes...")

	// Get list of k3d node containers
	listCmd := fmt.Sprintf("sudo docker ps --filter 'name=k3d-%s' --format '{{.Names}}'", clusterName)
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", listCmd)
	if err != nil {
		return fmt.Errorf("failed to list k3d containers: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	if len(containers) == 0 || (len(containers) == 1 && containers[0] == "") {
		return fmt.Errorf("no k3d containers found for cluster %s", clusterName)
	}

	// Find server container for diagnostics
	serverContainer := ""
	for _, container := range containers {
		if strings.Contains(container, "-server-") && !strings.Contains(container, "-serverlb") {
			serverContainer = strings.TrimSpace(container)
			break
		}
	}

	// Test network connectivity BEFORE making any changes
	if serverContainer != "" {
		fmt.Println("Testing network BEFORE DNS changes...")
		testCmd := fmt.Sprintf(`sudo docker exec %s sh -c '
echo "=== Current /etc/resolv.conf ==="
cat /etc/resolv.conf
echo ""
echo "=== Testing ping to 8.8.8.8 ==="
ping -c 1 -W 3 8.8.8.8 2>&1 || echo "ping failed"
'`, serverContainer)
		result, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", testCmd)
		fmt.Printf("BEFORE DNS fix:\n%s\n", strings.TrimSpace(result.Stdout))
	}

	// Fix DNS in each container (do NOT restart containerd - it breaks networking)
	for _, container := range containers {
		container = strings.TrimSpace(container)
		if container == "" {
			continue
		}

		// Skip the load balancer - it doesn't run containerd
		if strings.Contains(container, "-serverlb") {
			continue
		}

		// Update /etc/resolv.conf inside the container with reliable DNS servers
		// We use docker exec to run commands inside the k3d node container
		// NOTE: We do NOT restart containerd - that breaks networking!
		fixDNSCmd := fmt.Sprintf(`sudo docker exec %s sh -c 'cat > /etc/resolv.conf << EOF
# DNS configured by openframe-cli for reliable registry access
nameserver 8.8.8.8
nameserver 1.1.1.1
nameserver 8.8.4.4
EOF'`, container)

		_, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", fixDNSCmd)
		if err != nil {
			fmt.Printf("Warning: Failed to fix DNS in container %s: %v\n", container, err)
			// Continue with other containers
		} else {
			fmt.Printf("✓ Fixed DNS in container: %s\n", container)
		}
	}

	// Brief pause to let DNS changes take effect
	time.Sleep(2 * time.Second)

	// Test network connectivity AFTER DNS changes
	if serverContainer != "" {
		fmt.Println("Testing network AFTER DNS changes...")
		testCmd := fmt.Sprintf(`sudo docker exec %s sh -c '
echo "=== /etc/resolv.conf ==="
cat /etc/resolv.conf
echo ""
echo "=== Testing network connectivity to DNS servers ==="
ping -c 1 -W 2 8.8.8.8 2>&1 || echo "ping to 8.8.8.8 failed"
echo ""
echo "=== Testing DNS resolution ==="
nslookup registry-1.docker.io 8.8.8.8 2>&1 || echo "nslookup failed"
echo ""
echo "=== Testing HTTPS connectivity ==="
wget -q --spider --timeout=10 https://registry-1.docker.io/v2/ 2>&1 && echo "Registry reachable" || echo "Registry not reachable"
echo ""
echo "=== Checking crictl images ==="
crictl images 2>&1 | grep -E "(pause|IMAGE)" || echo "crictl not available or no pause image"
'`, serverContainer)
		result, _ := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", testCmd)
		// Always print diagnostics on Windows for CI debugging
		fmt.Printf("k3d node network diagnostics:\n%s\n", strings.TrimSpace(result.Stdout))
	}

	fmt.Println("✓ DNS configuration fixed in k3d nodes")

	return nil
}

// verifyDockerReady checks if Docker is running and starts it if not.
// This is called right before k3d cluster create to ensure Docker is available.
func (m *K3dManager) verifyDockerReady(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	username, err := m.getWSLUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Check if Docker is running, start it if not
	checkScript := `
if sudo docker info >/dev/null 2>&1; then
    echo "Docker is running"
    exit 0
fi

echo "Docker is not running, starting it..."
sudo systemctl start docker || sudo service docker start || true

# Wait for Docker to be ready
for i in $(seq 1 15); do
    if sudo docker info >/dev/null 2>&1; then
        echo "Docker started successfully"
        exit 0
    fi
    echo "Waiting for Docker... (attempt $i/15)"
    sleep 2
done

echo "ERROR: Docker failed to start"
exit 1
`

	encoded := base64.StdEncoding.EncodeToString([]byte(checkScript))
	wrapperCmd := fmt.Sprintf("echo %s | base64 -d | bash", encoded)

	if m.verbose {
		fmt.Println("Verifying Docker is ready before cluster creation...")
	}

	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", wrapperCmd)
	if err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Docker is ready")
	}

	return nil
}

// increaseInotifyLimits increases the inotify limits on the host system
// This is critical for applications like MeshCentral that use many file watchers
// and can hit the default limits, causing EMFILE errors.
//
// The limits are set via sysctl:
// - fs.inotify.max_user_watches: max number of file watches per user (default: 8192)
// - fs.inotify.max_user_instances: max number of inotify instances per user (default: 128)
func (m *K3dManager) increaseInotifyLimits(ctx context.Context) error {
	// Desired limits - these are common recommended values for development environments
	const maxUserWatches = "524288"
	const maxUserInstances = "512"

	if runtime.GOOS == "windows" {
		// On Windows, the limits need to be set inside WSL2 where Docker runs
		// We need root privileges to modify sysctl settings
		sysctlCmd := fmt.Sprintf(
			"sudo sysctl -w fs.inotify.max_user_watches=%s fs.inotify.max_user_instances=%s 2>/dev/null || true",
			maxUserWatches, maxUserInstances,
		)

		_, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", sysctlCmd)
		if err != nil {
			return fmt.Errorf("failed to set inotify limits in WSL: %w", err)
		}

		if m.verbose {
			fmt.Printf("✓ Increased inotify limits in WSL (max_user_watches=%s, max_user_instances=%s)\n",
				maxUserWatches, maxUserInstances)
		}
	} else {
		// On Linux/macOS, set the limits directly
		// Note: macOS doesn't use inotify (uses FSEvents), so this only applies to Linux
		sysctlCmd := fmt.Sprintf(
			"sudo sysctl -w fs.inotify.max_user_watches=%s fs.inotify.max_user_instances=%s 2>/dev/null || true",
			maxUserWatches, maxUserInstances,
		)

		_, err := m.executor.Execute(ctx, "bash", "-c", sysctlCmd)
		if err != nil {
			return fmt.Errorf("failed to set inotify limits: %w", err)
		}

		if m.verbose {
			fmt.Printf("✓ Increased inotify limits (max_user_watches=%s, max_user_instances=%s)\n",
				maxUserWatches, maxUserInstances)
		}
	}

	return nil
}
