This test file validates the functionality of the chart command module, ensuring proper command structure and metadata.

## Key Components

- **TestChartRootCommand**: Main test function that validates the chart command's basic properties and structure
- **init()**: Initializes test mode using the testutil package
- **GetChartCmd()**: Function under test that returns the chart command instance

## Usage Example

```go
// Running the test
func TestChartRootCommand(t *testing.T) {
    cmd := GetChartCmd()
    
    // Validate command properties
    assert.Equal(t, "chart", cmd.Name())
    assert.NotEmpty(t, cmd.Short)
    assert.NotNil(t, cmd.RunE)
    
    // Verify command descriptions contain expected content
    assert.Contains(t, cmd.Short, "Manage Helm charts")
    assert.Contains(t, cmd.Long, "chart lifecycle management")
}
```

The test ensures the chart command has proper naming, descriptions, and a RunE function for execution. It verifies that help text contains appropriate keywords related to Helm chart management functionality.