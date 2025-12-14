Test file for the cluster cleanup command that verifies the cleanup functionality works correctly with mocked command execution.

## Key Components

- **`TestCleanupCommand`** - Main test function that validates the cleanup command behavior
- **`setupFunc`** - Test setup that configures a mock executor for command testing
- **`teardownFunc`** - Test teardown that resets global flags after test execution
- **`init()`** - Initializes test mode for the package

## Usage Example

```go
// Run the cleanup command test
func TestCleanupCommand(t *testing.T) {
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    testutil.TestClusterCommand(t, "cleanup", getCleanupCmd, setupFunc, teardownFunc)
}
```

The test uses the `testutil.TestClusterCommand` helper to validate that the cleanup command can be properly instantiated and executed with mocked dependencies, ensuring the command logic works without requiring actual cluster resources.