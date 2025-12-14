This file contains unit tests for the cluster command functionality, verifying the root cluster command behavior using the project's testing utilities.

## Key Components

- `TestClusterRootCommand` - Test function that validates the root cluster command execution
- `init()` - Initializes test mode using the testutil package
- `testutil.TestClusterCommand` - Utility function for testing cluster commands

## Usage Example

```go
// Run the test
go test -v ./cluster

// The test verifies that the cluster command can be executed
// without any additional setup or teardown requirements
func TestClusterRootCommand(t *testing.T) {
    testutil.TestClusterCommand(t, "cluster", GetClusterCmd, nil, nil)
}
```

The test uses the `testutil.TestClusterCommand` helper with the command name "cluster", the `GetClusterCmd` function, and no setup/teardown functions (both nil parameters), ensuring the basic cluster command functionality works correctly.