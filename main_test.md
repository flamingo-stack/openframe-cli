Integration test file that validates the main CLI application by building and executing it with various command-line arguments to ensure proper functionality and error handling.

## Key Components

- **TestMainIntegration**: Main integration test function that builds the CLI binary and tests different command scenarios
- **Test cases**: Validates help flag, version flag, and invalid flag handling
- **Binary building**: Dynamically compiles the application for testing
- **Output validation**: Captures and verifies stdout/stderr content

## Usage Example

```go
// Run the integration test
go test -v main_test.go main.go

// The test will:
// 1. Build a test binary named "openframe-test-main"
// 2. Execute it with different flags
// 3. Verify expected output and exit codes

// Test structure follows this pattern:
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

This test ensures the CLI application correctly handles standard flags like `--help` and `--version`, while properly rejecting invalid arguments with appropriate error messages.