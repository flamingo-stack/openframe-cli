Contains test functions for validating the cluster cleanup command functionality using a standardized testing framework.

## Key Components

- **`init()`** - Initializes test mode for the package
- **`TestCleanupCommand(t *testing.T)`** - Main test function that validates the cleanup command behavior
- **`setupFunc`** - Test setup that configures a mock executor for isolated testing
- **`teardownFunc`** - Test cleanup that resets global flags after test execution

## Usage Example

```go
// Run the cleanup command test
func TestCleanupCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Clean up after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute standardized cluster command test
    testutil.TestClusterCommand(t, "cleanup", getCleanupCmd, setupFunc, teardownFunc)
}
```

The test follows the standard pattern of setup, execution, and teardown to ensure the cleanup command works correctly in isolation without affecting other tests or the system state.