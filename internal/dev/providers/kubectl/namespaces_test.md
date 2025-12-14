<!-- source-hash: cd4d7f1ab70bd505b019a05c5f4c1034 -->
Test suite for kubectl namespace management functionality that validates the Provider's ability to retrieve and validate Kubernetes namespaces through mock command execution.

## Key Components

- **TestProvider_GetNamespaces**: Comprehensive test for `GetNamespaces()` method covering multiple scenarios including successful retrieval, empty clusters, command failures, and malformed output handling
- **TestProvider_ValidateNamespace**: Test suite for `ValidateNamespace()` method verifying namespace existence validation and error handling
- **Mock-based testing**: Uses `testutil.NewTestMockExecutor()` to simulate kubectl command responses without requiring actual Kubernetes cluster access

## Usage Example

```go
// Test setup pattern used throughout the file
testutil.InitializeTestMode()
mockExecutor := testutil.NewTestMockExecutor()
provider := NewProvider(mockExecutor, false)

// Setting up mock responses for kubectl commands
mockExecutor.SetResponse("kubectl get namespaces", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "default kube-system openframe",
})

// Testing namespace retrieval
namespaces, err := provider.GetNamespaces(context.Background())
assert.NoError(t, err)
assert.Equal(t, []string{"default", "kube-system", "openframe"}, namespaces)

// Testing namespace validation
err = provider.ValidateNamespace(context.Background(), "openframe")
assert.NoError(t, err)
```

The tests cover edge cases like empty output, whitespace handling, connection failures, and non-existent namespaces to ensure robust kubectl integration.