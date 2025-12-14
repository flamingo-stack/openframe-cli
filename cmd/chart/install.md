Implements the "install" subcommand for the chart module that installs ArgoCD and app-of-apps on Kubernetes clusters. Provides both interactive and non-interactive modes with extensive configuration options.

## Key Components

- **`getInstallCmd()`** - Returns the cobra command for chart installation with usage examples and flag definitions
- **`runInstallCommand()`** - Main command execution handler that processes arguments and delegates to installation service
- **`InstallFlags`** - Struct containing all installation configuration flags (force, dry-run, GitHub settings, deployment mode, etc.)
- **`extractInstallFlags()`** - Extracts and validates command flags with deployment mode validation
- **`addInstallFlags()`** - Configures all available command flags including force, dry-run, GitHub repo/branch, and deployment modes

## Usage Example

```go
// Get the install command for cobra CLI
installCmd := getInstallCmd()

// Example flag extraction in command handler
flags, err := extractInstallFlags(cmd)
if err != nil {
    return err
}

// Installation request setup
req := types.InstallationRequest{
    Args:           args,
    Force:          flags.Force,
    DryRun:         flags.DryRun,
    DeploymentMode: flags.DeploymentMode,
    NonInteractive: flags.NonInteractive,
}

// Execute installation
err = services.InstallChartsWithConfig(req)
```