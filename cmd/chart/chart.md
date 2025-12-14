<!-- source-hash: 33f8519c549959fb3459a042fe93d359 -->
This file defines the root command for Helm chart management operations within the OpenFrame CLI, specifically focused on ArgoCD installation and management.

## Key Components

- **GetChartCmd()** - Returns the main chart command with aliases and subcommands
- **PersistentPreRunE** - Handles logo display and prerequisite checking for subcommands
- **RunE** - Shows help when the chart command is run without subcommands
- **getInstallCmd()** - Referenced subcommand for chart installation (defined elsewhere)

## Usage Example

```go
// Get the chart command and add it to your CLI
chartCmd := GetChartCmd()
rootCmd.AddCommand(chartCmd)

// The command supports these usage patterns:
// openframe chart          - shows help
// openframe chart install  - installs ArgoCD
// openframe c install      - same as above (using alias)
```

The command includes prerequisite checking through the `prerequisites.NewInstaller()` and displays the OpenFrame logo contextually. It requires an existing cluster created with the cluster management commands and provides a foundation for ArgoCD lifecycle management operations.