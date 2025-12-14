Integration test suite that validates the main CLI application by building and executing the binary with various command-line arguments. Tests core functionality like help, version, and error handling through subprocess execution.

## Key Components

- **TestMainIntegration**: Primary test function that builds the CLI binary and runs integration tests
- **Test cases**: Structured test scenarios covering help display, version output, and invalid flag handling
- **Binary execution**: Uses `os/exec` to run the compiled CLI as a subprocess
- **Output validation**: Captures and validates both stdout and stderr output

## Usage Example

```go
// Run the integration tests
go test -v ./main_test.go

// The test automatically:
// 1. Builds the CLI binary
// 2. Executes it with different arguments
// 3. Validates expected output and exit codes

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

The test suite ensures the CLI behaves correctly across different scenarios by testing the actual compiled binary rather than internal functions, providing true end-to-end validation.