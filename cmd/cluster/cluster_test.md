<!-- source-hash: 94a1927783f5cb767318305eb7ab606c -->
This file contains unit tests for the cluster command functionality, verifying that the root cluster command works correctly without requiring any setup.

## Key Components

- **`init()`** - Initializes test mode using the testutil package
- **`TestClusterRootCommand()`** - Tests the root cluster command execution
- **Test utilities** - Uses `testutil.TestClusterCommand` for standardized command testing

## Usage Example

```go
// Run the cluster command tests
func TestClusterRootCommand(t *testing.T) {
    // Tests the basic cluster command with no subcommands
    testutil.TestClusterCommand(t, "cluster", GetClusterCmd, nil, nil)
}
```

The test validates that the cluster root command can be executed successfully without any additional configuration or setup requirements. It leverages the testutil framework to ensure consistent testing patterns across the CLI application.