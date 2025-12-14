Test file for the cluster cleanup command functionality, ensuring proper cleanup operations work correctly in a controlled test environment.

## Key Components

- **`TestCleanupCommand`** - Main test function that validates the cleanup command behavior
- **`init()`** - Initializes test mode for the entire package
- **Setup/teardown functions** - Configure mock executor and reset global state between tests

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

The test uses a mock executor to simulate cleanup operations without affecting the actual system, and includes proper setup/teardown to ensure test isolation. The `testutil.TestClusterCommand` helper function provides standardized testing patterns for cluster commands.