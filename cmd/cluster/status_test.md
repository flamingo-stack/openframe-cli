Test file for the cluster status command functionality, containing unit tests to verify the status command behavior in a controlled environment.

## Key Components

- **`TestStatusCommand`** - Main test function that validates the cluster status command execution
- **`setupFunc`** - Initializes test environment with a mock executor for command testing
- **`teardownFunc`** - Cleans up global state after test completion
- **`init()`** - Initializes the testing framework in test mode

## Usage Example

```go
// Run the status command test
func TestStatusCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Clean up after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the cluster command test
    testutil.TestClusterCommand(t, "status", getStatusCmd, setupFunc, teardownFunc)
}
```

The test uses a mock executor to simulate command execution without actually running cluster operations, ensuring reliable and isolated testing of the status command functionality.