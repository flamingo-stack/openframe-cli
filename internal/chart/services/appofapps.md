<!-- source-hash: 742588bd35d66151ecfff2346a2cac72 -->
Service responsible for managing ArgoCD app-of-apps pattern installations from GitHub repositories. It orchestrates repository cloning, configuration validation, and Helm-based deployment.

## Key Components

- **AppOfApps**: Main service struct managing app-of-apps lifecycle
- **NewAppOfApps()**: Constructor function accepting helm manager, git repository, and path resolver dependencies
- **Install()**: Clones git repository and installs app-of-apps chart using Helm
- **IsInstalled()**: Checks if app-of-apps is currently installed in cluster
- **GetStatus()**: Returns detailed status information about the installation

## Usage Example

```go
helmManager := helm.NewHelmManager()
gitRepo := git.NewRepository()
pathResolver := config.NewPathResolver()

appOfApps := NewAppOfApps(helmManager, gitRepo, pathResolver)

// Install app-of-apps
config := config.ChartInstallConfig{
    ClusterName: "production",
    AppOfApps: &config.AppOfAppsConfig{
        GitHubRepo:   "https://github.com/my-org/my-apps",
        GitHubBranch: "main",
        ChartPath:    "charts/app-of-apps",
    },
}

err := appOfApps.Install(ctx, config)
if err != nil {
    log.Fatal(err)
}

// Check installation status
installed, err := appOfApps.IsInstalled(ctx, "argocd")
```