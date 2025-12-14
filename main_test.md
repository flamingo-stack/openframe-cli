Integration test file that validates the main CLI application by building and executing the binary with various command-line arguments.

## Key Components

- **TestMainIntegration**: Main test function that builds a test binary and runs integration tests against the CLI application
- **Test cases**: Validates help flag, version flag, and error handling for invalid flags
- **Binary execution**: Uses `os/exec` to run the compiled binary and capture output

## Usage Example

```go
// Run the integration test
go test -v ./main_test.go

// The test will:
// 1. Build a test binary named "openframe-test-main"
// 2. Execute it with different arguments
// 3. Verify expected outputs and exit codes

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

The test ensures the CLI properly handles help/version flags and gracefully fails on invalid arguments, providing confidence that the main application entry point works correctly.