<!-- source-hash: 1305620e2607ea5f908fa0cc1f813d97 -->
This file contains comprehensive unit tests for the SystemService configuration component, validating system initialization, log directory management, and error handling scenarios.

## Key Components

- **TestNewSystemService** - Tests basic service instantiation with default log directory
- **TestNewSystemServiceWithOptions** - Validates custom log directory configuration
- **TestSystemService_Initialize** - Comprehensive initialization testing with multiple directory scenarios
- **TestSystemService_GetLogDirectory** - Verifies log directory path retrieval
- **TestSystemService_InitializeErrorHandling** - Tests error handling for invalid directory paths
- **TestSystemService_MultipleInitialize** - Ensures multiple initialization calls are safe

## Usage Example

```go
// Run all tests
go test -v ./config

// Run specific test
go test -run TestNewSystemService ./config

// Test with coverage
go test -cover ./config

// Example test output validation
service := NewSystemService()
logDir := service.GetLogDirectory()
// logDir should be: /tmp/openframe-deployment-logs
```

The tests cover normal operations, edge cases like permission errors, and ensure the service handles directory creation robustly across different scenarios including nested directories and multiple initialization attempts.