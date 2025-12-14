<!-- source-hash: 3c01e63f10cffc26d1a8a7b8e73e3eed -->
This file implements a comprehensive Helm chart installation service for the OpenFrame CLI, providing both interactive and non-interactive deployment workflows with graceful error handling and cleanup.

## Key Components

**ChartService**
- Core service managing chart operations with dry-run and verbose mode support
- Integrates cluster management, configuration, UI operations, Helm, and Git repositories

**InstallationWorkflow** 
- Orchestrates the complete installation process with context-aware cancellation
- Supports three modes: fully interactive, partial non-interactive, and CI/CD non-interactive
- Handles signal interruption (CTRL-C) with proper cleanup

**Key Methods**
- `Install()` / `InstallWithContext()` - Main entry points for chart installation
- `ExecuteWithContext()` - Core workflow execution with cancellation support
- `selectCluster()` - Interactive cluster selection
- `runConfigurationWizard()` / `loadExistingConfiguration()` - Configuration management

## Usage Example

```go
// Create service with dry-run and verbose modes
chartService := NewChartService(false, true)

// Interactive installation
req := utilTypes.InstallationRequest{
    Args:           []string{"my-cluster"},
    DryRun:        false,
    NonInteractive: false,
    Verbose:       true,
}
err := chartService.Install(req)

// Non-interactive CI/CD installation
req = utilTypes.InstallationRequest{
    DeploymentMode: "oss-tenant",
    NonInteractive: true,
    DryRun:        false,
}
err = chartService.InstallWithContext(ctx, req)
```