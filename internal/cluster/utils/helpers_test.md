<!-- source-hash: aa61ddffaca826a1176848d3f3529acb -->
A comprehensive test suite for cluster utility helper functions, covering validation, parsing, and error handling for Kubernetes cluster operations.

## Key Components

- **TestValidateClusterName**: Tests cluster name validation with various edge cases including empty strings, whitespace, and invalid characters
- **TestParseClusterType**: Tests parsing and case-insensitive handling of cluster types (k3d, GKE) with fallback to default
- **TestGetNodeCount**: Tests node count validation with boundary conditions and default value handling
- **TestClusterSelectionResult**: Tests the cluster selection result structure
- **TestCreateClusterError**: Tests error creation and wrapping for cluster operations
- **TestTypeAliases**: Validates model type aliases work correctly
- **TestEdgeCases**: Comprehensive edge case testing for all utility functions

## Usage Example

```go
// Run specific test functions
func TestExample(t *testing.T) {
    // Test cluster name validation
    err := ValidateClusterName("my-cluster")
    assert.NoError(t, err)
    
    // Test cluster type parsing
    clusterType := ParseClusterType("k3d")
    assert.Equal(t, models.ClusterTypeK3d, clusterType)
    
    // Test node count handling
    nodeCount := GetNodeCount(0) // Returns default of 3
    assert.Equal(t, 3, nodeCount)
    
    // Test error creation
    err = CreateClusterError("create", "test-cluster", models.ClusterTypeK3d, someError)
    assert.Error(t, err)
}
```