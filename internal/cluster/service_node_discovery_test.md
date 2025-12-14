This test file validates the Kubernetes cluster service's node discovery functionality for k3d clusters. It tests the ability to discover, filter, and identify k3d worker nodes while excluding infrastructure containers.

## Key Components

- **TestClusterService_getK3dClusterNodes**: Tests cluster node discovery with various scenarios including empty clusters, failed commands, and successful discovery
- **TestClusterService_filterK3dNodes**: Validates filtering logic for extracting valid k3d nodes from Docker output
- **TestClusterService_isK3dWorkerNode**: Tests worker node identification, distinguishing between actual worker nodes and infrastructure containers
- **TestClusterService_cleanupDockerResources_Integration**: Integration test verifying the complete flow of node discovery and cleanup command execution

## Usage Example

```go
// Example test structure for node discovery
func TestYourClusterFunction(t *testing.T) {
    mockExec := executor.NewMockCommandExecutor()
    
    // Mock Docker response
    mockExec.SetResponse("docker ps", &executor.CommandResult{
        Stdout: "k3d-test-cluster-server-0\nk3d-test-cluster-agent-0",
    })
    
    service := NewClusterService(mockExec)
    
    nodes, err := service.getK3dClusterNodes(context.Background(), "test-cluster")
    
    assert.NoError(t, err)
    assert.Equal(t, []string{"k3d-test-cluster-server-0", "k3d-test-cluster-agent-0"}, nodes)
}
```

The tests use table-driven patterns and mock executors to verify proper node filtering that excludes serverlb and tools containers while preserving server and agent nodes.