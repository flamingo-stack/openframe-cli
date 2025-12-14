Test file that validates the bootstrap command structure and functionality in the OpenFrame CLI. It ensures the bootstrap command is properly configured with correct metadata, argument validation, and executable behavior.

## Key Components

- **TestBootstrapCommand**: Basic structure validation for the bootstrap command
- **TestBootstrapCommandStructure**: Comprehensive testing of command metadata, descriptions, and examples  
- **TestBootstrapArgumentValidation**: Validates argument parsing logic (0-1 arguments accepted)
- **testutil.InitializeTestMode()**: Test environment initialization in init function

## Usage Example

```go
// Run all bootstrap command tests
go test ./bootstrap

// Run specific test function
go test -run TestBootstrapCommand ./bootstrap

// Test with verbose output
go test -v ./bootstrap

// Example of what the tests validate:
cmd := GetBootstrapCmd()
assert.Equal(t, "bootstrap", cmd.Name())
assert.Contains(t, cmd.Short, "Bootstrap complete OpenFrame environment")
assert.NotNil(t, cmd.RunE)
```

The tests ensure the bootstrap command follows CLI conventions with proper help text, argument validation (accepting optional cluster name), and contains references to underlying commands like `openframe cluster create` and `openframe chart install`.