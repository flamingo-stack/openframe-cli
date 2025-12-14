<!-- source-hash: b9217ebce51596f6e868878062d2c1e9 -->
Manages the chart configuration workflow through a user-friendly wizard interface that guides users through Helm values setup with different deployment and configuration modes.

## Key Components

- **ConfigurationWizard**: Main struct that orchestrates the configuration process through specialized configurators
- **NewConfigurationWizard()**: Factory function that creates a wizard with all necessary configurators
- **ConfigureHelmValues()**: Primary method that runs the full configuration workflow with user prompts
- **ConfigureHelmValuesWithMode()**: Alternative method that skips deployment mode selection when mode is pre-determined

## Usage Example

```go
package main

import (
    "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
    "github.com/flamingo-stack/openframe-cli/internal/chart/ui/configuration"
)

func main() {
    // Create a new configuration wizard
    wizard := configuration.NewConfigurationWizard()
    
    // Run interactive configuration
    config, err := wizard.ConfigureHelmValues()
    if err != nil {
        log.Fatal(err)
    }
    
    // Or configure with predetermined deployment mode
    config, err = wizard.ConfigureHelmValuesWithMode(types.Production)
    if err != nil {
        log.Fatal(err)
    }
}
```

The wizard internally manages branch, Docker, and ingress configuration through specialized configurators and provides both default and interactive configuration paths.