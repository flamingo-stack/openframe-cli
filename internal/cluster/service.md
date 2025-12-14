This Go file provides a comprehensive cluster management service that handles Kubernetes cluster lifecycle operations using k3d as the underlying provider.

## Key Components

### Types
- **`ClusterService`** - Main service struct that orchestrates cluster operations
- **`isTerminalEnvironment()`** - Helper function to detect terminal environments

### Constructor Functions
- **`NewClusterService()`** - Creates service with default UI enabled
- **`NewClusterServiceSuppressed()`** - Creates service with UI suppressed for automation
- **`NewClusterServiceWithOptions()`** - Creates service with custom k3d manager

### Core Operations
- **`CreateCluster()`** - Creates new clusters with existence checking and user feedback
- **`DeleteCluster()`** - Removes clusters with progress indication
- **`ListClusters()`** - Returns information about all available clusters
- **`GetClusterStatus()`** - Retrieves status for a specific cluster
- **`DetectClusterType()`** - Identifies the type of an existing cluster

### Cleanup Operations
- **`CleanupCluster()`** - Orchestrates comprehensive cluster cleanup
- **`cleanupK3dCluster()`** - Handles k3d-specific cleanup operations
- **`cleanupHelmReleases()`** - Removes all Helm releases
- **`cleanupKubernetesResources()`** - Cleans up namespace resources
- **`cleanupDockerResources()`** - Removes Docker images and containers

## Usage Example

```go
// Create a cluster service
executor := executor.NewCommandExecutor()
service := cluster.NewClusterService(executor)

// Create a new cluster
config := models.ClusterConfig{
    Name: "my-cluster",
    Type: models.ClusterTypeK3d,
}
err := service.CreateCluster(config)

// List all clusters
clusters, err := service.ListClusters()

// Clean up a cluster before deletion
err = service.CleanupCluster("my-cluster", models.ClusterTypeK3d, true, false)

// Delete the cluster
err = service.DeleteCluster("my-cluster", models.ClusterTypeK3d, false)
```