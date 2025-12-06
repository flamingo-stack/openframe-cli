package manager

import (
	"runtime"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// ManagerType represents the type of cluster manager to use
type ManagerType string

const (
	ManagerTypeK3d  ManagerType = "k3d"
	ManagerTypeKind ManagerType = "kind"
)

// GetManagerType returns the appropriate manager type for the current OS
func GetManagerType() ManagerType {
	switch runtime.GOOS {
	case "windows":
		// Windows uses native KIND executable (no WSL dependency)
		return ManagerTypeKind
	case "darwin", "linux":
		// Unix-like systems use K3d
		return ManagerTypeK3d
	default:
		// Fallback to K3d
		return ManagerTypeK3d
	}
}

// ManagerFactory is a function type that creates a ClusterManager
type ManagerFactory func(exec executor.CommandExecutor, verbose bool) ClusterManager

// registry holds the registered manager factories
var registry = make(map[ManagerType]ManagerFactory)

// RegisterManager registers a manager factory for a specific manager type
func RegisterManager(managerType ManagerType, factory ManagerFactory) {
	registry[managerType] = factory
}

// GetClusterManager returns the appropriate ClusterManager for the current OS
func GetClusterManager(exec executor.CommandExecutor, verbose bool) ClusterManager {
	managerType := GetManagerType()
	if factory, exists := registry[managerType]; exists {
		return factory(exec, verbose)
	}
	// Fallback: try to get any registered manager
	for _, factory := range registry {
		return factory(exec, verbose)
	}
	return nil
}

// GetClusterManagerByType returns a ClusterManager of the specified type
func GetClusterManagerByType(managerType ManagerType, exec executor.CommandExecutor, verbose bool) ClusterManager {
	if factory, exists := registry[managerType]; exists {
		return factory(exec, verbose)
	}
	return nil
}
