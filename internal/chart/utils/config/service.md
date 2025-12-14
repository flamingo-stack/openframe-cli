<!-- source-hash: e6db28a5152447990abee28cdbfdf036 -->
Provides centralized configuration management for Helm chart operations, including path resolution, certificate handling, and installation configuration building.

## Key Components

- **Service**: Main configuration service that orchestrates path resolution and system configuration
- **NewService()**: Factory function to create a new configuration service instance
- **Initialize()**: Sets up required directories and initializes the configuration system
- **BuildInstallConfig()**: Creates complete installation configuration for chart deployments
- **Path Methods**: `GetCertificateDirectory()`, `GetCertificateFiles()`, `GetHelmValuesFile()`, `GetLogDirectory()` for accessing various file system paths
- **GetDefaultManifestsPath()**: Discovers the default location for chart manifests

## Usage Example

```go
// Create and initialize the configuration service
configService := config.NewService()
if err := configService.Initialize(); err != nil {
    log.Fatal("Failed to initialize config:", err)
}

// Build installation configuration
appConfig := &models.AppOfAppsConfig{
    Name: "my-app",
}
installConfig := configService.BuildInstallConfig(
    true,  // force
    false, // dry-run
    true,  // verbose
    "production-cluster",
    appConfig,
)

// Get certificate paths
certFile, keyFile := configService.GetCertificateFiles()
```