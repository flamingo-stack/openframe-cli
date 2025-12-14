<!-- source-hash: a9ce2ed0b9be7c52149fb3f06c3a499c -->
This file defines the configuration structure for chart installation operations within the OpenFrame CLI tool.

## Key Components

**ChartInstallConfig**
- Main configuration struct containing installation parameters and flags
- Includes cluster targeting, execution modes (dry-run, force, verbose), and UI behavior controls
- Contains optional app-of-apps configuration for GitOps workflows

**HasAppOfApps() method**
- Boolean helper method that validates whether app-of-apps configuration is properly set
- Checks for both non-nil configuration and valid GitHub repository

## Usage Example

```go
import "github.com/flamingo-stack/openframe-cli/internal/config"

// Basic chart installation config
config := &config.ChartInstallConfig{
    ClusterName:    "production",
    Force:          false,
    DryRun:         true,
    Verbose:        true,
    NonInteractive: true,
}

// With app-of-apps configuration
config.AppOfApps = &models.AppOfAppsConfig{
    GitHubRepo: "my-org/my-apps-repo",
}

// Check if app-of-apps is configured
if config.HasAppOfApps() {
    // Proceed with GitOps installation
}
```