<!-- source-hash: 75bc2a38056eaf2ec7ff2832293b1ca8 -->
This test file validates the cleanup and signal handling functionality of the intercept service, ensuring proper teardown of Telepresence connections and restoration of original namespace context.

## Key Components

- **TestService_SetupCleanupHandler** - Tests the initialization of signal handling and cleanup mechanisms
- **TestService_Cleanup** - Comprehensive test suite covering various cleanup scenarios including failure cases
- **TestSignalHandling** - Validates proper handling of SIGINT and SIGTERM signals
- **TestCleanupState_Management** - Tests state management during intercept lifecycle

## Usage Example

```go
// Testing cleanup handler setup
testutil.InitializeTestMode()
mockExecutor := testutil.NewTestMockExecutor()
service := NewService(mockExecutor, false)

service.setupCleanupHandler("test-service")
assert.NotNil(t, service.signalChannel)

// Testing cleanup with mocked commands
mockExecutor.SetResponse("telepresence leave", &executor.CommandResult{ExitCode: 0})
mockExecutor.SetResponse("telepresence quit", &executor.CommandResult{ExitCode: 0})

// Simulate intercept state
service.isIntercepting = true
service.currentService = "api-service"
service.originalNamespace = "default"

// Verify cleanup commands are executed
assert.True(t, mockExecutor.WasCommandExecuted("telepresence leave"))
```

The tests cover edge cases like command failures, namespace restoration logic, and proper signal channel management to ensure robust cleanup behavior.