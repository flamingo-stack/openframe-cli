This test file validates the cluster status command functionality using a standardized testing framework with mock executors.

## Key Components

- **TestStatusCommand**: Main test function that validates the status command behavior
- **setupFunc**: Configures test environment with mock executor for command simulation
- **teardownFunc**: Cleans up global flags after test execution
- **testutil.TestClusterCommand**: Shared testing utility for cluster command validation

## Usage Example

```go
// Run the status command test
func TestStatusCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }
    
    // Execute standardized cluster command test
    testutil.TestClusterCommand(t, "status", getStatusCmd, setupFunc, teardownFunc)
}
```

The test follows a standard pattern of setup → execute → teardown, ensuring the status command works correctly in isolation without affecting other tests or requiring real cluster infrastructure.