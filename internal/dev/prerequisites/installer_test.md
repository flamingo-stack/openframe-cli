<!-- source-hash: 584aa0c302cde0872f353392b7eb7d26 -->
Test file for the prerequisites installer component that validates prerequisite tool detection and installation functionality.

## Key Components

- **TestNewInstaller**: Validates proper initialization of the installer instance and its internal checker
- **TestInstaller_CheckSilent**: Tests silent prerequisite checking without user prompts, verifying return value consistency
- **TestInstaller_CheckSpecificTools**: Comprehensive test suite for checking specific tools with various scenarios (single/multiple tools, empty lists, unknown tools, case sensitivity)
- **TestCheckTelepresenceAndJq**: Backward compatibility test (currently skipped due to user interaction requirements)
- **TestInstaller_Integration**: Integration test validating installer-checker interaction and install instruction generation

## Usage Example

```go
// Run the complete test suite
go test ./prerequisites

// Run specific test
go test -run TestNewInstaller ./prerequisites

// Run with verbose output
go test -v ./prerequisites

// Test specific tool checking scenarios
go test -run TestInstaller_CheckSpecificTools ./prerequisites
```

The tests use table-driven testing patterns and validate both success and edge cases, ensuring the installer correctly handles tool detection, error scenarios, and provides appropriate feedback for missing prerequisites.