This file contains unit tests for the chart command functionality in the OpenFrame CLI tool.

## Key Components

- **TestChartRootCommand**: Tests the basic structure and properties of the chart command
- **init()**: Initializes test mode using testutil package
- **GetChartCmd()**: Function being tested that returns the chart command

## Usage Example

```go
// Run the test
go test ./chart

// Test validates that the chart command:
// - Has correct name "chart"
// - Contains proper descriptions
// - Has a valid RunE function
// - Includes expected help text about Helm chart management

func TestChartRootCommand(t *testing.T) {
    cmd := GetChartCmd()
    assert.Equal(t, "chart", cmd.Name())
    assert.Contains(t, cmd.Short, "Manage Helm charts")
}
```

The test ensures the chart command is properly configured with appropriate metadata, descriptions, and execution handlers for Helm chart lifecycle management operations.