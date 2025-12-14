<!-- source-hash: edec45d072de93f8f5663ef0e3eeccbd -->
Test suite for UI messaging functions that validates error handling and message display functionality without panics.

## Key Components

- **TestShowOperationError**: Tests error message display with troubleshooting tips, including edge cases with empty and nil tip arrays
- **TestShowNoResourcesMessage**: Validates "no resources found" message display with various input combinations
- **TestShowOperationStart**: Tests operation start message display with custom message maps and fallback scenarios
- **TestShowOperationSuccess**: Tests operation success message display with custom message maps and fallback handling

## Usage Example

```go
// Run specific test
go test -run TestShowOperationError

// Run all UI message tests
go test ./ui -v

// Test with coverage
go test -cover ./ui

// Example of what the tests validate:
tips := []TroubleshootingTip{
    {Description: "Check status", Command: "kubectl get pods"},
}
ShowOperationError("deploy", "my-app", errors.New("failed"), tips)

customMessages := map[string]string{
    "cleanup": "Cleanup completed!",
}
ShowOperationSuccess("cleanup", "test-resource", customMessages)
```

The tests focus on panic prevention rather than output validation, ensuring the UI functions handle various input scenarios gracefully including nil values, empty strings, and missing map entries.