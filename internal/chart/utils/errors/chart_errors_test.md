<!-- source-hash: b1c1b7c3e45ec98115acc3f3f2c439ad -->
Test file for chart-specific error handling functionality, providing comprehensive coverage for custom error types used in chart operations like installation, validation, and configuration.

## Key Components

- **ChartError Tests**: Tests for basic chart error creation, error message formatting, cluster association, and recovery configuration
- **InstallationError Tests**: Tests for installation-specific errors with phases and troubleshooting steps
- **ValidationError Tests**: Tests for field validation errors with constraint checking
- **ConfigurationError Tests**: Tests for configuration file errors with missing key tracking
- **Utility Function Tests**: Tests for error type checking (`IsTimeout`, `IsRecoverable`), retry delay retrieval, error wrapping, and error combination

## Usage Example

```go
// Test chart error creation and properties
func TestChartErrorBasics() {
    cause := errors.New("helm timeout")
    chartErr := NewChartError("installation", "ArgoCD", cause)
    
    // Verify error properties
    assert.Equal(t, "installation", chartErr.Operation)
    assert.Equal(t, "ArgoCD", chartErr.Component) 
    assert.False(t, chartErr.Recoverable)
}

// Test recoverable error with retry logic
func TestRecoverableError() {
    retryAfter := 30 * time.Second
    err := NewRecoverableChartError("installation", "Helm", cause, retryAfter)
    
    assert.True(t, IsRecoverable(err))
    assert.Equal(t, retryAfter, GetRetryDelay(err))
}
```