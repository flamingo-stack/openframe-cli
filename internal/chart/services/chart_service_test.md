<!-- source-hash: 8f3e2aac5c85f90a56ed929f62fd5643 -->
Test file for ChartService containing unit tests and mock implementations for chart installation workflows. It provides test coverage for service initialization, cluster listing functionality, and installation request validation.

## Key Components

- **MockClusterLister**: Mock implementation of ClusterLister interface for testing
- **NewMockClusterLister()**: Factory function to create mock cluster lister
- **Test Functions**: Comprehensive unit tests for ChartService initialization and InstallationWorkflow
- **SetClusters/SetError**: Mock configuration methods for testing different scenarios

## Usage Example

```go
// Create a mock cluster lister for testing
mockLister := NewMockClusterLister()

// Configure mock with test data
clusters := []clusterDomain.ClusterInfo{
    {Name: "test-cluster", Status: "running"},
}
mockLister.SetClusters(clusters)

// Test service initialization
service := NewChartService(false, false)
assert.NotNil(t, service)

// Create workflow with mocks
workflow := &InstallationWorkflow{
    chartService:   service,
    clusterService: mockLister,
}

// Test installation request
req := types.InstallationRequest{
    Args:         []string{"test-cluster"},
    GitHubRepo:   "https://github.com/test/repo",
    GitHubBranch: "main",
}
```