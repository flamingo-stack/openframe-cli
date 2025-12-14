<!-- source-hash: c582a5747fef56767981f5b1602b0463 -->
This test file validates the DisplayService component responsible for providing visual feedback during Helm chart installation operations, including progress indicators, success/error messages, and dry run results.

## Key Components

- **TestNewDisplayService**: Validates DisplayService instantiation
- **TestDisplayService_ShowInstallProgress**: Tests progress display functionality
- **TestDisplayService_ShowInstallSuccess**: Tests success message display with chart information
- **TestDisplayService_ShowInstallError**: Tests error message display
- **TestDisplayService_ShowPreInstallCheck**: Tests pre-installation check messaging
- **TestDisplayService_ShowDryRunResults**: Tests dry run output formatting (multiple scenarios)
- **TestDisplayService_getChartDisplayName**: Tests chart type name conversion

## Usage Example

```go
func TestDisplayServiceUsage(t *testing.T) {
    service := NewDisplayService()
    
    // Test progress display
    service.ShowInstallProgress(models.ChartTypeArgoCD, "Installing...")
    
    // Test dry run output capture
    var buf bytes.Buffer
    results := []string{"Would install ArgoCD v8.2.7"}
    service.ShowDryRunResults(&buf, results)
    
    // Verify output was written to buffer
    assert.Contains(t, buf.String(), "Would install ArgoCD v8.2.7")
    
    // Test chart info display
    chartInfo := models.ChartInfo{
        Name: "test-chart",
        Namespace: "test-namespace",
    }
    service.ShowInstallSuccess(models.ChartTypeArgoCD, chartInfo)
}
```