<!-- source-hash: 0784f293abc3ca36758a773631f50d63 -->
Test file for the k3d package that validates k3d installer functionality across different operating systems and environments.

## Key Components

- **TestNewK3dInstaller**: Tests instantiation of k3d installer
- **TestK3dInstaller_GetInstallHelp**: Validates platform-specific installation help messages
- **TestK3dInstaller_Install**: Tests installation logic and error handling for different platforms
- **TestCommandExists**: Verifies command existence checking utility
- **TestInstallScript**: Linux-specific test for script-based installation
- **containsSubstring**: Helper function for string matching in tests

## Usage Example

```go
// Run specific test
go test -run TestNewK3dInstaller

// Run all tests
go test ./...

// Run with verbose output
go test -v

// Skip platform-specific tests
go test -short
```

The tests cover platform-specific behavior for macOS (brew), Linux (curl/script), and Windows (chocolatey/manual), ensuring appropriate error messages and installation paths are provided for each environment. Tests are designed to work in CI environments where actual installation cannot be performed.