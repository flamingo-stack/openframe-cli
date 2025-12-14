<!-- source-hash: 888513db42af1264814a8c0a9bef89bf -->
This file contains comprehensive unit tests for the UI display service that handles console output formatting for cluster management operations. The tests verify message formatting, table display, and error handling scenarios.

## Key Components

- **TestNewDisplayService**: Tests service instantiation
- **TestDisplayService_ShowCluster***: Tests for cluster lifecycle messages (creation, deletion, start)
- **TestDisplayService_ShowClusterList**: Tests cluster listing with table formatting
- **TestDisplayService_ShowClusterStatus**: Tests detailed cluster status display with node information
- **TestDisplayService_ShowConfigurationSummary**: Tests configuration summary with dry-run and wizard modes
- **TestClusterDisplayInfo/TestNodeDisplayInfo**: Tests data structure creation

## Usage Example

```go
func TestExample(t *testing.T) {
    // Create display service
    service := NewDisplayService()
    var buf bytes.Buffer
    
    // Test cluster creation message
    service.ShowClusterCreationStart("my-cluster", "k3d", &buf)
    output := buf.String()
    assert.Contains(t, output, "Creating k3d cluster 'my-cluster'...")
    
    // Test cluster list display
    clusters := []ClusterDisplayInfo{
        {
            Name: "test-cluster",
            Type: "k3d", 
            Status: "running",
            NodeCount: 3,
            CreatedAt: time.Now(),
        },
    }
    service.ShowClusterList(clusters, &buf)
}
```

The tests use `bytes.Buffer` for output capture and `stretchr/testify` for assertions, covering both happy path scenarios and edge cases like empty inputs.