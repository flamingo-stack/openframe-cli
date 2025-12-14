<!-- source-hash: 0c33a2faf233f6a00fc476fdf1860c7b -->
This file contains unit tests for the Docker installer functionality, verifying cross-platform installation behavior and help text generation.

## Key Components

- **TestNewDockerInstaller** - Tests creation of new Docker installer instances
- **TestDockerInstaller_GetInstallHelp** - Validates platform-specific installation help messages
- **TestDockerInstaller_Install** - Tests installation logic and error handling across different operating systems
- **containsSubstring** - Helper function for substring matching in test assertions

## Usage Example

```go
// Run all tests
go test ./docker

// Run specific test
go test -run TestDockerInstaller_GetInstallHelp ./docker

// Run tests with verbose output
go test -v ./docker

// Example of what the tests verify:
installer := NewDockerInstaller()

// Tests verify help contains platform-appropriate content
help := installer.GetInstallHelp()
// macOS: contains "brew" or https URLs
// Linux: contains "package manager" or https URLs  
// Windows: contains https URLs

// Tests verify installation error handling
err := installer.Install()
// Unsupported platforms: returns "automatic Docker installation not supported"
// Windows: suggests "Please install Docker Desktop"
// macOS without brew: suggests "Homebrew is required"
```

The tests focus on cross-platform compatibility and proper error messaging rather than actual installation execution, making them suitable for CI/CD environments.