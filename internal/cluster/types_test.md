This file contains comprehensive unit tests for the cluster package's type definitions and flag management system.

## Key Components

- **FlagContainer Tests**: Tests for `NewFlagContainer()`, `SyncGlobalFlags()`, and `Reset()` methods that manage command-line flags across different cluster operations
- **Domain Model Tests**: Validation of cluster types, configurations, and information structures from the `models` package
- **Error Type Tests**: Tests for custom error types including `ClusterNotFoundError`, `ProviderNotFoundError`, and `InvalidConfigError`
- **Interface Compliance Tests**: Verification that concrete types properly implement `ClusterService` and `ClusterManager` interfaces
- **Flag Structure Tests**: Tests for various flag types including global, create, delete, and list flags

## Usage Example

```go
// Test flag container creation and synchronization
container := NewFlagContainer()
container.Global.Verbose = true
container.SyncGlobalFlags()

// Test domain models
config := models.ClusterConfig{
    Name:       "test-cluster",
    Type:       models.ClusterTypeK3d,
    NodeCount:  3,
    K8sVersion: "v1.25.0-k3s1",
}

// Test error handling
err := models.NewClusterNotFoundError("my-cluster")
var notFoundErr models.ErrClusterNotFound
errors.As(err, &notFoundErr) // true
```

The tests ensure type safety, proper flag synchronization, and correct error handling throughout the cluster management system.