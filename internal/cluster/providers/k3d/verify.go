package k3d

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
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

	// No Windows branch: the CLI forwards into WSL and runs as linux (see wsllauncher),
	// so the file-based kubeconfig is always used.
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

	// No Windows branch: the CLI forwards into WSL and runs as linux (see wsllauncher).
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

// cleanupStaleLockFiles removes any stale kubeconfig lock files.
//
// No Windows branch: the CLI forwards into WSL and runs as linux (see wsllauncher).
func (m *K3dManager) cleanupStaleLockFiles(ctx context.Context) error {
	// Linux/macOS: Remove lock files
	cleanupCmd := "rm -f ~/.kube/config.lock ~/.kube/config.lock.* 2>/dev/null || true"
	_, err := m.executor.Execute(ctx, "bash", "-c", cleanupCmd)
	if err != nil {
		return fmt.Errorf("failed to cleanup lock files: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Cleaned up stale kubeconfig lock files")
	}

	return nil
}
