<!-- source-hash: 27a0c55c3e55eff214b752ae1de06198 -->
This file contains comprehensive unit tests for namespace management functionality in the intercept package, focusing on Telepresence integration for Kubernetes namespace operations.

## Key Components

- **TestService_GetCurrentNamespace**: Tests the `getCurrentNamespace()` method with various scenarios including successful retrievals, empty responses, and command failures
- **TestService_SwitchNamespace**: Tests the `switchNamespace()` method covering successful switches, partial failures (quit fails but connect succeeds), and connection errors  
- **TestTelepresenceStatus_JSONParsing**: Tests JSON unmarshaling for the `TelepresenceStatus` struct with valid/invalid JSON inputs
- **Mock setup utilities**: Uses `testutil.NewTestMockExecutor()` for simulating command execution scenarios

## Usage Example

```go
// Run specific test
go test -run TestService_GetCurrentNamespace

// Run all namespace tests
go test ./internal/intercept -v

// Example of how the mocked executor is configured in tests
mockExecutor := testutil.NewTestMockExecutor()
mockExecutor.SetResponse("bash", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "production",
})
service := NewService(mockExecutor, false)
namespace, err := service.getCurrentNamespace(context.Background())
```

The tests validate error handling, default namespace fallback behavior, and proper command execution ordering for Telepresence operations.