This file provides the CLI command implementation for installing ArgoCD and app-of-apps on Kubernetes clusters. It handles command-line flag parsing, validation, and delegates the actual installation logic to service functions.

## Key Components

- **`getInstallCmd()`** - Creates the Cobra command for the `install` subcommand with usage examples and help text
- **`runInstallCommand()`** - Main command handler that extracts flags and calls installation services
- **`InstallFlags`** - Struct containing all command-line flags (force, dry-run, GitHub repo/branch, cert directory, deployment mode, non-interactive)
- **`extractInstallFlags()`** - Parses and validates command flags, including deployment mode validation
- **`addInstallFlags()`** - Defines all available command-line flags with defaults
- **`getVerboseFlag()`** - Utility to extract verbose flag from command hierarchy

## Usage Example

```go
// Create the install command
installCmd := getInstallCmd()

// Example command execution would parse flags like:
// openframe chart install my-cluster --deployment-mode=oss-tenant --dry-run

// The command supports various deployment modes:
// - oss-tenant
// - saas-tenant  
// - saas-shared

// Non-interactive mode requires deployment-mode to be specified
// openframe chart install --deployment-mode=saas-shared --non-interactive
```

The command validates that non-interactive mode requires a deployment mode and ensures only valid deployment modes are accepted.