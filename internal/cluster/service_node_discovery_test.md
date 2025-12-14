Test file for k3d cluster node discovery and filtering functionality in the cluster service package. Contains comprehensive test cases for identifying valid worker nodes and excluding infrastructure containers.

## Key Components

- `TestClusterService_getK3dClusterNodes` - Tests the main node discovery method with various scenarios including empty clusters, command failures, and successful discovery
- `TestClusterService_filterK3dNodes` - Tests filtering logic for raw Docker output, handling whitespace and invalid node names
- `TestClusterService_isK3dWorkerNode` - Tests validation of worker nodes vs infrastructure containers (serverlb, tools)
- `TestClusterService_cleanupDockerResources_Integration` - Integration test verifying the complete cleanup workflow

## Usage Example

```go
// Run specific test
go test -run TestClusterService_getK3dClusterNodes

// Run all cluster discovery tests
go test ./cluster -v

// Test with coverage
go test -cover ./cluster

// The tests verify filtering behavior:
// Valid nodes: k3d-cluster-server-0, k3d-cluster-agent-0
// Filtered out: k3d-cluster-serverlb, k3d-cluster-tools
```

The tests use mock executors to simulate Docker commands and verify that only actual worker nodes (server/agent) are processed while infrastructure containers are properly excluded from operations like cleanup.