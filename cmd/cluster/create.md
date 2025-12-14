This file implements the `create` command for the OpenFrame CLI cluster management functionality, providing both interactive and non-interactive cluster creation modes.

## Key Components

- **`getCreateCmd()`** - Returns a configured Cobra command for cluster creation with comprehensive flag handling and validation
- **`runCreateCluster()`** - Main execution logic that handles both interactive wizard mode and direct creation from flags
- **Interactive Mode** - Uses UI layer for step-by-step cluster configuration when `--skip-wizard` is not set
- **Non-interactive Mode** - Builds cluster configuration directly from command flags and arguments
- **Configuration Validation** - Validates cluster names, node counts, and other parameters before creation

## Usage Example

```go
// Interactive mode with default cluster name
// openframe cluster create

// Interactive mode with custom name  
// openframe cluster create my-dev-cluster

// Direct creation with flags (skip wizard)
cmd := getCreateCmd()
cmd.SetArgs([]string{"--skip-wizard", "--nodes", "3", "--type", "k3d", "my-cluster"})
err := cmd.Execute()

// Dry run to preview configuration
cmd.SetArgs([]string{"--dry-run", "--nodes", "5", "test-cluster"})
err := cmd.Execute()
```

The command supports flexible cluster creation with sensible defaults, comprehensive validation, and both guided and direct configuration approaches for different user preferences.