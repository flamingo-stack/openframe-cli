This file defines the root chart command for the OpenFrame CLI, providing Helm chart management functionality with ArgoCD lifecycle operations.

## Key Components

- **GetChartCmd()** - Returns the main chart command with subcommands for Helm chart management
- **chart command** - Root command with alias "c" for managing Helm charts
- **Prerequisites check** - Validates and installs required dependencies before command execution
- **UI integration** - Shows logo and provides help when appropriate

## Usage Example

```go
// Get the chart command and add it to your CLI
chartCmd := GetChartCmd()
rootCmd.AddCommand(chartCmd)

// The command supports these operations:
// openframe chart install
// openframe chart install my-cluster
// openframe c install  // using alias
```

The command includes a persistent pre-run hook that checks prerequisites and conditionally displays the UI logo. When run without subcommands, it shows the logo and help text. The actual install functionality is provided by the `getInstallCmd()` subcommand (not shown in this file).