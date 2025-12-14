This file contains unit tests for the bootstrap command functionality, validating command structure, arguments, and configuration.

## Key Components

- **TestBootstrapCommand()** - Tests basic command structure including name, descriptions, and RunE function presence
- **TestBootstrapCommandStructure()** - Validates bootstrap-specific command properties like usage syntax, subcommands, and help content
- **TestBootstrapArgumentValidation()** - Tests argument validation rules for the bootstrap command
- **init()** - Initializes test mode using testutil package

## Usage Example

```go
// Run all bootstrap command tests
go test -v ./bootstrap

// Run specific test
go test -run TestBootstrapCommand ./bootstrap

// Test command structure validation
func TestCustomBootstrap(t *testing.T) {
    cmd := GetBootstrapCmd()
    assert.Equal(t, "bootstrap", cmd.Name())
    assert.NotNil(t, cmd.RunE)
}
```

The tests verify that the bootstrap command:
- Has proper name and descriptions
- Uses correct syntax (`bootstrap [cluster-name]`)
- Contains expected help content mentioning cluster creation and chart installation
- Accepts 0 or 1 arguments but rejects multiple arguments
- Functions as a leaf command without subcommands