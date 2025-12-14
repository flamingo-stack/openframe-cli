<!-- source-hash: a5681f5654befa7758cc1346ff11d05e -->
Test file for the cluster delete command functionality, providing unit tests to verify the delete command works correctly with mocked executors.

## Key Components

- **TestDeleteCommand**: Main test function that validates the cluster delete command behavior
- **setupFunc**: Configures test environment with a mock executor for safe testing
- **teardownFunc**: Cleans up global flags after test execution
- **init()**: Initializes test mode for the package

## Usage Example

```go
// Run the delete command test
func TestDeleteCommand(t *testing.T) {
    // Setup mock executor to avoid actual cluster operations
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the test with proper setup/teardown
    testutil.TestClusterCommand(t, "delete", getDeleteCmd, setupFunc, teardownFunc)
}
```

The test uses the standard testing framework with mock executors to safely test cluster deletion functionality without affecting real cluster resources.