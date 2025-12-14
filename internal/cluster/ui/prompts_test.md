<!-- source-hash: cecd54459430d19a36a976b318eb7ddc -->
This file contains comprehensive unit tests for UI prompt functions related to cluster management operations, ensuring proper validation and error handling for interactive cluster selection workflows.

## Key Components

- **TestSelectClusterByName**: Tests cluster selection from a list, including empty cluster handling and name extraction validation
- **TestHandleClusterSelection**: Tests the main cluster selection flow with command-line arguments and fallback to interactive prompts
- **TestConfirmClusterDeletion**: Tests deletion confirmation prompts with force flag handling
- **TestFormatClusterSuccessMessage**: Tests message formatting for successful cluster operations
- **TestValidationLogic**: Tests input validation for cluster names
- **TestConstants**: Tests cluster type constant definitions

## Usage Example

```go
// Run specific test functions
func TestYourClusterFunction(t *testing.T) {
    clusters := []ClusterInfo{
        {Name: "test-cluster", Status: "running"},
    }
    
    // Test with command line args
    result, err := HandleClusterSelection(clusters, []string{"my-cluster"}, "Select")
    assert.NoError(t, err)
    assert.Equal(t, "my-cluster", result)
    
    // Test confirmation with force flag
    confirmed, err := ConfirmClusterDeletion("test-cluster", true)
    assert.NoError(t, err)
    assert.True(t, confirmed)
    
    // Test message formatting
    message := FormatClusterSuccessMessage("test", "k3d", "running")
    assert.Contains(t, message, "Cluster: test")
}
```