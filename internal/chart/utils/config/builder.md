<!-- source-hash: 3e5dd4a0f4125b6503718cfbdddff832 -->
This file provides a configuration builder for Helm chart installations, handling branch selection from Helm values files and constructing installation configurations with optional GitHub repository integration.

## Key Components

- **Builder**: Main struct that orchestrates configuration building with a config service and operations UI
- **HelmValues**: YAML structure mapping for Helm values files with OSS and SaaS deployment configurations
- **NewBuilder()**: Factory function to create a new Builder instance
- **getBranchForDeploymentMode()**: Extracts the appropriate Git branch based on deployment mode (OSS, SaaS Tenant, or SaaS Shared)
- **getBranchFromHelmValues()**: Reads branch information from default Helm values file location
- **BuildInstallConfig()**: Constructs standard installation configuration with optional GitHub repo setup
- **BuildInstallConfigWithCustomHelmPath()**: Advanced configuration builder supporting custom Helm values file paths and deployment modes

## Usage Example

```go
// Create a new configuration builder
operationsUI := &chartUI.OperationsUI{}
builder := config.NewBuilder(operationsUI)

// Build basic installation config
config, err := builder.BuildInstallConfig(
    false,           // force
    false,           // dryRun  
    true,            // verbose
    "my-cluster",    // clusterName
    "github.com/my-org/my-repo", // githubRepo
    "main",          // githubBranch
    "/certs",        // certDir
)

// Build config with custom Helm values
config, err = builder.BuildInstallConfigWithCustomHelmPath(
    false, false, true, false, // force, dryRun, verbose, nonInteractive
    "my-cluster",
    "github.com/my-org/my-repo",
    "main",
    "/certs",
    "custom-values.yaml", // helmValuesPath
    "saas-shared",       // deploymentMode
)
```