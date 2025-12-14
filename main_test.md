Integration test suite that builds and tests the main CLI binary to verify command-line interface functionality and output validation.

## Key Components

- **TestMainIntegration**: Main test function that builds the CLI binary and runs integration tests
- **Test binary compilation**: Dynamically builds the main executable for testing
- **Command execution tests**: Validates CLI behavior for help, version, and error scenarios
- **Output validation**: Checks both stdout and stderr for expected content

## Usage Example

```go
// Run the integration tests
go test -v ./main_test.go

// The test automatically:
// 1. Builds a test binary named "openframe-test-main"
// 2. Executes various CLI commands with different flags
// 3. Validates exit codes and output content
// 4. Cleans up the test binary

// Example test case structure:
tests := []struct {
    name     string
    args     []string  
    wantErr  bool
    contains string
}{
    {
        name:     "help",
        args:     []string{"--help"},
        wantErr:  false,
        contains: "OpenFrame CLI",
    },
}
```

The test suite ensures the CLI properly handles help flags, version requests, and invalid arguments while producing appropriate output messages.