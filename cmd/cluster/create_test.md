This test file validates the cluster creation command functionality using a standardized testing framework with mock executors.

## Key Components

- **`TestCreateCommand`** - Main test function that validates the cluster create command behavior
- **`setupFunc`** - Configures test environment with mock executor for isolated testing
- **`teardownFunc`** - Cleanup function that resets global flags after test execution
- **`testutil.TestClusterCommand`** - Shared test utility for cluster command validation

## Usage Example

```go
// Run the create command test
func TestCreateCommand(t *testing.T) {
    // Setup mock executor for testing
    setupFunc := func() {
        utils.SetTestExecutor(testutil.NewTestMockExecutor())
    }
    
    // Clean up after test
    teardownFunc := func() {
        utils.ResetGlobalFlags()
    }

    // Execute standardized cluster command test
    testutil.TestClusterCommand(t, "create", getCreateCmd, setupFunc, teardownFunc)
}
```

The test initializes test mode on package load and uses the common testing pattern to ensure the create command works correctly in isolation without affecting global state.