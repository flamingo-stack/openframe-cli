Test file for the cluster delete command functionality, validating the delete command behavior using mocked executors.

## Key Components

- **TestDeleteCommand**: Main test function that validates the cluster delete command
- **setupFunc**: Configures test environment with mock executor
- **teardownFunc**: Cleans up global flags after test execution
- **testutil.TestClusterCommand**: Utility function for testing cluster commands

## Usage Example

```go
// Run the delete command test
func TestDeleteCommand(t *testing.T) {
    // Setup test environment with mock executor
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the test using the testutil framework
    testutil.TestClusterCommand(t, "delete", getDeleteCmd, setupFunc, teardownFunc)
}
```

The test initializes test mode on package load and uses the testutil framework to systematically test the delete command with proper setup and cleanup procedures.