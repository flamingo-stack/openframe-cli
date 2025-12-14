Integration test file that validates the main CLI application by building and executing the binary with various command-line arguments. It ensures core functionality like help, version, and error handling work correctly.

## Key Components

- **TestMainIntegration**: Main integration test function that builds the CLI binary and tests various command-line scenarios
- **Test cases**: Covers help flag, version flag, and invalid flag error handling
- **Binary execution**: Uses `os/exec` to run the compiled binary and capture output
- **Output validation**: Checks both exit codes and output content

## Usage Example

```go
// Run the integration test
go test -v ./main_test.go

// The test automatically:
// 1. Builds the main binary
// 2. Executes it with different flags
// 3. Validates expected output and exit codes

// Test structure shows how to add new CLI test cases:
tests := []struct {
    name     string
    args     []string  
    wantErr  bool
    contains string
}{
    {
        name:     "new command",
        args:     []string{"--new-flag"},
        wantErr:  false,
        contains: "expected output",
    },
}
```

This test ensures the CLI behaves correctly across different scenarios and provides confidence that the main application entry point functions as expected.