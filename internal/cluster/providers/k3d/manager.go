package k3d

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"k8s.io/client-go/rest"
)

// Constants for configuration
const (
	defaultK3sImage    = "rancher/k3s:v1.31.5-k3s1"
	defaultTimeout     = "300s"
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
		if configContent, err := os.ReadFile(configFile); err == nil { // #nosec G304 -- reads a temp config file this process just created
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

	// Verify the cluster is reachable and get the rest.Config via the native
	// client (client-go). This is the sole verification — the previous best-effort
	// kubectl double-check was removed with the kubectl migration.
	restConfig, err := m.verifyClusterReachable(ctx, config.Name)
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, fmt.Errorf("cluster created but not reachable: %w", err))
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
				return models.NewClusterOperationError("delete", name, fmt.Errorf("failed to delete cluster %s (cleanup also failed: %w): %w", name, cleanupErr, err))
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
		username = "runner" // fallback to runner
	}

	// Select containers by the k3d.cluster label (exact match). A name= filter
	// is an unanchored regex: deleting cluster "dev" would also match the
	// containers of "dev-2", "dev-old", ... (T0-2).
	cleanupCmd := fmt.Sprintf(
		"sudo docker ps -aq --filter 'label=k3d.cluster=%s' | xargs -r sudo docker rm -f 2>/dev/null || true",
		clusterName,
	)
	_, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", cleanupCmd)
	if err != nil {
		return fmt.Errorf("failed to cleanup containers via WSL: %w", err)
	}

	// Also remove the network
	networkCleanupCmd := fmt.Sprintf("sudo docker network rm k3d-%s 2>/dev/null || true", clusterName)
	if _, nerr := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", networkCleanupCmd); nerr != nil && m.verbose {
		fmt.Printf("Warning: failed to remove k3d network for %s: %v\n", clusterName, nerr)
	}

	return nil
}

// forceCleanupDockerContainersDirect removes k3d containers directly (non-Windows)
func (m *K3dManager) forceCleanupDockerContainersDirect(ctx context.Context, clusterName string) error {
	// Select containers by the k3d.cluster label (exact match). A name= filter
	// is an unanchored regex: deleting cluster "dev" would also match the
	// containers of "dev-2", "dev-old", ... (T0-2).
	result, err := m.executor.Execute(ctx, "docker", "ps", "-aq", "--filter", fmt.Sprintf("label=k3d.cluster=%s", clusterName))
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.TrimSpace(result.Stdout)
	if containerIDs != "" {
		// Remove each container
		for _, id := range strings.Split(containerIDs, "\n") {
			id = strings.TrimSpace(id)
			if id != "" {
				if _, rerr := m.executor.Execute(ctx, "docker", "rm", "-f", id); rerr != nil && m.verbose {
					fmt.Printf("Warning: failed to remove container %s: %v\n", id, rerr)
				}
			}
		}
	}

	// Also remove the network
	if _, nerr := m.executor.Execute(ctx, "docker", "network", "rm", fmt.Sprintf("k3d-%s", clusterName)); nerr != nil && m.verbose {
		fmt.Printf("Warning: failed to remove k3d network for %s: %v\n", clusterName, nerr)
	}

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

	// Find available ports, preferring standard ports (80, 443) with fallback to high ports
	ports, err := m.findAvailablePorts()
	if err != nil {
		return "", fmt.Errorf("failed to find available ports: %w", err)
	}
	apiPort := strconv.Itoa(ports.API)
	httpPort := strconv.Itoa(ports.HTTP)
	httpsPort := strconv.Itoa(ports.HTTPS)

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
          - all%s
ports:
  - port: %s:80
    nodeFilters:
      - loadbalancer
  - port: %s:443
    nodeFilters:
      - loadbalancer`, hostIP, hostIP, apiPort, tlsSanArg, httpPort, httpsPort)

	tmpFile, err := os.CreateTemp("", "k3d-config-*.yaml")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// Factory functions for backward compatibility

// CreateClusterManagerWithExecutor creates a K3D cluster manager with a specific command executor
func CreateClusterManagerWithExecutor(exec executor.CommandExecutor) *K3dManager {
	if exec == nil {
		panic("Executor cannot be nil - must be provided by calling code to avoid import cycles")
	}
	return NewK3dManager(exec, false)
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
