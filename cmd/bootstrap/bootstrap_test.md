<!-- source-hash: d5189ccce94d8c8e55593890fbc45556 -->
Test suite for the bootstrap command functionality that validates command structure, arguments, and metadata.

## Key Components

- **TestBootstrapCommand** - Validates basic bootstrap command structure including name, descriptions, and RunE function presence
- **TestBootstrapCommandStructure** - Tests detailed command properties like usage syntax, aliases, subcommands, and help text content
- **TestBootstrapArgumentValidation** - Verifies argument validation logic accepts 0-1 arguments and rejects multiple arguments
- **testutil.InitializeTestMode()** - Test initialization helper called in init function

## Usage Example

```go
// Run specific test function
go test -run TestBootstrapCommand

// Run all bootstrap tests
go test ./bootstrap

// Test with verbose output
go test -v ./bootstrap

// Example of what the tests validate:
func TestBootstrapCommand(t *testing.T) {
    cmd := GetBootstrapCmd()
    assert.Equal(t, "bootstrap", cmd.Name())
    assert.NotNil(t, cmd.RunE)
}
```

The tests ensure the bootstrap command follows expected CLI patterns with proper usage syntax ("bootstrap [cluster-name]"), contains required help text mentioning related commands, and validates arguments correctly.