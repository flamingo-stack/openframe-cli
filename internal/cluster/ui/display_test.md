<!-- source-hash: 049351306b5cc20251b9a9c45660cb53 -->
This file contains comprehensive test coverage for UI display utilities used in a Kubernetes cluster management application. It validates status color mapping, table rendering functions, and various UI formatting components.

## Key Components

- **TestGetStatusColor**: Tests status-to-color mapping (green for running/ready, yellow for stopped/pending, red for errors, gray for unknown)
- **TestRenderTableWithFallback**: Tests general table rendering with fallback handling for various column counts
- **TestRenderOverviewTable**: Tests cluster overview table rendering for property-value pairs
- **TestRenderNodeTable**: Tests Kubernetes node table rendering with node-specific columns
- **TestShowSuccessBox**: Tests success notification display functionality
- **TestFormatAge**: Tests time duration formatting (days, hours, minutes, seconds)
- **TestShowClusterCreationNextSteps**: Tests post-creation guidance display
- **TestShowNoResourcesMessage**: Tests empty state messaging for resource lists

## Usage Example

```go
func TestStatusColorMapping(t *testing.T) {
    // Test status color function
    colorFunc := GetStatusColor("running")
    result := colorFunc("cluster-1")
    expected := pterm.Green("cluster-1")
    assert.Equal(t, expected, result)
    
    // Test age formatting
    pastTime := time.Now().Add(-2 * time.Hour)
    age := FormatAge(pastTime)
    assert.Equal(t, "2h", age)
    
    // Test table rendering
    data := pterm.TableData{
        {"Name", "Status", "Age"},
        {"cluster-1", "running", "2h"},
    }
    err := RenderTableWithFallback(data, true)
    assert.NoError(t, err)
}
```