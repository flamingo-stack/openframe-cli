<!-- source-hash: 7558b1e18539949b56287cc28ec99f2d -->
Orchestrates the complete Helm chart installation process using ArgoCD and app-of-apps pattern. Handles sequential installation of ArgoCD followed by app-of-apps configuration with proper error handling and context support.

## Key Components

- **`Installer`** - Main service struct that coordinates chart installations
- **`InstallCharts()`** - Simplified installation method using background context
- **`InstallChartsWithContext()`** - Full installation method with custom context support
- **Dependencies**: `ArgoCDService` and `AppOfAppsService` for handling specific installation steps

## Usage Example

```go
// Create installer with required services
installer := &Installer{
    argoCDService:    argoCDSvc,
    appOfAppsService: appOfAppsSvc,
}

// Install charts with default context
config := config.ChartInstallConfig{
    ClusterName: "my-cluster",
    // ... other config fields
}
err := installer.InstallCharts(config)

// Or install with custom context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()
err = installer.InstallChartsWithContext(ctx, config)
```

The installer follows a two-phase approach: first installing ArgoCD, then optionally installing app-of-apps patterns with application readiness verification.