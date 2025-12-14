Test file for the bootstrap service package that validates the service initialization and structure without performing full integration testing.

## Key Components

- **TestNewService**: Verifies the `NewService()` constructor returns a valid Service instance
- **TestServiceStructure**: Validates that the service can access required cluster and chart commands and their subcommands
- **TestServiceExecuteMethodExists**: Confirms the Execute method exists on the service
- **TestServiceArgumentHandling**: Tests service behavior with different argument patterns (empty, single cluster name, whitespace handling)
- **TestServiceVerboseFlagHandling**: Validates service compatibility with verbose flag states

## Usage Example

```go
// Run all bootstrap service tests
go test ./bootstrap

// Run specific test
go test -run TestNewService ./bootstrap

// Run with verbose output
go test -v ./bootstrap

// Example of what the tests validate:
service := NewService()
assert.NotNil(t, service)
assert.IsType(t, &Service{}, service)

// Tests verify command structure without execution
clusterCmd := clusterCmd.GetClusterCmd()
assert.NotNil(t, clusterCmd)
```

The tests focus on structural validation and method availability rather than end-to-end execution to avoid complex integration testing scenarios.