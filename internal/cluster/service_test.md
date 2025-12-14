Test suite for the cluster service functionality, providing comprehensive unit tests for cluster management operations.

## Key Components

- **createTestExecutor()** - Creates a mock executor with pre-configured responses for k3d cluster operations
- **Test Functions**:
  - `TestNewClusterService` - Tests basic service initialization
  - `TestNewClusterServiceWithOptions` - Tests service initialization with custom manager
  - `TestClusterService_CreateCluster` - Tests cluster creation functionality
  - `TestClusterService_DeleteCluster` - Tests cluster deletion
  - `TestClusterService_ListClusters` - Tests cluster listing
  - `TestClusterService_GetClusterStatus` - Tests cluster status retrieval
  - `TestClusterService_DetectClusterType` - Tests cluster type detection
  - `TestClusterService_CleanupCluster` - Tests cluster cleanup operations
  - `TestClusterService_ShowClusterStatus` - Tests cluster status display
  - `TestClusterService_DisplayClusterList` - Tests cluster list display formatting
  - `TestClusterService_WithRealExecutor` - Tests with real executor in dry-run mode

## Usage Example

```go
func TestCustomClusterOperation(t *testing.T) {
    // Create a test executor with mock responses
    exec := createTestExecutor()
    service := NewClusterService(exec)
    
    // Test cluster creation
    config := models.ClusterConfig{
        Name:       "my-test-cluster",
        Type:       models.ClusterTypeK3d,
        NodeCount:  3,
        K8sVersion: "v1.25.0",
    }
    
    err := service.CreateCluster(config)
    if err != nil {
        t.Errorf("Failed to create cluster: %v", err)
    }
}
```