This test file validates the chart command's basic structure and functionality within the OpenFrame CLI application.

## Key Components

- **TestChartRootCommand**: Main test function that verifies the chart command structure, descriptions, and help content
- **testutil.InitializeTestMode()**: Initialization function called to set up test environment

## Usage Example

```go
// Run the chart command tests
func TestChartRootCommand(t *testing.T) {
    cmd := GetChartCmd()
    
    // Verify command structure
    assert.Equal(t, "chart", cmd.Name())
    assert.NotEmpty(t, cmd.Short)
    assert.Contains(t, cmd.Short, "Manage Helm charts")
}

// Initialize test environment before running tests
func init() {
    testutil.InitializeTestMode()
}
```

The test ensures the chart command has proper naming, descriptions mentioning "Manage Helm charts" and "chart lifecycle management", and includes a valid RunE function for command execution.