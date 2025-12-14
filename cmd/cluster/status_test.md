<!-- source-hash: fde15aafb95ba8ff38adae7793013497 -->
Test file for validating the cluster status command functionality using a standardized testing framework.

## Key Components

- **TestStatusCommand**: Main test function that validates the status command behavior
- **setupFunc**: Initializes test environment with mock executor
- **teardownFunc**: Cleans up global state after test execution
- **testutil.TestClusterCommand**: Shared testing utility for cluster commands

## Usage Example

```go
func TestStatusCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Run standardized cluster command test
    testutil.TestClusterCommand(t, "status", getStatusCmd, setupFunc, teardownFunc)
}
```

The test leverages the `testutil.TestClusterCommand` framework to ensure consistent testing patterns across cluster commands, with proper setup and teardown of mock dependencies.