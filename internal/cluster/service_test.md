<!-- source-hash: f3e2225c195de0618c4e765372263f1f -->
Test file for the cluster service package that validates cluster management functionality using mock and real executors.

## Key Components

- **createTestExecutor()** - Creates a mock command executor with pre-configured responses for k3d cluster operations
- **TestNewClusterService()** - Tests basic service initialization with default configuration
- **TestNewClusterServiceWithOptions()** - Tests service initialization with custom cluster manager
- **TestClusterService_CreateCluster()** - Validates cluster creation functionality
- **TestClusterService_DeleteCluster()** - Tests cluster deletion operations
- **TestClusterService_ListClusters()** - Verifies cluster listing capability
- **TestClusterService_GetClusterStatus()** - Tests cluster status retrieval
- **TestClusterService_DetectClusterType()** - Validates cluster type detection
- **TestClusterService_CleanupCluster()** - Tests cluster cleanup operations
- **TestClusterService_ShowClusterStatus()** - Validates status display functionality
- **TestClusterService_DisplayClusterList()** - Tests cluster list display with various options
- **TestClusterService_WithRealExecutor()** - Tests service with real executor in dry-run mode

## Usage Example

```go
func TestCustomCluster(t *testing.T) {
    // Create test executor with mock responses
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
}
```