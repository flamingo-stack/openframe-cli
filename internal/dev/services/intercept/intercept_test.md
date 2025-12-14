<!-- source-hash: e8b8c436e9d0ee37274cb56b31a7df95 -->
This file contains comprehensive unit tests for the intercept service functionality, testing Telepresence intercept command generation and execution with various configuration options.

## Key Components

- **TestService_CreateIntercept**: Main test suite covering intercept creation with different flag combinations including port mapping, headers, environment files, global intercepts, and error handling
- **TestService_GetRemotePortName**: Tests the logic for determining remote port names (custom names vs. port numbers)
- **TestIntercept_CommandConstruction**: Validates complete command construction with all available flags enabled

## Usage Example

```go
func TestCustomIntercept(t *testing.T) {
    testutil.InitializeTestMode()
    mockExecutor := testutil.NewTestMockExecutor()
    service := NewService(mockExecutor, false)
    
    flags := &models.InterceptFlags{
        Port:      8080,
        Namespace: "default",
        Header:    []string{"X-User=admin"},
        Global:    true,
    }
    
    mockExecutor.SetResponse("telepresence intercept", &executor.CommandResult{ExitCode: 0})
    
    err := service.createIntercept(context.Background(), "my-service", flags)
    assert.NoError(t, err)
    
    commands := mockExecutor.GetExecutedCommands()
    assert.Contains(t, commands[0], "telepresence intercept my-service")
}
```

The tests verify command argument construction, mock executor interactions, and proper error handling for failed intercept operations.