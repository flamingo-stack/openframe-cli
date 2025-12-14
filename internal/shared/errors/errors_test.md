<!-- source-hash: 0bcf456aac2eb6a80794da7d1c831d1a -->
This test file provides comprehensive test coverage for a custom error handling package, focusing on validation errors and command execution errors.

## Key Components

- **ValidationError Tests**: Verifies error message formatting for field validation failures with various scenarios (with/without values, empty fields)
- **CommandError Tests**: Tests command execution error formatting and unwrapping functionality 
- **ErrorHandler Tests**: Validates error handler creation and error processing for different error types
- **Helper Functions**: Tests utility functions like `CreateValidationError`, `CreateCommandError`, `IsValidationError`, `IsCommandError`
- **Edge Cases & Benchmarks**: Covers special characters, Unicode, empty values, and performance testing

## Usage Example

```go
// Test ValidationError formatting
func TestValidationError_Error(t *testing.T) {
    err := &ValidationError{
        Field:   "name",
        Value:   "invalid-name", 
        Message: "must contain only letters",
    }
    expected := "validation failed for name 'invalid-name': must contain only letters"
    assert.Equal(t, expected, err.Error())
}

// Test CommandError with unwrapping
func TestCommandError_Unwrap(t *testing.T) {
    originalErr := errors.New("original error")
    cmdErr := &CommandError{
        Command: "kubectl",
        Args:    []string{"get", "pods"},
        Err:     originalErr,
    }
    assert.True(t, errors.Is(cmdErr, originalErr))
}

// Test error handler creation
func TestNewErrorHandler(t *testing.T) {
    handler := NewErrorHandler(true) // verbose mode
    assert.NotNil(t, handler)
}
```