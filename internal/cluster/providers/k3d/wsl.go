package k3d

import (
	"context"
	"fmt"
	"time"
)

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

// prepareKubeconfigDirectory ensures ~/.kube directory exists with proper permissions.
//
// No Windows branch: the CLI forwards into WSL and runs as linux (see wsllauncher).
func (m *K3dManager) prepareKubeconfigDirectory(ctx context.Context) error {
	// Linux/macOS: Create .kube directory with proper permissions
	createCmd := "mkdir -p ~/.kube && chmod 755 ~/.kube"
	_, err := m.executor.Execute(ctx, "bash", "-c", createCmd)
	if err != nil {
		return fmt.Errorf("failed to create .kube directory: %w", err)
	}

	if m.verbose {
		fmt.Println("✓ Prepared kubeconfig directory")
	}

	return nil
}

// fixKubeconfigPermissions fixes kubeconfig file permissions.
// This is needed because k3d running with sudo creates ~/.kube/config with root ownership.
//
// No Windows branch: the CLI forwards into WSL and runs as linux (see wsllauncher).
func (m *K3dManager) fixKubeconfigPermissions(ctx context.Context) error {
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

	return nil
}
