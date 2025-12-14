<!-- source-hash: 64b1ae8db3bd8a70cb0c9fc7b99f8b12 -->
Test file for the K3d cluster manager implementation that verifies functionality for creating, managing, and monitoring k3d clusters.

## Key Components

- **MockExecutor**: Mock implementation of CommandExecutor interface for testing
- **Test Functions**:
  - `TestNewK3dManager` - Constructor testing with verbose mode options
  - `TestCreateClusterManagerWithExecutor` - Factory function validation with nil checks
  - `TestK3dManager_CreateCluster` - Cluster creation with various configurations and error scenarios
  - `TestK3dManager_DeleteCluster` - Cluster deletion testing
  - `TestK3dManager_StartCluster` - Cluster startup validation
  - `TestK3dManager_ListClusters` - JSON parsing and cluster listing functionality
  - `TestK3dManager_GetClusterStatus` - Status retrieval testing

## Usage Example

```go
// Create mock executor for testing
executor := &MockExecutor{}
executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(
    &execPkg.CommandResult{Stdout: "success"}, nil)

// Test cluster creation
manager := NewK3dManager(executor, false)
config := models.ClusterConfig{
    Name:      "test-cluster",
    Type:      models.ClusterTypeK3d,
    NodeCount: 3,
}

err := manager.CreateCluster(context.Background(), config)
assert.NoError(t, err)
executor.AssertExpectations(t)
```

The tests cover successful operations, error conditions, input validation, and JSON response parsing for comprehensive K3d manager functionality verification.