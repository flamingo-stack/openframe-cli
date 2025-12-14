<!-- source-hash: 058ae34326fa63ed3afd6cea7208a30a -->
Test suite for the models package's custom error types. Validates error creation, type assertions, message formatting, and error chaining behavior.

## Key Components

**Test Functions:**
- `TestErrClusterNotFound` - Tests cluster not found error creation and unwrapping
- `TestErrProviderNotFound` - Tests provider not found error with different cluster types
- `TestErrInvalidClusterConfig` - Tests configuration validation errors with various field types
- `TestErrClusterAlreadyExists` - Tests duplicate cluster error handling
- `TestErrClusterOperation` - Tests operation errors with cause wrapping
- `TestErrorFormatting` - Validates error message formats
- `TestErrorChaining` - Tests complex error wrapping scenarios

**Tested Error Types:**
- `ErrClusterNotFound` - Missing cluster errors
- `ErrProviderNotFound` - Unsupported provider errors  
- `ErrInvalidClusterConfig` - Configuration validation errors
- `ErrClusterAlreadyExists` - Duplicate cluster errors
- `ErrClusterOperation` - Operational failure errors

## Usage Example

```go
func TestCustomError(t *testing.T) {
    // Test error creation and type assertion
    err := NewClusterNotFoundError("test-cluster")
    
    var clusterErr ErrClusterNotFound
    if errors.As(err, &clusterErr) {
        assert.Equal(t, "test-cluster", clusterErr.Name)
    }
    
    // Test error chaining
    configErr := NewInvalidConfigError("name", "", "empty name")
    opErr := NewClusterOperationError("create", "test", configErr)
    
    // Verify both errors exist in chain
    assert.True(t, errors.As(opErr, &clusterErr))
}
```