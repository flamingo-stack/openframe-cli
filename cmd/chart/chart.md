A command factory function that creates the root Helm chart management command for the OpenFrame CLI, providing ArgoCD installation and management capabilities.

## Key Components

- **GetChartCmd()**: Factory function that returns the configured chart command with subcommands
- **Prerequisites checker**: Automatically validates and installs required dependencies before command execution
- **UI integration**: Shows logo and context-aware help output
- **Command aliases**: Supports shorthand "c" alias for the chart command

## Usage Example

```go
// Register the chart command with the root CLI
rootCmd := &cobra.Command{Use: "openframe"}
rootCmd.AddCommand(chart.GetChartCmd())

// The command supports these operations:
// openframe chart install           - Install ArgoCD with interactive prompts
// openframe c install my-cluster   - Install ArgoCD on specific cluster
// openframe chart                  - Show help with logo
```

The command automatically handles prerequisite checking, UI branding, and delegates to subcommands like `install` for actual chart operations. It requires an existing cluster created with the cluster management commands.