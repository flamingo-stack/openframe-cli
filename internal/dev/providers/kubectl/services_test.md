<!-- source-hash: 122e2d2873d6bcb6bd89bdad71a72a2e -->
Test suite for kubectl service operations that validates service discovery, retrieval, and validation functionality. Contains comprehensive tests for the kubectl provider's service management capabilities with mock command execution.

## Key Components

- **TestProvider_GetServices** - Tests retrieval of all services in a namespace
- **TestProvider_GetService** - Tests fetching a specific service by name  
- **TestProvider_ValidateService** - Tests service existence validation
- **Mock Setup Functions** - Configure mock kubectl responses for different scenarios
- **Test Data Structures** - Define expected service information and error conditions

## Usage Example

```go
// Example test structure for service operations
func TestServiceOperation(t *testing.T) {
    testutil.InitializeTestMode()
    mockExecutor := testutil.NewTestMockExecutor()
    provider := NewProvider(mockExecutor, false)
    
    // Setup mock kubectl response
    mockExecutor.SetResponse("kubectl get services", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   `{"items": [{"metadata": {"name": "api-service"}}]}`,
    })
    
    // Test the operation
    services, err := provider.GetServices(context.Background(), "default")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 1, len(services))
    assert.Equal(t, "api-service", services[0].Name)
}
```

The tests cover success scenarios with JSON parsing, error handling for missing services, and kubectl command failures with proper mock cleanup between test runs.