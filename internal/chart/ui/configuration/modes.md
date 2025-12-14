<!-- source-hash: fb139504ff3191d94b21ba99969a2e9b -->
This file implements user interface logic for selecting deployment and configuration modes in the OpenFrame CLI configuration wizard.

## Key Components

- **`showDeploymentModeSelection()`** - Interactive prompt for selecting deployment type (OSS, SaaS, or SaaS Shared)
- **`showConfigurationModeSelection()`** - Prompt for choosing between default or interactive configuration
- **`configureWithDefaults()`** - Creates a default configuration without user interaction
- **`configureInteractive()`** - Runs the full interactive configuration wizard

## Usage Example

```go
wizard := &ConfigurationWizard{}

// Show deployment mode selection
deploymentMode, err := wizard.showDeploymentModeSelection()
if err != nil {
    return err
}

// Show configuration mode selection
configMode, err := wizard.showConfigurationModeSelection()
if err != nil {
    return err
}

// Configure based on selected mode
var config *types.ChartConfiguration
if configMode == "default" {
    config, err = wizard.configureWithDefaults(deploymentMode)
} else {
    config, err = wizard.configureInteractive(deploymentMode)
}
```

The file uses `promptui` for interactive selection menus with styled templates and handles three deployment modes: OSS (default), SaaS, and SaaS Shared. Both configuration paths load base values, set the deployment mode, and create temporary values files for Helm chart deployment.