This file contains unit tests for the cluster creation command functionality in the OpenFrame CLI tool.

## Key Components

- **TestCreateCommand**: Main test function that validates the cluster create command behavior
- **setupFunc**: Anonymous function that configures test environment with mock executor
- **teardownFunc**: Anonymous function that cleans up global flags after testing
- **Test initialization**: `init()` function that prepares the test environment

## Usage Example

```go
// Run the cluster create command test
func TestCreateCommand(t *testing.T) {
    // Setup test environment with mock executor
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test execution
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute the test using the common cluster command test framework
    testutil.TestClusterCommand(t, "create", getCreateCmd, setupFunc, teardownFunc)
}
```

The test leverages the `testutil` package's common testing framework to validate the create command with proper setup and teardown procedures, ensuring isolated test execution with mocked dependencies.