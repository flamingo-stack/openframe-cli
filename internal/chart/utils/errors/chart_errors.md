<!-- source-hash: e5cd93dc7ed2bb8bdde0de204e805c46 -->
Provides comprehensive error handling for chart operations with enhanced context, retry logic, and specialized error types for different failure scenarios.

## Key Components

**Core Types:**
- `ChartError` - Base error type with operation context, cluster info, and retry capabilities
- `InstallationError` - Installation-specific errors with phase tracking and troubleshooting steps
- `ValidationError` - Field validation errors with constraint details
- `ConfigurationError` - Configuration file errors with missing key tracking
- `SkippedInstallationError` - Represents intentionally skipped installations

**Predefined Errors:**
- Common chart errors: `ErrChartNotFound`, `ErrChartAlreadyInstalled`, `ErrClusterNotReady`
- Infrastructure errors: `ErrHelmNotAvailable`, `ErrKubectlNotAvailable`, `ErrInsufficientResources`

**Utility Functions:**
- `IsTimeout()`, `IsRecoverable()`, `GetRetryDelay()` - Error classification helpers
- `WrapAsChartError()`, `CombineErrors()` - Error manipulation utilities

## Usage Example

```go
// Create a recoverable chart error
err := NewRecoverableChartError("install", "nginx", 
    ErrClusterNotReady, 30*time.Second).
    WithCluster("production")

// Handle installation error with troubleshooting
installErr := NewInstallationError("prometheus", "deployment", err).
    WithSuggestions([]string{"Check storage class availability"})

// Validate configuration
validationErr := NewValidationError("memory", "256Mi", "minimum 512Mi required")

// Check error types
if IsRecoverable(err) {
    delay := GetRetryDelay(err)
    time.Sleep(delay)
}

// Combine multiple errors
combinedErr := CombineErrors([]error{installErr, validationErr})
```