<!-- source-hash: 4d20018f6cca686f172d4907cc850c16 -->
This file provides validation services for Helm chart configurations in non-interactive deployment scenarios. It ensures that required configuration values are present based on the selected deployment mode (OSS, SaaS, or SaaS Shared).

## Key Components

- **ConfigurationValidator**: Main validator struct that validates helm-values.yaml configurations
- **NewConfigurationValidator()**: Factory function to create a new validator instance
- **ValidateConfiguration()**: Primary validation method that routes to mode-specific validators
- **validateOSSConfiguration()**: Validates Open Source deployment requirements
- **validateSaaSConfiguration()**: Validates SaaS deployment requirements including repository passwords and GHCR credentials
- **validateSaaSSharedConfiguration()**: Validates SaaS Shared deployment requirements
- Helper methods: `isDeploymentEnabled()`, `hasPassword()`, `hasGHCRCredentials()`, `hasBranch()`

## Usage Example

```go
package main

import (
    "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
    "github.com/flamingo-stack/openframe-cli/internal/services"
)

func main() {
    // Create validator
    validator := services.NewConfigurationValidator()
    
    // Prepare configuration
    mode := types.DeploymentModeSaaS
    config := &types.ChartConfiguration{
        DeploymentMode: &mode,
        ExistingValues: map[string]interface{}{
            "deployment": map[string]interface{}{
                "saas": map[string]interface{}{
                    "enabled": true,
                    "repository": map[string]interface{}{
                        "password": "secret123",
                    },
                },
            },
            "registry": map[string]interface{}{
                "ghcr": map[string]interface{}{
                    "username": "user",
                    "password": "token",
                },
            },
        },
    }
    
    // Validate configuration
    if err := validator.ValidateConfiguration(config); err != nil {
        panic(err)
    }
}
```