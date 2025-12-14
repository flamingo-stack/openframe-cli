Integration test file that validates the main CLI application by building and executing the binary with various command-line arguments to ensure proper behavior and output.

## Key Components

- **TestMainIntegration**: Main integration test function that builds the CLI binary and tests different command scenarios
- **Test Cases**: Validates help display, version output, and error handling for invalid flags
- **Binary Execution**: Uses `os/exec` to run the built binary as a subprocess and capture output

## Usage Example

```go
// Run the integration test
go test -v ./main_test.go

// The test automatically:
// 1. Builds a test binary
// 2. Executes it with different arguments
// 3. Validates output and exit codes
// 4. Cleans up the binary

// Test cases cover:
// - Help flag: expects "OpenFrame CLI" in output
// - Version flag: expects "dev" in output  
// - Invalid flag: expects error and "unknown flag" message
```

The test ensures the CLI application properly handles standard command-line interactions and provides appropriate feedback to users.