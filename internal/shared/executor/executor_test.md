<!-- source-hash: b0049878696d3bd1155c26f6a8e87073 -->
This file contains comprehensive test coverage for the executor package, testing command execution functionality including dry-run mode, error handling, and various execution options.

## Key Components

- **TestCommandResult_Output** - Tests the output formatting method that combines stdout and stderr
- **TestNewRealCommandExecutor** - Validates constructor with different dry-run and verbose configurations
- **TestRealCommandExecutor_Execute_*** - Tests basic command execution in various scenarios (dry-run, real commands, failures)
- **TestRealCommandExecutor_ExecuteWithOptions_*** - Tests advanced execution with timeouts, environment variables, and working directories
- **TestRealCommandExecutor_buildEnvStrings** - Tests environment variable string formatting
- **Interface compliance tests** - Ensures proper interface implementation

## Usage Example

```go
// Run tests for the entire package
go test ./executor

// Run specific test with verbose output
go test -v -run TestRealCommandExecutor_Execute_RealCommand ./executor

// Run tests with coverage
go test -cover ./executor

// Test dry-run functionality
func ExampleDryRunTest() {
    executor := NewRealCommandExecutor(true, false)
    ctx := context.Background()
    result, err := executor.Execute(ctx, "echo", "hello")
    // In dry-run mode, no actual command is executed
    assert.NoError(t, err)
    assert.Equal(t, 0, result.ExitCode)
}
```