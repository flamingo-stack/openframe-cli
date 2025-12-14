<!-- source-hash: beee0fee9aa7788d4c49a681fa501c1e -->
This file contains comprehensive unit tests for the UI prompts package, focusing on testing validation logic and configuration setup for interactive command-line prompts built with promptui.

## Key Components

- **Validation Function Tests**: Tests for user input validation in confirmation prompts (y/yes/n/no)
- **Result Parsing Tests**: Tests for converting user input to boolean values with case-insensitive handling
- **Template Configuration Tests**: Verification of promptui template setup with Unicode arrows and color formatting
- **Error Handling Tests**: Validation of array length matching in multi-choice prompts
- **Utility Function Tests**: Tests for `boolToString`, `ValidateNonEmpty`, and `ValidateIntRange` functions
- **Benchmark Tests**: Performance testing for validation and parsing functions

## Usage Example

```go
// Run validation tests
go test -v ./ui -run TestConfirmAction_ValidateFunction

// Run benchmark tests
go test -bench=. ./ui

// Test specific validator functions
func TestCustomValidator(t *testing.T) {
    validator := ValidateIntRange(1, 100, "port")
    err := validator("80")
    assert.NoError(t, err)
    
    err = validator("invalid")
    assert.Error(t, err)
}

// Test multi-choice error handling
func TestArrayLengthValidation(t *testing.T) {
    items := []string{"item1", "item2"}
    defaults := []bool{true} // Mismatched length
    // This would trigger validation error in GetMultiChoice
}
```

The tests provide comprehensive coverage of edge cases, input validation, and error conditions while working around the limitations of testing interactive CLI components that depend on stdin/stdout.