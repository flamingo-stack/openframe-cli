<!-- source-hash: 0c83714ca8a99753c694e5587acc78ff -->
This file contains comprehensive unit tests for the configuration wizard component, testing Helm values file manipulation, deployment mode configurations, and integration workflows for chart configurations.

## Key Components

- **TestConfigurationWizard_ConfigureWithDefaults_OSS**: Tests basic OSS configuration setup and temporary file creation
- **TestConfigurationWizard_ConfigureWithExistingFile**: Validates loading and parsing of existing Helm values files
- **TestConfigurationWizard_Integration_LoadAndApply**: End-to-end integration test for configuration loading, modification, and persistence
- **TestConfigurationWizard_DeploymentModes**: Unit tests for deployment mode type validation (OSS/SaaS)
- **TestConfigurationWizard_LoadBaseValues**: Tests base configuration loading functionality

## Usage Example

```go
// Run the full test suite
go test ./configuration -v

// Run specific test cases
go test ./configuration -run TestConfigurationWizard_ConfigureWithDefaults_OSS

// Test deployment mode validation
func TestCustomDeploymentMode(t *testing.T) {
    mode := types.DeploymentModeOSS
    assert.Equal(t, "oss", string(mode))
}

// Test configuration modification workflow
modifier := templates.NewHelmValuesModifier()
values, err := modifier.LoadOrCreateBaseValues()
require.NoError(t, err)

config := &types.ChartConfiguration{
    DeploymentMode: &types.DeploymentModeOSS,
    // ... other config fields
}
err = modifier.ApplyConfiguration(values, config)
assert.NoError(t, err)
```