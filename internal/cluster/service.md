This service provides comprehensive cluster lifecycle management operations for Kubernetes clusters, primarily focused on K3d cluster operations with support for creation, deletion, status checking, and cleanup.

## Key Components

- **ClusterService**: Main service struct handling cluster operations with optional UI suppression for automation
- **NewClusterService()**: Creates a new service instance with default configuration
- **NewClusterServiceSuppressed()**: Creates a service with UI elements suppressed for automation
- **CreateCluster()**: Handles cluster creation with duplicate detection and status display
- **DeleteCluster()**: Manages cluster deletion with progress indicators
- **ListClusters()**: Retrieves all available clusters
- **GetClusterStatus()**: Gets detailed status information for a specific cluster
- **CleanupCluster()**: Performs comprehensive cleanup of cluster resources including Helm releases, Kubernetes resources, and Docker images

## Usage Example

```go
// Create a new cluster service
exec := executor.NewRealCommandExecutor()
service := NewClusterService(exec)

// Create a new cluster
config := models.ClusterConfig{
    Name: "my-cluster",
    Type: models.ClusterTypeK3d,
}
err := service.CreateCluster(config)

// List all clusters
clusters, err := service.ListClusters()

// Get cluster status
info, err := service.GetClusterStatus("my-cluster")

// Clean up cluster resources
err = service.CleanupCluster("my-cluster", models.ClusterTypeK3d, true, false)

// Delete cluster
err = service.DeleteCluster("my-cluster", models.ClusterTypeK3d, false)
```