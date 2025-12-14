Integration test file for the main OpenFrame CLI application that validates command-line interface behavior by building and executing the binary with various arguments.

## Key Components

- **TestMainIntegration**: Main test function that builds a test binary and validates CLI behavior
- **Test cases**: Covers help flag, version flag, and invalid flag scenarios
- **Binary execution**: Uses `os/exec` to run the compiled binary with different arguments
- **Output validation**: Captures stdout/stderr and validates expected content and error states

## Usage Example

```go
// Run the integration test
go test -v ./main_test.go

// The test automatically:
// 1. Builds a test binary named "openframe-test-main"
// 2. Executes it with various flags (--help, --version, --invalid)
// 3. Validates output contains expected strings
// 4. Cleans up the test binary

// Example test case structure:
{
    name:     "help",
    args:     []string{"--help"},
    wantErr:  false,
    contains: "OpenFrame CLI",
}
```

This test ensures the CLI properly handles common flags and provides appropriate help/version information while gracefully handling invalid inputs.