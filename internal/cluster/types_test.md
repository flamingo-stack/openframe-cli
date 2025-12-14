This file contains comprehensive test cases for the cluster package's type system and data structures.

## Key Components

- **FlagContainer Tests**: Validates the creation, synchronization, and reset functionality of the flag container that manages CLI command flags
- **Domain Type Tests**: Tests cluster types (`k3d`, `gke`) and domain constants from the models package
- **Model Structure Tests**: Validates `ClusterConfig`, `ClusterInfo`, `NodeInfo`, and `ProviderOptions` data structures
- **Error Type Tests**: Tests custom error types including cluster not found, provider not found, invalid config, and operation errors
- **Interface Compliance Tests**: Verifies that implementations correctly satisfy `ClusterService` and `ClusterManager` interfaces

## Usage Example

```go
func TestCustomClusterConfig(t *testing.T) {
    // Test flag container operations
    container := NewFlagContainer()
    container.Global.Verbose = true
    container.SyncGlobalFlags()
    
    // Test domain models
    config := models.ClusterConfig{
        Name:       "my-cluster",
        Type:       models.ClusterTypeK3d,
        NodeCount:  3,
        K8sVersion: "v1.31.5-k3s1",
    }
    
    // Test error handling
    err := models.NewClusterNotFoundError("missing-cluster")
    var clusterErr models.ErrClusterNotFound
    if errors.As(err, &clusterErr) {
        // Handle cluster not found error
    }
}
```