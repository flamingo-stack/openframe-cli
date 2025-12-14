<!-- source-hash: f352486529729738428045d23fabdc88 -->
Test file for the prerequisites checker module that validates system requirements for cluster operations.

## Key Components

- **TestNewPrerequisiteChecker**: Tests initialization of the prerequisite checker with expected requirements (Docker, kubectl, k3d)
- **TestCommandExists**: Verifies command existence checking functionality through Docker installer
- **TestInstallHelp**: Tests platform-specific installation help messages for all prerequisite tools
- **TestCheckAllWithMissingTools**: Tests detection and reporting of missing prerequisite tools
- **TestCheckAllWithAllTools**: Tests successful validation when all tools are present
- **TestGetInstallInstructions**: Tests generation of installation instructions for missing tools
- **containsAny**: Helper function to check if a string contains any of the specified substrings

## Usage Example

```go
// Run tests for prerequisite checker
go test -v ./internal/cluster/prerequisites/

// Run specific test
go test -v ./internal/cluster/prerequisites/ -run TestCheckAllWithMissingTools

// Test installation help generation
checker := NewPrerequisiteChecker()
missing := []string{"Docker", "k3d"}
instructions := checker.GetInstallInstructions(missing)
```

The tests validate cross-platform compatibility by checking that installation help messages contain appropriate references for different operating systems (brew for macOS, package managers for Linux, chocolatey for Windows).