<!-- source-hash: 0ccf18d05ec1b05a9074620bd5658f80 -->
This file contains unit tests for the chart command functionality in the OpenFrame CLI, specifically testing the root chart command structure and properties.

## Key Components

- **TestChartRootCommand**: Main test function that validates the chart command's basic structure, descriptions, and functionality
- **init()**: Test initialization function that sets up test mode using testutil

## Usage Example

```go
// Run the chart command tests
go test ./chart

// Example of what the test validates:
cmd := GetChartCmd()
assert.Equal(t, "chart", cmd.Name())
assert.Contains(t, cmd.Short, "Manage Helm charts")
assert.Contains(t, cmd.Long, "chart lifecycle management")
```

The test ensures the chart command has proper naming, descriptions mentioning Helm chart management, and includes a RunE function for execution. It uses the testify assertion library for validation and testutil for test environment setup.