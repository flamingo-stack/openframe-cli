package k3d

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

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
