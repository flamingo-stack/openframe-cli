<!-- source-hash: bd56edfb7ac8b9c7857f43d10d0c87eb -->
Integration test file that validates the main CLI application's behavior by building and executing the binary with various command-line arguments. Tests core functionality like help, version, and error handling.

## Key Components

- **TestMainIntegration**: Main integration test function that builds a test binary and runs it with different argument combinations
- **Test Cases**: Covers help flag (`--help`), version flag (`--version`), and invalid flag handling
- **Binary Building**: Dynamically builds the main application for testing
- **Output Validation**: Captures and validates both stdout and stderr output

## Usage Example

```go
// Run the integration tests
go test -v ./main_test.go

// The test will automatically:
// 1. Build a test binary from the main package
// 2. Execute it with various flags
// 3. Validate expected outputs and error conditions

// Example test case execution:
cmd := exec.Command("./openframe-test-main", "--help")
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()

// Validates that help output contains "OpenFrame CLI"
assert.NoError(t, err)
assert.Contains(t, stdout.String(), "OpenFrame CLI")
```

This test ensures the CLI application properly handles command-line arguments and produces expected output before deployment.