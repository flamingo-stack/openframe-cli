Integration test file that validates the main CLI application by building and executing the binary with various command-line arguments.

## Key Components

- **TestMainIntegration**: Main test function that builds the CLI binary and executes integration tests
- **Test cases**: Validates help output, version display, and error handling for invalid flags
- **Binary execution**: Uses `os/exec` to run the compiled binary with different arguments

## Usage Example

```go
// Run the integration test
go test -v main_test.go main.go

// The test automatically:
// 1. Builds a temporary test binary
// 2. Executes it with various flags
// 3. Validates output and exit codes
// 4. Cleans up the binary

// Test cases covered:
// - ./binary --help (expects "OpenFrame CLI" in output)
// - ./binary --version (expects "dev" in output)  
// - ./binary --invalid (expects error with "unknown flag")
```

The test ensures the CLI application properly handles standard flags and provides appropriate feedback for invalid inputs. It uses testify assertions for clean test validation and automatically manages the test binary lifecycle.