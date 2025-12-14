<!-- source-hash: 73400ace79f0d5d45e91312b6fee4c67 -->
Test file for the cluster creation command that validates the create command functionality using a mock executor.

## Key Components

- **TestCreateCommand**: Main test function that validates the cluster create command behavior
- **setupFunc**: Configures test environment with a mock executor for isolated testing
- **teardownFunc**: Resets global state after test execution
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

    // Execute test with setup/teardown
    testutil.TestClusterCommand(t, "create", getCreateCmd, setupFunc, teardownFunc)
}
```

The test uses the testutil framework to systematically test the cluster create command with proper isolation through setup and teardown functions, ensuring no side effects between test runs.