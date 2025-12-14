A Go CLI command implementation that handles installation of ArgoCD and app-of-apps charts on Kubernetes clusters with comprehensive flag management and validation.

## Key Components

**Main Functions:**
- `getInstallCmd()` - Creates the cobra command with usage documentation and examples
- `runInstallCommand()` - Executes the installation workflow using extracted flags
- `extractInstallFlags()` - Parses and validates all command-line flags
- `addInstallFlags()` - Defines available command flags and their defaults

**Types:**
- `InstallFlags` - Struct containing all installation configuration options

**Key Features:**
- Supports multiple deployment modes (oss-tenant, saas-tenant, saas-shared)
- Interactive and non-interactive modes
- Dry-run capability and force installation options
- Configurable GitHub repository and branch selection

## Usage Example

```go
// Create the install command
cmd := getInstallCmd()

// Example command usage scenarios:
// openframe chart install my-cluster
// openframe chart install --deployment-mode=oss-tenant --non-interactive
// openframe chart install --github-branch develop --dry-run

// The command automatically validates flags and delegates to services.InstallChartsWithConfig()
// for the actual installation logic
```

The command provides comprehensive validation including deployment mode verification and ensures non-interactive mode has required parameters.