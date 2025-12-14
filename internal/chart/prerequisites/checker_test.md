<!-- source-hash: c418156df72047ce7cec0816a0692c36 -->
Test file for the prerequisites checker system that validates prerequisite tool detection and installation guidance functionality.

## Key Components

- **TestNewPrerequisiteChecker**: Verifies that the checker initializes with 4 required tools (Git, Helm, Memory, Certificates) in the correct order
- **TestInstallHelp**: Validates that each prerequisite checker provides non-empty installation instructions
- **TestCheckAllWithMissingTools**: Tests detection of missing prerequisites and proper reporting
- **TestCheckAllWithAllTools**: Verifies behavior when all prerequisites are satisfied
- **TestGetInstallInstructions**: Tests retrieval of installation instructions for specific missing tools

## Usage Example

```go
// Run all tests
go test ./internal/chart/prerequisites

// Run specific test
go test -run TestNewPrerequisiteChecker ./internal/chart/prerequisites

// Test with verbose output
go test -v ./internal/chart/prerequisites

// The tests mock prerequisite states:
checker := NewPrerequisiteChecker()
checker.requirements[0].IsInstalled = func() bool { return false } // Mock Git as missing
allPresent, missing := checker.CheckAll()
// Validates that missing tools are properly detected and reported
```