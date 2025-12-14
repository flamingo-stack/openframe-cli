This file contains comprehensive unit tests for the cluster service package, validating all public methods and ensuring proper initialization with both mock and real command executors.

## Key Components

- **`createTestExecutor()`** - Creates a mock command executor with pre-configured responses for testing k3d operations
- **Service Constructor Tests** - Validates `NewClusterService()` and `NewClusterServiceWithOptions()` initialization
- **Core Operation Tests** - Tests for cluster creation, deletion, listing, status checking, and cleanup operations
- **Display Method Tests** - Tests for status display and cluster list formatting with various output modes
- **Real Executor Test** - Validates service works with actual command executor in dry-run mode

## Usage Example

```go
// Run the complete test suite
go test ./internal/cluster

// Run specific test
go test -run TestNewClusterService ./internal/cluster

// Test with verbose output
go test -v ./internal/cluster

// Example of how the test creates a mock service
exec := createTestExecutor()
service := NewClusterService(exec)
config := models.ClusterConfig{
    Name:       "test-cluster",
    Type:       models.ClusterTypeK3d,
    NodeCount:  1,
    K8sVersion: "v1.25.0",
}
err := service.CreateCluster(config)
```

The tests use mock executors to avoid requiring actual k3d installations and provide isolated unit testing of the service logic.