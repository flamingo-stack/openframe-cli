This file provides a comprehensive cluster management service that handles Kubernetes cluster lifecycle operations using k3d as the underlying provider.

## Key Components

- **ClusterService**: Main service struct with cluster manager, command executor, and UI suppression options
- **Constructor Functions**: 
  - `NewClusterService()` - Creates service with default configuration
  - `NewClusterServiceSuppressed()` - Creates service with UI suppression for automation
  - `NewClusterServiceWithOptions()` - Creates service with custom k3d manager
- **Core Operations**:
  - `CreateCluster()` - Handles cluster creation with duplicate checking
  - `DeleteCluster()` - Manages cluster deletion with progress tracking
  - `ListClusters()` - Lists all available clusters
  - `GetClusterStatus()` - Retrieves cluster status information
  - `CleanupCluster()` - Comprehensive cleanup of cluster resources
- **Cleanup Functions**:
  - `cleanupHelmReleases()` - Removes all Helm releases
  - `cleanupKubernetesResources()` - Cleans up namespaced resources
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

// Get cluster status
status, err := service.GetClusterStatus("my-cluster")

// Clean up cluster resources
err = service.CleanupCluster("my-cluster", models.ClusterTypeK3d, true, false)

// Delete cluster
err = service.DeleteCluster("my-cluster", models.ClusterTypeK3d, false)
```