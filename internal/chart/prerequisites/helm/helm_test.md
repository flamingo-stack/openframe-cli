<!-- source-hash: fc341f57fc218cfef05d951f28a6fddb -->
This file contains comprehensive test cases for the Helm installer functionality, validating installer creation, platform-specific installation help, and installation scripts.

## Key Components

- **TestNewHelmInstaller**: Tests the creation of a new Helm installer instance
- **TestHelmInstaller_GetInstallHelp**: Validates platform-specific installation help messages for macOS, Linux, and Windows
- **TestHelmInstaller_Install**: Basic structural testing of the installer without performing actual installation
- **TestInstallScript**: Linux-specific test for the installation script functionality
- **containsSubstring**: Helper function for substring matching in test assertions

## Usage Example

```go
// Run specific test
go test -run TestNewHelmInstaller

// Run all Helm installer tests
go test ./helm

// Run with verbose output
go test -v ./helm

// Skip platform-specific tests
go test -run "^((?!TestInstallScript).)*$" ./helm
```

The tests include platform detection logic and graceful error handling for network-dependent operations. The `TestInstallScript` test is automatically skipped on non-Linux systems, and actual installation testing is avoided to prevent slow, environment-dependent test execution.