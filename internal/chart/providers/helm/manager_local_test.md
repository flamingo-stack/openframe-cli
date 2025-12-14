<!-- source-hash: 88bb568fb24ca6c95acf1a71b1ba5018 -->
Tests for the Helm manager's local app-of-apps installation functionality. Contains table-driven tests that validate configuration handling, error cases, and successful Helm chart installations with various parameters.

## Key Components

- **TestHelmManager_InstallAppOfAppsFromLocal**: Main test function that validates the `InstallAppOfAppsFromLocal` method
- **Test cases**: Cover nil config, empty chart path, successful installation, and dry-run scenarios
- **MockExecutor**: Used to simulate Helm command execution without actual system calls
- **Configuration validation**: Tests proper handling of required fields like chart path and app-of-apps config

## Usage Example

```go
// Run the test suite
go test -v ./internal/chart/helm

// Run specific test case
go test -run TestHelmManager_InstallAppOfAppsFromLocal -v

// Example of what the test validates:
config := config.ChartInstallConfig{
    AppOfApps: &models.AppOfAppsConfig{
        ChartPath:  "/tmp/chart/manifests/app-of-apps",
        ValuesFile: "/path/to/values.yaml",
        Namespace:  "argocd",
        Timeout:    "60m",
    },
}

manager := NewHelmManager(mockExecutor)
err := manager.InstallAppOfAppsFromLocal(ctx, config, certFile, keyFile)
```

The tests ensure proper error handling for invalid configurations and verify that the correct Helm commands are constructed with appropriate flags for both normal and dry-run installations.