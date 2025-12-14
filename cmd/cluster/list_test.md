<!-- source-hash: 589ade71ff035b3542ffbea0bb206499 -->
Test file that validates the cluster list command functionality using mock executors and standardized test utilities.

## Key Components

- **`TestListCommand`** - Main test function that validates the list command behavior
- **`setupFunc`** - Initializes test environment with mock executor
- **`teardownFunc`** - Cleans up global state after test execution
- **Test utilities** - Leverages `testutil.TestClusterCommand` for standardized cluster command testing

## Usage Example

```go
// Run the test
go test -v ./internal/cluster -run TestListCommand

// The test automatically:
// 1. Sets up mock executor via setupFunc
// 2. Tests the list command through testutil.TestClusterCommand
// 3. Cleans up global flags via teardownFunc
```

The test follows a standard pattern for cluster command testing, ensuring the list functionality works correctly in isolation with mocked dependencies.