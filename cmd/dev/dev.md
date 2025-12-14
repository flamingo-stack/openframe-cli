This file provides the main entry point for development workflow commands in the openframe CLI, supporting local Kubernetes development with Telepresence and Skaffold tools.

## Key Components

- **GetDevCmd()**: Returns the root `dev` command with subcommands for local development workflows
- **PersistentPreRunE**: Validates prerequisites for intercept and skaffold operations
- **Subcommands**: Integrates `intercept` and `skaffold` commands for traffic interception and live reloading
- **Global flags**: Adds shared configuration flags via `models.AddGlobalFlags()`

## Usage Example

```go
// Get the dev command and add it to your root CLI
devCmd := GetDevCmd()
rootCmd.AddCommand(devCmd)

// The command supports these workflows:
// openframe dev intercept my-service    - Intercept cluster traffic locally
// openframe dev skaffold my-cluster     - Deploy with live reloading
// openframe dev                         - Show help and logo
```

The command includes automatic prerequisite checking for Telepresence (intercept) and Skaffold tools, displays the application logo for subcommands, and provides a consistent interface for local Kubernetes development workflows.