<!-- source-hash: 73f1477a1809a2469ab88fcbfef72887 -->
Test suite for the ChartInfo model and ChartType constants, verifying struct initialization, field assignments, and type conversions.

## Key Components

- **TestChartInfo_DefaultValues**: Validates that ChartInfo struct fields are empty by default
- **TestChartInfo_WithValues**: Tests ChartInfo struct with populated values
- **TestChartType_Constants**: Verifies ChartType constant definitions
- **TestChartType_StringConversion**: Tests string conversion of ChartType values

## Usage Example

```go
// Run specific test
go test -run TestChartInfo_WithValues

// Run all tests in the package
go test ./models

// Example of the tested ChartInfo struct usage
info := &ChartInfo{
    Name:       "argocd",
    Namespace:  "argocd", 
    Status:     "deployed",
    Version:    "1.0.0",
    AppVersion: "v2.8.0",
}

// Example of tested ChartType constants
chartType := ChartTypeArgoCD        // equals "argocd"
appOfAppsType := ChartTypeAppOfApps // equals "app-of-apps"
```

The tests ensure the ChartInfo model behaves correctly with both empty initialization and explicit value assignment, while also validating that ChartType constants maintain their expected string representations.