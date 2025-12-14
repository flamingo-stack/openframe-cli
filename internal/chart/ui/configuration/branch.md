<!-- source-hash: a568737460f171a1db575658b76f9df0 -->
This file implements branch configuration functionality for Git repository deployment settings, allowing users to select or specify custom branches for OSS deployments.

## Key Components

**BranchConfigurator**
- Main struct that handles Git branch configuration logic
- Contains a `HelmValuesModifier` for reading and modifying Helm values

**NewBranchConfigurator(modifier)**
- Constructor function that creates a new `BranchConfigurator` instance

**Configure(config)**
- Main configuration method that prompts users for branch selection
- Skips branch configuration for SaaS deployment modes
- Provides options to keep current branch or specify a custom one
- Updates the configuration with branch changes

## Usage Example

```go
// Create a branch configurator
modifier := &templates.HelmValuesModifier{}
branchConfig := NewBranchConfigurator(modifier)

// Configure branch settings
config := &types.ChartConfiguration{
    DeploymentMode: &types.DeploymentModeOSS,
    ExistingValues: existingHelmValues,
}

err := branchConfig.Configure(config)
if err != nil {
    log.Fatalf("Branch configuration failed: %v", err)
}

// Check if branch was modified
if config.Branch != nil {
    fmt.Printf("New branch selected: %s", *config.Branch)
}
```