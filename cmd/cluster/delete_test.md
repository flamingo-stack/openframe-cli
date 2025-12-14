This file contains unit tests for the cluster delete command functionality, ensuring the delete operation works correctly in a controlled test environment.

## Key Components

- **TestDeleteCommand**: Main test function that validates the delete command behavior
- **setupFunc**: Initializes test environment with mock executor for safe testing
- **teardownFunc**: Cleans up global state after test execution
- **init()**: Sets up test mode for the entire package

## Usage Example

```go
// Run the delete command tests
func TestDeleteCommand(t *testing.T) {
    // Setup mock executor to prevent actual cluster operations
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Clean up after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the cluster delete command test
    testutil.TestClusterCommand(t, "delete", getDeleteCmd, setupFunc, teardownFunc)
}
```

The test uses a mock executor pattern to safely test delete operations without affecting real cluster resources, following the standard setup/teardown testing pattern for reliable and isolated test execution.