This file contains unit tests for the cluster package's root command functionality, ensuring the main cluster command works correctly in isolation.

## Key Components

- **`init()`** - Initializes test mode using the testutil package
- **`TestClusterRootCommand()`** - Tests the root cluster command without any subcommands or additional setup

## Usage Example

```go
// Run the cluster command tests
func TestClusterRootCommand(t *testing.T) {
    // Tests the basic cluster command functionality
    testutil.TestClusterCommand(t, "cluster", GetClusterCmd, nil, nil)
}

// The test leverages testutil for standardized command testing
// with no additional setup or teardown functions needed
```

The test uses the `testutil.TestClusterCommand` helper to validate that the `GetClusterCmd` function returns a properly configured root command that can be executed without errors.