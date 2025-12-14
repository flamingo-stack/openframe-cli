Test file for the cluster create command functionality, verifying command execution and behavior in a controlled testing environment.

## Key Components

- **TestCreateCommand**: Main test function that validates the cluster create command using a standardized testing framework
- **setupFunc**: Test setup function that configures a mock executor for isolated testing
- **teardownFunc**: Test cleanup function that resets global flags after test execution
- **init()**: Initializes test mode for the entire package

## Usage Example

```go
// Run the create command test
func TestCreateCommand(t *testing.T) {
    // Setup mock executor
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Cleanup after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute standardized cluster command test
    testutil.TestClusterCommand(t, "create", getCreateCmd, setupFunc, teardownFunc)
}
```

This test follows the standard pattern for testing cluster commands, using mock executors to simulate command execution without affecting the actual system state.