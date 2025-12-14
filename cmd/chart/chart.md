Provides the root chart command for managing Helm charts and ArgoCD lifecycle operations in the OpenFrame CLI.

## Key Components

- **GetChartCmd()**: Returns the main chart command with subcommands for Helm chart management
- **Prerequisites integration**: Automatically checks and installs required dependencies before command execution
- **UI integration**: Displays logo and context-aware messaging
- **Subcommand structure**: Supports the `install` subcommand for ArgoCD deployment

## Usage Example

```go
// Get the chart command with all subcommands
chartCmd := GetChartCmd()

// Add to root command
rootCmd.AddCommand(chartCmd)

// The command supports these usage patterns:
// openframe chart            - shows help with logo
// openframe chart install    - installs ArgoCD
// openframe c install        - using alias
```

The command includes persistent pre-run hooks that validate prerequisites and conditionally display the OpenFrame logo. It serves as the entry point for all chart-related operations, requiring an existing cluster created with the cluster management commands.