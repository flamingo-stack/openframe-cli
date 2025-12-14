<!-- source-hash: 64a6c27d01fdb5febfd7c402d45e6786 -->
This file contains unit tests for the OpenFrame CLI root command functionality, including command structure validation, help/version output testing, and system service initialization.

## Key Components

- **TestRootCommand**: Validates basic root command properties (name, description)
- **TestRootCommandHelp**: Tests help flag output using testutil helpers
- **TestRootCommandVersion**: Tests version flag output with default version info
- **TestGetRootCmd**: Tests root command creation with custom version information
- **TestSystemService**: Validates system service initialization and log directory creation
- **TestVersionInfo**: Ensures default version information is properly initialized
- **TestExecuteWithVersion**: Verifies the ExecuteWithVersion function exists without panicking

## Usage Example

```go
// Run specific test
go test -run TestRootCommand ./cmd

// Test version command output
cmd := GetRootCmd(DefaultVersionInfo)
testutil.TestCLICommand(t, cmd, []string{"--version"}, false, "dev", "none", "unknown")

// Test with custom version info
versionInfo := VersionInfo{
    Version: "v1.0.0",
    Commit:  "abc123",
    Date:    "2024-01-01",
}
cmd := GetRootCmd(versionInfo)
```

The tests use a testutil package for CLI command validation and enable test mode to suppress UI output during testing.