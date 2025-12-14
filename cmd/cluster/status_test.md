Test file that validates the cluster status command functionality using a standardized testing framework with mocked command execution.

## Key Components

- **TestStatusCommand**: Main test function that validates the status command behavior
- **setupFunc**: Configures test environment with a mock command executor
- **teardownFunc**: Cleans up global state after test execution
- **testutil.TestClusterCommand**: Shared test utility for testing cluster commands

## Usage Example

```go
// Run the status command test
func TestStatusCommand(t *testing.T) {
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    testutil.TestClusterCommand(t, "status", getStatusCmd, setupFunc, teardownFunc)
}
```

The test follows a setup-execute-teardown pattern, using a mock executor to simulate command execution without running actual cluster operations. This ensures consistent, isolated testing of the status command's logic and output handling.