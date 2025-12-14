<!-- source-hash: b43cd4664389c08445a0c39007514917 -->
This file implements a service for managing ArgoCD installations and operations within the OpenFrame CLI tool. It provides a high-level interface for installing ArgoCD via Helm and monitoring ArgoCD applications.

## Key Components

- **ArgoCD struct**: Service wrapper containing Helm manager, path resolver, and ArgoCD manager
- **NewArgoCD()**: Constructor that creates a new ArgoCD service with a non-verbose executor
- **Install()**: Installs or upgrades ArgoCD using Helm with progress indication
- **WaitForApplications()**: Waits for all ArgoCD applications to reach ready state
- **IsInstalled()**: Checks if ArgoCD is currently installed in the cluster
- **GetStatus()**: Retrieves the current status and information about the ArgoCD installation

## Usage Example

```go
// Create ArgoCD service
helmMgr := helm.NewHelmManager(exec)
pathResolver := config.NewPathResolver()
argoCDService := NewArgoCD(helmMgr, pathResolver, exec)

// Install ArgoCD
config := config.ChartInstallConfig{
    ClusterName: "my-cluster",
    Verbose: true,
}
err := argoCDService.Install(ctx, config)

// Check installation status
installed, err := argoCDService.IsInstalled(ctx)
if installed {
    status, err := argoCDService.GetStatus(ctx)
}

// Wait for applications to be ready
err = argoCDService.WaitForApplications(ctx, config)
```