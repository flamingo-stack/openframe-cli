<!-- source-hash: e0be7fbde34a7a6be9a8be9b5b555068 -->
Test file for the installer service providing comprehensive unit testing for chart installation workflows. Tests the `Installer` service using mock implementations of ArgoCD and App-of-Apps services.

## Key Components

**Mock Services:**
- `MockArgoCDService` - Mock implementation of ArgoCDService interface
- `MockAppOfAppsService` - Mock implementation of AppOfAppsService interface

**Test Functions:**
- `TestInstaller_InstallCharts` - Main installation workflow testing with various scenarios
- `TestInstaller_InstallCharts_RecoverableError` - Tests recoverable error handling
- `TestInstaller_InstallCharts_NoWaitWithoutAppOfApps` - Verifies conditional application waiting
- `TestInstaller_InstallCharts_ErrorTypes` - Tests different error type wrapping and handling

## Usage Example

```go
// Create mock services for testing
mockArgoCD := new(MockArgoCDService)
mockAppOfApps := new(MockAppOfAppsService)

// Setup mock expectations
mockArgoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
mockAppOfApps.On("Install", mock.Anything, mock.Anything).Return(nil)

// Create installer with mocks
installer := &Installer{
    argoCDService:    mockArgoCD,
    appOfAppsService: mockAppOfApps,
}

// Test configuration
config := config.ChartInstallConfig{
    ClusterName: "test-cluster",
    AppOfApps: &models.AppOfAppsConfig{
        GitHubRepo: "owner/repo",
    },
}

// Execute and verify
err := installer.InstallCharts(config)
assert.NoError(t, err)
mockArgoCD.AssertExpectations(t)
```