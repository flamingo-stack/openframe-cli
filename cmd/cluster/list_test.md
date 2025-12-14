Test file for the cluster list command functionality, providing unit tests to verify the command executes correctly with mock dependencies.

## Key Components

- **TestListCommand**: Main test function that validates the cluster list command behavior
- **setupFunc**: Configures test environment with mock executor for isolated testing
- **teardownFunc**: Cleans up global state after test execution
- **testutil.TestClusterCommand**: Utility function that handles common cluster command testing patterns

## Usage Example

```go
// Run the test
go test ./internal/cluster -run TestListCommand

// The test automatically:
// 1. Sets up mock executor via setupFunc
// 2. Tests the list command through testutil.TestClusterCommand
// 3. Cleans up state via teardownFunc

// Example of how the test validates command execution:
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