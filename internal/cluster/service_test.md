This file contains comprehensive unit tests for the cluster service functionality, testing both mock and real command execution scenarios.

## Key Components

- **`createTestExecutor()`** - Creates a mock command executor with predefined k3d responses for testing
- **Service Constructor Tests** - Validates `NewClusterService()` and `NewClusterServiceWithOptions()` initialization
- **Core Functionality Tests** - Tests for cluster creation, deletion, listing, status checking, and cleanup operations
- **Display Function Tests** - Tests for status display and cluster list formatting with various output modes
- **Real Executor Test** - Validates service behavior with actual command executor in dry-run mode

## Usage Example

```go
func TestCustomClusterOperation(t *testing.T) {
    // Create mock executor for isolated testing
    exec := createTestExecutor()
    service := NewClusterService(exec)
    
    // Test cluster creation
    config := models.ClusterConfig{
        Name:       "my-cluster",
        Type:       models.ClusterTypeK3d,
        NodeCount:  3,
        K8sVersion: "v1.25.0",
    }
    
    err := service.CreateCluster(config)
    if err != nil {
        t.Errorf("Failed to create cluster: %v", err)
    }
    
    // Test cluster listing
    clusters, err := service.ListClusters()
    if err == nil && len(clusters) > 0 {
        t.Log("Clusters found:", len(clusters))
    }
}
```

The tests use a mock executor approach to ensure reliable, isolated testing without requiring actual k3d installation or cluster operations.