<!-- source-hash: 6fe903b0a6f70bd4c9470d670c43968f -->
This test file provides comprehensive testing for SSL/TLS certificate installation functionality, covering certificate installer creation, platform-specific installation help, and mkcert tool management.

## Key Components

- **TestNewCertificateInstaller**: Tests certificate installer instantiation
- **TestCertificateInstaller_GetInstallHelp**: Validates platform-specific installation help messages (macOS/Homebrew, Linux/download, Windows/manual)
- **TestCertificateInstaller_Install**: Basic installer functionality validation without actual installation
- **TestAreCertificatesGenerated**: Tests certificate detection logic
- **TestIsMkcertInstalled**: Validates mkcert tool presence detection
- **TestInstallMkcert**: Tests mkcert installation with graceful error handling
- **containsSubstring**: Helper function for string matching in tests

## Usage Example

```go
// Run specific test
go test -run TestNewCertificateInstaller

// Run all certificate tests
go test ./certificates

// Test with verbose output
go test -v ./certificates

// Example of what the tests validate
installer := NewCertificateInstaller()
help := installer.GetInstallHelp()
// help contains platform-specific instructions mentioning mkcert

// Tests verify error handling for installation failures
err := installer.installMkcert()
// Error messages are validated for expected content
```

The tests focus on functional validation while avoiding slow integration tests, using environment detection to provide appropriate platform-specific behavior verification.