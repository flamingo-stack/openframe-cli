<!-- source-hash: 872767f518c4f0deb6aee435ec7c516f -->
Test file for the configuration wizard's summary display functionality, ensuring the `ShowConfigurationSummary` method handles various configuration scenarios without errors.

## Key Components

- **TestConfigurationWizard_ShowConfigurationSummary_NoChanges**: Tests summary display with empty configuration
- **TestConfigurationWizard_ShowConfigurationSummary_WithChanges**: Tests summary display with basic configuration changes (branch, deployment mode, docker, ingress)
- **TestConfigurationWizard_ShowConfigurationSummary_WithNgrokConfig**: Tests summary display with ngrok ingress configuration
- **TestConfigurationWizard_ShowConfigurationSummary_WithSaaSConfig**: Tests summary display with SaaS deployment configuration

## Usage Example

```go
// Run all configuration summary tests
go test -v ./internal/configuration -run TestConfigurationWizard_ShowConfigurationSummary

// Run specific test case
go test -v ./internal/configuration -run TestConfigurationWizard_ShowConfigurationSummary_WithChanges

// Test the configuration wizard summary functionality
wizard := NewConfigurationWizard()
config := &types.ChartConfiguration{
    ModifiedSections: []string{"deployment"},
    // ... other config fields
}
wizard.ShowConfigurationSummary(config) // Should not panic
```