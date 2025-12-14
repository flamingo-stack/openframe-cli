Test file for the cluster list command functionality, verifying the list command behavior using mock executors and test utilities.

## Key Components

- **TestListCommand**: Main test function that validates the cluster list command
- **setupFunc**: Initializes test environment with mock executor
- **teardownFunc**: Cleans up global flags after test execution
- **init()**: Initializes test mode for the package

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

The test uses the standard Go testing pattern with setup/teardown functions to ensure proper isolation. It leverages the `testutil.TestClusterCommand` helper to run comprehensive tests against the list command implementation, using mock executors to avoid external dependencies during testing.