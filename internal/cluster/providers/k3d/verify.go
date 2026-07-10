package k3d

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

	// Extract host and port from restConfig.Host for TCP check
	host, port, err := extractHostPort(restConfig.Host)
	if err != nil {
		if m.verbose {
			fmt.Printf("Warning: Could not extract host:port from %s: %v\n", restConfig.Host, err)
		}
		// Default to 127.0.0.1:6550 for k3d
		host = "127.0.0.1"
		port = "6550"
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
			_ = conn.Close()
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
