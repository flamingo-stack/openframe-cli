<!-- source-hash: 1e2ef8c90a0bea8e5648538cb1c325fa -->
The K3D cluster manager provides comprehensive operations for managing K3D Kubernetes clusters. It handles cluster creation, deletion, lifecycle management, and configuration with intelligent port allocation and resource management.

## Key Components

- **K3dManager**: Main struct implementing cluster management operations
- **ClusterManager**: Interface defining cluster management contract
- **NewK3dManager/NewK3dManagerWithTimeout**: Constructors for creating manager instances
- **CreateCluster**: Creates clusters using dynamic config generation
- **DeleteCluster/StartCluster**: Lifecycle management operations
- **ListClusters/GetClusterStatus**: Query operations for cluster information
- **DetectClusterType**: Determines if a cluster is K3D-managed
- **GetKubeconfig**: Retrieves cluster authentication configuration

## Usage Example

```go
// Create a new K3D manager
executor := executor.NewRealExecutor()
manager := NewK3dManager(executor, true)

// Create a cluster
config := models.ClusterConfig{
    Name:      "dev-cluster",
    Type:      models.ClusterTypeK3d,
    NodeCount: 3,
    K8sVersion: "v1.31.5-k3s1",
}
err := manager.CreateCluster(context.Background(), config)

// List all clusters
clusters, err := manager.ListClusters(context.Background())

// Get cluster status
status, err := manager.GetClusterStatus(context.Background(), "dev-cluster")

// Clean up
err = manager.DeleteCluster(context.Background(), "dev-cluster", models.ClusterTypeK3d, false)
```