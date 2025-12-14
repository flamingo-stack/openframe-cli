Test file for the cluster delete command functionality, verifying the command's behavior using mock execution and standard test utilities.

## Key Components

- **TestDeleteCommand**: Main test function that validates the delete command implementation
- **setupFunc**: Configures test environment with mock executor for command testing
- **teardownFunc**: Cleans up global state by resetting flags after test execution
- **init()**: Initializes test mode for the testing environment

## Usage Example

```go
// Run the delete command test
func TestDeleteCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the cluster delete command test
    testutil.TestClusterCommand(t, "delete", getDeleteCmd, setupFunc, teardownFunc)
}
```

The test uses a standardized testing pattern with setup/teardown functions to ensure isolated test execution and proper cleanup of global state between test runs.