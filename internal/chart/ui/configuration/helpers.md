<!-- source-hash: 8890edcd1e97aade97dd4eca786153cf -->
This file provides helper methods for the ConfigurationWizard to manage Helm chart configurations, including loading base values, creating temporary files, and displaying configuration summaries.

## Key Components

- **`loadBaseValues()`** - Loads existing Helm values from the current directory or creates default values, returning a ChartConfiguration struct
- **`createTemporaryValuesFile()`** - Applies configuration changes and creates a temporary Helm values file for installation
- **`ShowConfigurationSummary()`** - Displays a formatted summary of all modified configuration sections with visual indicators

## Usage Example

```go
// Load base configuration
config, err := wizard.loadBaseValues()
if err != nil {
    log.Fatal(err)
}

// Apply changes and create temporary file
err = wizard.createTemporaryValuesFile(config)
if err != nil {
    log.Fatal(err)
}

// Display what was configured
wizard.ShowConfigurationSummary(config)
// Output:
// Configuration Summary:
// ✓ Deployment mode: production
// ✓ Branch updated: main
// ✓ Ingress type updated: nginx
```

These helper methods work together to provide a complete workflow for configuration management, from loading existing values through applying changes to providing user feedback about modifications made.