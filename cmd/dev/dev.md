This file provides the main `dev` command for the OpenFrame CLI, offering local Kubernetes development workflows through traffic interception and service deployment tools.

## Key Components

- **`GetDevCmd()`**: Returns the root dev command with subcommands for intercept and skaffold operations
- **Subcommands**: Integrates `getInterceptCmd()` and `getScaffoldCmd()` for specific development workflows
- **Prerequisites validation**: Automatically checks tool availability based on the subcommand being executed
- **Global flags**: Inherits shared configuration options through `models.AddGlobalFlags()`

## Usage Example

```go
// Get the complete dev command with all subcommands
devCmd := GetDevCmd()

// Add to root CLI command
rootCmd.AddCommand(devCmd)

// The command supports these workflows:
// openframe dev intercept my-service    # Traffic interception
// openframe dev skaffold my-cluster     # Live reloading deployment
// openframe dev                         # Show help and logo
```

The command automatically validates prerequisites (Telepresence for intercept, Skaffold for deployment) and displays the OpenFrame logo when appropriate. It serves as the entry point for local development workflows that integrate with remote Kubernetes clusters.