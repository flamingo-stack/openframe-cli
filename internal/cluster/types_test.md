<!-- source-hash: 0c21a64ba6e481832a4861d3442f3d5f -->
Test file for the cluster package types and data structures, verifying flag containers, domain models, error types, and interface implementations.

## Key Components

- **FlagContainer Tests**: Validates flag management including creation, synchronization, and reset functionality
- **Domain Model Tests**: Verifies cluster configuration, info, and node structures work correctly
- **Error Type Tests**: Tests custom error types like `ClusterNotFoundError`, `ProviderNotFoundError`, etc.
- **Interface Compliance Tests**: Ensures K3d manager implements required `ClusterService` and `ClusterManager` interfaces
- **Flag Structure Tests**: Validates various flag types (global, create, delete, list) and their properties

## Usage Example

```go
// Test flag container functionality
container := NewFlagContainer()
container.Global.Verbose = true
container.SyncGlobalFlags() // Propagates global flags to all commands

// Test domain models
config := models.ClusterConfig{
    Name:       "test-cluster",
    Type:       models.ClusterTypeK3d,
    NodeCount:  3,
    K8sVersion: "v1.25.0-k3s1",
}

// Test custom errors
err := models.NewClusterNotFoundError("my-cluster")
var notFoundErr models.ErrClusterNotFound
isExpectedType := errors.As(err, &notFoundErr)

// Test interface implementation
mockExecutor := executor.NewMockCommandExecutor()
manager := k3d.NewK3dManager(mockExecutor, false)
var _ models.ClusterService = manager // Compile-time interface check
```