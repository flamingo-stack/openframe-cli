Test suite for the cluster service's node discovery functionality, covering k3d cluster node identification and Docker resource cleanup operations.

## Key Components

**Test Functions:**
- `TestClusterService_getK3dClusterNodes` - Tests discovery of k3d cluster nodes via Docker commands
- `TestClusterService_filterK3dNodes` - Tests filtering of valid k3d worker nodes from Docker output
- `TestClusterService_isK3dWorkerNode` - Tests identification of valid k3d worker nodes (server/agent)
- `TestClusterService_cleanupDockerResources_Integration` - Integration test for full cleanup workflow

**Test Coverage:**
- Node discovery with various Docker outputs
- Filtering logic for k3d infrastructure containers (serverlb, tools)
- Error handling for empty cluster names and failed commands
- Integration testing with mocked executor commands

## Usage Example

```go
// Run specific test
go test -run TestClusterService_getK3dClusterNodes

// Run all node discovery tests
go test -run TestClusterService

// Run with verbose output
go test -v ./cluster
```

The tests use a mock command executor to simulate Docker responses and verify that the cluster service correctly identifies k3d worker nodes while filtering out infrastructure containers like load balancers and tools.