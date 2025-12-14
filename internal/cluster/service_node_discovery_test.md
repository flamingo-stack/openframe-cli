<!-- source-hash: 65557cee0a7eedb13cd2596556ffe292 -->
Test suite for the k3d cluster node discovery functionality in the cluster service. This file validates node detection, filtering, and Docker resource cleanup operations.

## Key Components

- **`TestClusterService_getK3dClusterNodes`** - Tests discovery of k3d cluster nodes via Docker commands
- **`TestClusterService_filterK3dNodes`** - Tests filtering of Docker container output to extract valid k3d nodes
- **`TestClusterService_isK3dWorkerNode`** - Tests identification of valid k3d worker nodes (servers/agents vs infrastructure containers)
- **`TestClusterService_cleanupDockerResources_Integration`** - Integration test for the complete Docker resource cleanup workflow

## Usage Example

```go
// Running the node discovery tests
func TestExample(t *testing.T) {
    mockExec := executor.NewMockCommandExecutor()
    
    // Set up mock Docker response
    mockExec.SetResponse("docker ps", &executor.CommandResult{
        Stdout: "k3d-test-cluster-server-0\nk3d-test-cluster-agent-0",
    })
    
    service := NewClusterService(mockExec)
    nodes, err := service.getK3dClusterNodes(context.Background(), "test-cluster")
    
    assert.NoError(t, err)
    assert.Equal(t, []string{"k3d-test-cluster-server-0", "k3d-test-cluster-agent-0"}, nodes)
}
```

The tests cover edge cases like empty cluster names, Docker command failures, and filtering out infrastructure containers (serverlb, tools) from worker nodes.