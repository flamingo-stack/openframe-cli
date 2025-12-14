<!-- source-hash: 8f0bc703300a471d3b7a10fa40c48fd3 -->
This file contains comprehensive unit tests for the kubectl provider implementation, testing Kubernetes context management and connection verification functionality.

## Key Components

- **TestNewProvider** - Tests provider instantiation with mock executor
- **TestProvider_GetCurrentContext** - Tests retrieval of current kubectl context with various scenarios (success, whitespace handling, no context, command failures)
- **TestProvider_CheckConnection** - Tests Kubernetes cluster connection verification using `kubectl cluster-info`
- **TestProvider_SetContext** - Tests switching between kubectl contexts with success and error cases

## Usage Example

```go
// Run specific test
go test -run TestProvider_GetCurrentContext

// Run all provider tests with verbose output
go test -v ./kubectl/

// Test with coverage
go test -cover ./kubectl/
```

Each test function follows the table-driven testing pattern, using mock executors to simulate various kubectl command responses and failure scenarios. The tests verify proper error handling, output parsing, and command execution for kubectl operations.