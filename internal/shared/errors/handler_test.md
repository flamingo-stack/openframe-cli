<!-- source-hash: 06c680c0f9a046c23e21e408a8a499aa -->
Test file that validates error handling functions for user confirmation interrupts and error wrapping functionality. Provides comprehensive test coverage for edge cases and integration scenarios.

## Key Components

- **TestHandleConfirmationError**: Tests the main error handling function that detects interruption signals
- **TestWrapConfirmationError**: Validates error wrapping with contextual information
- **TestHandleConfirmationErrorIntegration**: Integration tests for error detection logic
- **TestErrorMessageFormatting**: Ensures proper error message formatting with context
- **TestNilErrorHandling**: Validates behavior with nil error inputs
- **TestEdgeCases**: Comprehensive testing of edge cases and boundary conditions

## Usage Example

```go
// Run specific test function
go test -run TestHandleConfirmationError

// Run all tests in the package
go test ./errors

// Run tests with verbose output
go test -v ./errors

// Test coverage analysis
go test -cover ./errors

// Run specific edge case test
go test -run TestEdgeCases ./errors
```

The test suite validates that:
- Nil errors return appropriate default values
- "interrupted" errors trigger special handling (simulated exit behavior)
- Non-interrupted errors are properly wrapped with context
- Error message formatting follows expected patterns
- Edge cases like empty strings and partial matches work correctly