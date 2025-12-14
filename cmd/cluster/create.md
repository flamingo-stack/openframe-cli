This file implements the `create` command for creating Kubernetes clusters in the OpenFrame CLI tool. It handles both interactive and non-interactive cluster creation modes with comprehensive flag validation and configuration management.

## Key Components

- **`getCreateCmd()`** - Returns a configured Cobra command for cluster creation with flags, validation, and help text
- **`runCreateCluster()`** - Main execution function that handles cluster configuration and creation logic
- **Interactive Mode** - Uses UI layer for step-by-step cluster configuration when `--skip-wizard` is not set
- **Non-Interactive Mode** - Builds cluster config directly from command flags and arguments
- **Validation** - Comprehensive validation for cluster names, node counts, and configuration flags

## Usage Example

```go
// Interactive cluster creation with selection menu
cmd := getCreateCmd()
cmd.Execute() // Shows creation mode selection

// Non-interactive creation with defaults
args := []string{"my-cluster"}
cmd.SetArgs(append(args, "--skip-wizard"))
cmd.Execute()

// Custom configuration with flags
args = []string{"production-cluster"}
cmd.SetArgs(append(args, "--nodes", "5", "--type", "k3d", "--skip-wizard"))
cmd.Execute()

// Dry run to preview configuration
cmd.SetArgs([]string{"test-cluster", "--dry-run"})
cmd.Execute()
```

The command supports default cluster names, validates input parameters, and provides both quick-start and customizable configuration options for different use cases.