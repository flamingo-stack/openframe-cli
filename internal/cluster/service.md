<!-- source-hash: d346b7aaea062218ee5b3d5458600420 -->
This file implements cluster lifecycle management operations, providing a service layer for creating, deleting, and managing Kubernetes clusters using k3d.

## Key Components

**ClusterService** - Main service struct that orchestrates cluster operations
- `manager` - K3d cluster manager for low-level operations
- `executor` - Command executor for running system commands
- `suppressUI` - Flag to disable interactive UI elements

**Constructor Functions:**
- `NewClusterService()` - Creates service with default configuration
- `NewClusterServiceSuppressed()` - Creates service with UI suppression for automation
- `NewClusterServiceWithOptions()` - Creates service with custom manager

**Core Operations:**
- `CreateCluster()` - Handles cluster creation with conflict detection
- `DeleteCluster()` - Manages cluster deletion with progress indicators
- `ListClusters()` - Returns list of all available clusters
- `GetClusterStatus()` - Retrieves status information for a specific cluster
- `CleanupCluster()` - Performs comprehensive cleanup of cluster resources

**Cleanup Functions:**
- `cleanupHelmReleases()` - Removes all Helm releases
- `cleanupKubernetesResources()` - Cleans up common namespaces
- `cleanupDockerResources()` - Prunes Docker images and containers

## Usage Example

```go
// Create a cluster service
executor := executor.NewCommandExecutor()
service := NewClusterService(executor)

// Create a new cluster
config := models.ClusterConfig{
    Name: "my-cluster",
    Type: models.ClusterTypeK3d,
}
err := service.CreateCluster(config)

// Check cluster status
status, err := service.GetClusterStatus("my-cluster")

// Clean up cluster resources
err = service.CleanupCluster("my-cluster", models.ClusterTypeK3d, true, false)
```