package k3d

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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

// isWSLAvailable checks if WSL with Ubuntu is available on Windows
// Returns false on non-Windows platforms
func isWSLAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Try to run a simple WSL command to check availability
	cmd := exec.Command("wsl", "-d", "Ubuntu", "echo", "ok")
	err := cmd.Run()
	return err == nil
}

// wslAvailable caches the WSL availability check result
var wslAvailable = isWSLAvailable()

func init() {
	// Register the k3d manager factory
	manager.RegisterManager(manager.ManagerTypeK3d, func(exec executor.CommandExecutor, verbose bool) manager.ClusterManager {
		return NewK3dManager(exec, verbose)
	})
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

	configFile, err := m.createK3dConfigFile(config)
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

	// Convert Windows path to WSL path if running on Windows with WSL available
	configFilePath := configFile
	if runtime.GOOS == "windows" && wslAvailable {
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

	// Verify the cluster is reachable and get the rest.Config
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

	if _, err := m.executor.Execute(ctx, "k3d", args...); err != nil {
		return models.NewClusterOperationError("delete", name, fmt.Errorf("failed to delete cluster %s: %w", name, err))
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

	result, err := m.executor.Execute(ctx, "k3d", args...)
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
	if _, err := m.executor.Execute(ctx, "k3d", args...); err != nil {
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
func (m *K3dManager) createK3dConfigFile(config models.ClusterConfig) (string, error) {
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

	// On Windows/WSL2, bind to 0.0.0.0 so the API is accessible via the WSL internal IP
	// This is necessary because the connectivity check uses the WSL eth0 IP to bypass Windows NAT
	// On native Windows (no WSL), use 127.0.0.1
	hostIP := "127.0.0.1"
	if runtime.GOOS == "windows" && wslAvailable {
		hostIP = "0.0.0.0"
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
          - all
ports:
  - port: %s:80
    nodeFilters:
      - loadbalancer
  - port: %s:443
    nodeFilters:
      - loadbalancer`, hostIP, hostIP, apiPort, httpPort, httpsPort)

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
	// First, try to get the user specified for the runner user (standard in GitHub Actions)
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", "runner", "whoami")
	if err == nil && strings.TrimSpace(result.Stdout) == "runner" {
		return "runner", nil
	}

	// If runner doesn't exist, try to find the first non-root user with a home directory
	result, err = m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "bash", "-c", "getent passwd | grep '/home/' | head -1 | cut -d: -f1")
	if err == nil && strings.TrimSpace(result.Stdout) != "" {
		username := strings.TrimSpace(result.Stdout)
		// Verify this user exists and has a home directory
		if verifyResult, verifyErr := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "whoami"); verifyErr == nil {
			if strings.TrimSpace(verifyResult.Stdout) == username {
				return username, nil
			}
		}
	}

	// If we can't detect a proper user, default to "runner" (common in CI environments)
	// This is safer than using root, which causes permission issues
	return "runner", nil
}

// prepareKubeconfigDirectory ensures ~/.kube directory exists with proper permissions on Windows/WSL and Linux
func (m *K3dManager) prepareKubeconfigDirectory(ctx context.Context) error {
	if runtime.GOOS == "windows" && wslAvailable {
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
	} else if runtime.GOOS == "windows" {
		// Native Windows: Create .kube directory using PowerShell
		kubeDir := filepath.Join(os.Getenv("USERPROFILE"), ".kube")
		if err := os.MkdirAll(kubeDir, 0755); err != nil {
			return fmt.Errorf("failed to create .kube directory: %w", err)
		}

		if m.verbose {
			fmt.Println("✓ Prepared kubeconfig directory on Windows")
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
	if runtime.GOOS == "windows" && wslAvailable {
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
	} else if runtime.GOOS == "windows" {
		// Native Windows: No permission fixes needed, Windows handles permissions differently
		if m.verbose {
			fmt.Println("✓ Kubeconfig permissions (native Windows - no changes needed)")
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

	// On Windows with WSL, load kubeconfig content directly from WSL into memory
	// to avoid Windows filesystem path issues
	if runtime.GOOS == "windows" && wslAvailable {
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

		// CRITICAL FIX: Use WSL internal IP to bypass Windows NAT/Firewall
		// The Windows NAT layer blocks connections to 127.0.0.1 for ports exposed by WSL2/Docker.
		// Using the internal WSL IP (e.g., 172.x.x.x) bypasses this restriction entirely.
		wslInternalIP, err := m.getWSLInternalIP(ctx)
		if err != nil {
			if m.verbose {
				fmt.Printf("Warning: Could not get WSL internal IP: %v\n", err)
				fmt.Println("Falling back to 127.0.0.1 (may fail due to Windows NAT)")
			}
			wslInternalIP = "127.0.0.1" // Fallback, but likely won't work
		} else if m.verbose {
			fmt.Printf("✓ Retrieved WSL internal IP: %s\n", wslInternalIP)
		}

		// Replace all server addresses with the WSL internal IP
		for clusterName, cluster := range config.Clusters {
			// Extract the port from the current server URL
			re := regexp.MustCompile(`:(\d+)`)
			match := re.FindStringSubmatch(cluster.Server)

			if len(match) > 1 {
				oldServer := cluster.Server
				cluster.Server = fmt.Sprintf("https://%s:%s", wslInternalIP, match[1])
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
	if runtime.GOOS == "windows" && wslAvailable {
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
	if runtime.GOOS == "windows" && wslAvailable {
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

// getWSLInternalIP retrieves the internal IP address of the WSL2 eth0 interface.
// This IP (e.g., 172.x.x.x) bypasses Windows NAT/Firewall that blocks 127.0.0.1.
func (m *K3dManager) getWSLInternalIP(ctx context.Context) (string, error) {
	// Get the WSL user for running the command
	username, err := m.getWSLUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get WSL user: %w", err)
	}

	// Get the internal IP of eth0 in WSL
	// This uses the standard ip command to extract the IPv4 address
	ipCmd := "ip a show eth0 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1"
	result, err := m.executor.Execute(ctx, "wsl", "-d", "Ubuntu", "-u", username, "bash", "-c", ipCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get WSL internal IP: %w", err)
	}

	ip := strings.TrimSpace(result.Stdout)

	// CRITICAL FIX: Clean up potential parsing artifacts from shell command
	// The shell pipeline may return "inet 172.x.x.x" instead of just the IP
	// due to shell escaping issues when executed through WSL
	ip = strings.TrimPrefix(ip, "inet ")
	ip = strings.TrimSpace(ip)

	// Also handle potential CIDR notation if cut -d/ didn't work
	if idx := strings.Index(ip, "/"); idx != -1 {
		ip = ip[:idx]
	}

	if ip == "" {
		return "", fmt.Errorf("WSL internal IP is empty")
	}

	// Validate that the IP looks like an IPv4 address
	if !strings.Contains(ip, ".") {
		return "", fmt.Errorf("invalid WSL internal IP format: %s", ip)
	}

	return ip, nil
}

// cleanupStaleLockFiles removes any stale kubeconfig lock files
func (m *K3dManager) cleanupStaleLockFiles(ctx context.Context) error {
	if runtime.GOOS == "windows" && wslAvailable {
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
	} else if runtime.GOOS == "windows" {
		// Native Windows: Remove lock files using Go
		kubeDir := filepath.Join(os.Getenv("USERPROFILE"), ".kube")
		lockFiles, _ := filepath.Glob(filepath.Join(kubeDir, "config.lock*"))
		for _, f := range lockFiles {
			os.Remove(f)
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
