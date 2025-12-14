A comprehensive test file for the cluster list command functionality, ensuring proper execution and cleanup in the OpenFrame CLI testing suite.

## Key Components

- **`TestListCommand`** - Primary test function that validates the cluster list command behavior
- **`setupFunc`** - Test setup function that configures a mock executor for isolated testing
- **`teardownFunc`** - Cleanup function that resets global flags after test execution
- **`init`** - Initialization function that enables test mode for the entire package

## Usage Example

```go
// Run the list command test
func TestListCommand(t *testing.T) {
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }
    
    testutil.TestClusterCommand(t, "list", getListCmd, setupFunc, teardownFunc)
}
```

The test leverages the `testutil.TestClusterCommand` helper to execute standardized cluster command testing with proper mocking and cleanup. The mock executor isolates the test from actual system calls, while the teardown ensures clean state between test runs.