<!-- source-hash: 8bf6b795b8a87b1f5b34dc136fb26003 -->
This file contains comprehensive unit tests for the Telepresence provider package, testing all major functionality for setting up and managing Kubernetes service intercepts using Telepresence.

## Key Components

- **TestNewProvider** - Tests provider initialization with mock executor and verbose settings
- **TestProvider_CheckTelepresenceInstallation** - Verifies Telepresence installation detection
- **TestProvider_ConnectToCluster** - Tests cluster connection functionality
- **TestProvider_CreateIntercept** - Tests service intercept creation with various configurations
- **TestProvider_CreateInterceptWithEnvFile** - Tests intercept creation with environment file support
- **TestProvider_TeardownIntercept** - Tests intercept cleanup functionality
- **TestProvider_Disconnect** - Tests Telepresence disconnection
- **TestProvider_ShowInterceptStatus** - Tests intercept status display
- **TestProvider_SetupIntercept_Integration** - End-to-end integration test
- **TestProvider_VerboseLogging** - Tests verbose logging configuration

## Usage Example

```go
func TestMyTelepresenceFeature(t *testing.T) {
    testutil.InitializeTestMode()
    mockExecutor := testutil.NewTestMockExecutor()
    provider := NewProvider(mockExecutor, true)

    // Set up mock responses
    mockExecutor.SetResponse("telepresence version", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   "Telepresence 2.19.1",
    })

    // Test the functionality
    err := provider.checkTelepresenceInstallation(context.Background())
    assert.NoError(t, err)
}
```

The tests use a mock executor pattern to simulate Telepresence command execution and validate both success and failure scenarios.