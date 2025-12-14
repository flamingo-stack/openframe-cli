<!-- source-hash: d7eec406129aadb093a54feedb353fbd -->
This file contains comprehensive unit tests for the Git checker component, verifying Git installation detection, version retrieval, and validation functionality.

## Key Components

- **TestNewGitChecker** - Tests the creation of a new Git checker instance
- **TestGitChecker_GetInstallInstructions** - Validates that installation instructions contain platform-specific information (macOS, Ubuntu) and verification commands
- **TestGitChecker_IsInstalled** - Tests the Git installation detection method
- **TestGitChecker_GetVersion** - Verifies Git version retrieval when Git is installed
- **TestGitChecker_Validate** - Tests the overall Git validation process
- **TestGitChecker_GetVersion_NotInstalled** - Ensures proper error handling when Git is not available

## Usage Example

```go
// Run all tests
go test ./git

// Run specific test
go test -run TestGitChecker_GetVersion ./git

// Run tests with verbose output
go test -v ./git

// Skip tests that require Git installation
go test -short ./git
```

The tests use conditional skipping to handle environments where Git may not be installed, making them robust across different development setups. Tests verify both success and error cases, ensuring comprehensive coverage of the Git checker functionality.