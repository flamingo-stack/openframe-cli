<!-- source-hash: 9b463371ac53de2b2dee7daee35757ca -->
Test suite for the chart installation configuration model, validating the behavior of `ChartInstallConfig` struct methods and field initialization.

## Key Components

- **TestChartInstallConfig_HasAppOfApps**: Tests the `HasAppOfApps()` method logic for various AppOfApps configuration states
- **TestChartInstallConfig_DefaultValues**: Validates default zero-value initialization of all config fields
- **TestChartInstallConfig_WithValues**: Tests proper assignment and retrieval of all configuration values

## Usage Example

```go
// Run the specific test for HasAppOfApps method
go test -run TestChartInstallConfig_HasAppOfApps

// Run all tests in the models_test.go file
go test -v ./internal/config

// Test the actual ChartInstallConfig struct based on test patterns
config := &ChartInstallConfig{
    ClusterName: "my-cluster",
    Force:       true,
    AppOfApps:   &models.AppOfAppsConfig{GitHubRepo: "https://github.com/my/repo"},
}

if config.HasAppOfApps() {
    // AppOfApps is properly configured
    fmt.Println("App of Apps configuration detected")
}
```