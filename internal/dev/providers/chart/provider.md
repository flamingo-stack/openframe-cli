<!-- source-hash: 4c1e3d50387d14ab5129ac136f83959d -->
This file provides chart installation operations specifically tailored for development environments using Helm and ArgoCD.

## Key Components

- **Provider**: Main struct that manages chart installations with configurable verbosity and command execution
- **NewProvider()**: Constructor function that creates a new Provider instance
- **InstallCharts()**: Installs charts on a cluster with custom Helm values
- **InstallChartsWithContext()**: Context-aware version supporting cancellation and aggressive cleanup
- **PrepareDevHelmValues()**: Prepares development-specific Helm values files
- **createDevHelmValuesFile()**: Creates Helm values with development settings (autoSync disabled)

## Usage Example

```go
// Create a new chart provider
executor := executor.NewCommandExecutor()
provider := NewProvider(executor, true) // verbose mode

// Install charts for development
err := provider.InstallCharts("dev-cluster", "custom-values.yaml")
if err != nil {
    log.Fatal(err)
}

// Or with context for cancellation support
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err = provider.InstallChartsWithContext(ctx, "dev-cluster", "helm-values.yaml")
if err != nil {
    log.Fatal(err)
}

// Prepare development-specific values
valuesFile, err := provider.PrepareDevHelmValues("base-values.yaml")
if err != nil {
    log.Fatal(err)
}
```

The provider automatically disables ArgoCD's autoSync feature for Skaffold development workflows and handles aggressive process cleanup when operations are cancelled.