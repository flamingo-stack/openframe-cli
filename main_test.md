Integration test file that validates the main CLI application by building and executing it as a separate binary to test command-line behavior.

## Key Components

- **TestMainIntegration**: Main test function that builds the CLI binary and runs integration tests
- **Test Cases**: Validates help flag, version flag, and error handling for invalid flags
- **Binary Building**: Uses `go build` to create a test executable for realistic CLI testing
- **Output Capture**: Captures both stdout and stderr to verify command responses

## Usage Example

```go
// Run the integration tests
go test -v ./main_test.go

// The test builds a temporary binary and tests various CLI scenarios:
// - Help command: expects "OpenFrame CLI" in output
// - Version command: expects "dev" in output  
// - Invalid flag: expects error and "unknown flag" message

// Test structure follows table-driven pattern:
tests := []struct {
    name     string
    args     []string
    wantErr  bool
    contains string
}{
    // test cases...
}
```

The test ensures the main application properly handles standard CLI patterns like help/version flags and gracefully reports errors for invalid input.