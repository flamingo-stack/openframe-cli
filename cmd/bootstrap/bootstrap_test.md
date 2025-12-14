Test file for the bootstrap command functionality, validating command structure and argument handling.

## Key Components

- **TestBootstrapCommand**: Tests basic command properties like name, descriptions, and RunE function presence
- **TestBootstrapCommandStructure**: Validates bootstrap-specific command structure including usage pattern, help text content, and command hierarchy
- **TestBootstrapArgumentValidation**: Tests argument validation logic to ensure proper handling of 0-1 cluster name arguments
- **testutil.InitializeTestMode()**: Test initialization in the `init()` function

## Usage Example

```go
// Run specific test
func TestBootstrapCommand(t *testing.T) {
    cmd := GetBootstrapCmd()
    assert.Equal(t, "bootstrap", cmd.Name())
    assert.NotEmpty(t, cmd.Short)
    assert.NotNil(t, cmd.RunE)
}

// Test argument validation
func TestBootstrapArgumentValidation(t *testing.T) {
    cmd := GetBootstrapCmd()
    
    // Valid cases
    err := cmd.Args(cmd, []string{})                    // No args
    err = cmd.Args(cmd, []string{"test-cluster"})       // One arg
    
    // Invalid case  
    err = cmd.Args(cmd, []string{"arg1", "arg2"})       // Too many args
    assert.Error(t, err)
}
```

This test suite ensures the bootstrap command maintains proper CLI structure and validates user input according to the expected `bootstrap [cluster-name]` usage pattern.