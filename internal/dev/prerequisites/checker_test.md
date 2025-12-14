<!-- source-hash: 03fae4ddd2a6acc318a412f1d81442d0 -->
This file contains unit tests for the prerequisite checker functionality that validates required development tools. It ensures the checker can properly detect missing tools and provide installation guidance.

## Key Components

- **TestNewPrerequisiteChecker**: Validates the checker initializes with the correct number of requirements (3 tools: Telepresence, jq, Skaffold)
- **TestPrerequisiteChecker_CheckAll**: Tests the main checking functionality that returns installation status and missing tools list
- **TestPrerequisiteChecker_GetInstallInstructions**: Comprehensive test with multiple scenarios including case sensitivity and unknown tools
- **TestCheckPrerequisites**: Placeholder test for interactive functionality (currently skipped)
- **TestRequirement_Structure**: Validates that each requirement has proper structure with non-empty fields and callable functions

## Usage Example

```go
func TestToolValidation(t *testing.T) {
    // Create a new checker instance
    checker := NewPrerequisiteChecker()
    
    // Verify it initializes with expected tools
    assert.Len(t, checker.requirements, 3)
    
    // Test installation checking
    allPresent, missing := checker.CheckAll()
    if !allPresent {
        // Get installation instructions for missing tools
        instructions := checker.GetInstallInstructions(missing)
        assert.NotEmpty(t, instructions)
    }
}
```