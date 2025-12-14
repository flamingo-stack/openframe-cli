<!-- source-hash: d61eb29b5f30e922d8ce5dc3bf177026 -->
Provides standardized error handling and custom error types for CLI applications, with formatted output using pterm and support for different error categories.

## Key Components

**Custom Error Types:**
- `ValidationError` - Field validation failures with optional values
- `CommandError` - Command execution failures with command details
- `BranchNotFoundError` - Git branch not found errors
- `AlreadyHandledError` - Wrapper for errors already displayed to users

**Error Handler:**
- `ErrorHandler` - Main error processing with verbose mode support
- `HandleError()` - Routes errors to specific handlers based on type
- `HandleGlobalError()` - Global entry point for CLI command error handling

**Utility Functions:**
- `CreateValidationError()`, `CreateCommandError()` - Error constructors
- `IsValidationError()`, `IsCommandError()` - Error type checking

## Usage Example

```go
// Create and handle validation errors
err := CreateValidationError("username", "invalid@user", "contains invalid characters")
handler := NewErrorHandler(true)
handler.HandleError(err)

// Global error handling in CLI commands
func myCommand(cmd *cobra.Command, args []string) error {
    if err := someOperation(); err != nil {
        return HandleGlobalError(err, verbose)
    }
    return nil
}

// Check error types
if IsValidationError(err) {
    // Handle validation-specific logic
}
```