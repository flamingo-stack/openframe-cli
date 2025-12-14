Integration test file that validates the main CLI application by building and executing the binary with various command-line arguments.

## Key Components

- **TestMainIntegration**: Main test function that builds the CLI binary and tests different command-line scenarios
- **Test binary creation**: Dynamically builds a test executable using `go build`
- **Test cases structure**: Defines test scenarios with expected arguments, error states, and output validation
- **Command execution**: Uses `os/exec` to run the built binary and capture output

## Usage Example

```go
// Run the integration test
go test -v ./main_test.go

// The test validates these scenarios:
// 1. Help flag displays CLI information
cmd := exec.Command("./openframe-test-main", "--help")
// Expected: contains "OpenFrame CLI"

// 2. Version flag shows version info  
cmd := exec.Command("./openframe-test-main", "--version")
// Expected: contains "dev"

// 3. Invalid flags return errors
cmd := exec.Command("./openframe-test-main", "--invalid")
// Expected: error with "unknown flag" message
```

The test ensures the CLI behaves correctly for common usage patterns including help display, version checking, and error handling for invalid arguments.