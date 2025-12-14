<!-- source-hash: b45f40ccf0ca7bdd9858a92858457eef -->
Unit test file for the intercept service, providing comprehensive test coverage for traffic interception functionality including service creation, input validation, and state management.

## Key Components

- **TestNewService**: Validates proper service initialization with executor, verbosity settings, and internal state
- **TestService_ValidateInputs**: Tests input validation for service names, ports, namespaces, and header formats
- **TestService_ShowInterceptInstructions**: Verifies instruction display functionality doesn't panic
- **TestService_GettersAndSetters**: Tests service state accessors and mutators
- **TestService_StopIntercept**: Validates intercept termination logic and error handling

## Usage Example

```go
func TestMyInterceptFeature(t *testing.T) {
    testutil.InitializeTestMode()
    
    mockExecutor := testutil.NewTestMockExecutor()
    service := NewService(mockExecutor, true)
    
    // Test input validation
    flags := &models.InterceptFlags{
        Port:      8080,
        Namespace: "production",
        Header:    []string{"X-User-ID=123"},
    }
    
    err := service.validateInputs("my-service", flags)
    assert.NoError(t, err)
    
    // Test service state
    assert.False(t, service.IsIntercepting())
    assert.Equal(t, "", service.GetCurrentService())
}
```