<!-- source-hash: a100fb454f6edbcb58ce61ecbf74551e -->
Manages Helm operations for chart installations and status checking, primarily focused on ArgoCD deployment and app-of-apps configuration.

## Key Components

- **HelmManager**: Main struct that wraps a command executor for Helm operations
- **NewHelmManager()**: Constructor function that creates a new Helm manager instance
- **IsHelmInstalled()**: Verifies Helm CLI availability in the system
- **IsChartInstalled()**: Checks if a specific Helm release exists in a namespace
- **InstallArgoCD()**: Basic ArgoCD installation with Helm repository management
- **InstallArgoCDWithProgress()**: Enhanced ArgoCD installation with progress indicators and verbose output
- **InstallAppOfAppsFromLocal()**: Installs app-of-apps chart from local filesystem with TLS certificate support
- **GetChartStatus()**: Retrieves status information for deployed Helm releases

## Usage Example

```go
// Create a new Helm manager
helmManager := NewHelmManager(executor.NewCommandExecutor())

// Check if Helm is available
if err := helmManager.IsHelmInstalled(ctx); err != nil {
    return fmt.Errorf("helm not available: %w", err)
}

// Install ArgoCD with progress tracking
config := config.ChartInstallConfig{
    Verbose:        true,
    Silent:        false,
    NonInteractive: false,
    DryRun:        false,
}

err := helmManager.InstallArgoCDWithProgress(ctx, config)
if err != nil {
    return fmt.Errorf("failed to install ArgoCD: %w", err)
}

// Check if a chart is installed
installed, err := helmManager.IsChartInstalled(ctx, "argo-cd", "argocd")
```