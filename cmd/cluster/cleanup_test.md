<!-- source-hash: 3e45e22e584753e99e05d1da22fb5e7a -->
Test file for the cluster cleanup command functionality, verifying the command's behavior using mock execution and test utilities.

## Key Components

- `TestCleanupCommand`: Main test function that validates the cleanup command behavior
- `setupFunc`: Configures test environment with mock executor
- `teardownFunc`: Resets global flags after test completion
- Test initialization with `testutil.InitializeTestMode()`

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
    
    // Execute test using cluster command test utility
    testutil.TestClusterCommand(t, "cleanup", getCleanupCmd, setupFunc, teardownFunc)
}
```

This test ensures the cleanup command properly integrates with the cluster command framework and handles mock execution scenarios correctly.