package k3d

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

// PortConfig holds the allocated ports for a k3d cluster
type PortConfig struct {
	API   int
	HTTP  int
	HTTPS int
}

// findAvailablePorts finds available TCP ports for API, HTTP, and HTTPS
// It prefers standard ports (6550, 80, 443) and falls back to high ports (6551, 8080, 8443) if needed
func (m *K3dManager) findAvailablePorts() (PortConfig, error) {
	// Get ports used by existing k3d clusters
	usedPorts := m.getUsedPortsByExistingClusters()

	config := PortConfig{}

	// Find API port (6550 preferred, 6551 fallback)
	config.API = m.findPort([]int{6550, 6551}, 6552, usedPorts)
	if config.API == 0 {
		return config, fmt.Errorf("could not find available API port")
	}

	// Find HTTP port (80 preferred, 8080 fallback)
	config.HTTP = m.findPort([]int{80, 8080}, 8081, usedPorts)
	if config.HTTP == 0 {
		return config, fmt.Errorf("could not find available HTTP port")
	}

	// Find HTTPS port (443 preferred, 8443 fallback)
	config.HTTPS = m.findPort([]int{443, 8443}, 8444, usedPorts)
	if config.HTTPS == 0 {
		return config, fmt.Errorf("could not find available HTTPS port")
	}

	return config, nil
}

// findPort tries preferred ports first, then searches from searchStart
func (m *K3dManager) findPort(preferred []int, searchStart int, usedPorts map[int]bool) int {
	// Try preferred ports first
	for _, port := range preferred {
		if !usedPorts[port] && m.isPortAvailable(port) {
			return port
		}
	}

	// Search for an available port
	for port := searchStart; port < searchStart+1000; port++ {
		if !usedPorts[port] && m.isPortAvailable(port) {
			return port
		}
	}

	return 0
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

// isPortAvailable checks if a TCP port is available by attempting to connect to it.
// If connection is refused, the port is available. This approach works regardless of
// user privileges (unlike bind-based checks which fail for ports < 1024 without root).
func (m *K3dManager) isPortAvailable(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err != nil {
		// Connection refused or timeout means port is available
		return true
	}
	// Connection succeeded means something is listening
	_ = conn.Close()
	return false
}
