<!-- source-hash: 53137974cb96cd48db62374978c134b6 -->
Provides configuration functionality for SaaS (Software as a Service) deployment settings within the OpenFrame CLI chart configuration system. This file handles both default and interactive configuration modes for SaaS-specific parameters including repository credentials, branch selection, and Docker registry setup.

## Key Components

- **`configureSaaSDefaults`** - Configures SaaS settings in default mode with minimal user prompts
- **`configureSaaSInteractive`** - Full interactive configuration for SaaS deployment settings
- **`configureSaaSBranch`** - Interactive branch selection for SaaS repository
- **`configureOSSBranchForSaaS`** - OSS repository branch configuration in SaaS context
- **`getSaaSBranchFromValues`** - Extracts current SaaS branch from existing Helm values

## Usage Example

```go
// Configure SaaS settings interactively
wizard := &ConfigurationWizard{}
config := &types.ChartConfiguration{
    ExistingValues: existingHelmValues,
}

// Interactive mode - prompts for all settings
err := wizard.configureSaaSInteractive(config)
if err != nil {
    log.Fatal("SaaS configuration failed:", err)
}

// Default mode - uses existing values where possible
err = wizard.configureSaaSDefaults(config)
if err != nil {
    log.Fatal("SaaS defaults configuration failed:", err)
}

// Access configured values
fmt.Printf("SaaS Branch: %s\n", config.SaaSConfig.SaaSBranch)
fmt.Printf("OSS Branch: %s\n", config.SaaSConfig.OSSBranch)
```