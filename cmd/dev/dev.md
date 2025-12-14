This file defines the main `dev` command for the OpenFrame CLI, providing development tools for local Kubernetes workflows including traffic interception and live reloading capabilities.

## Key Components

- **`GetDevCmd()`**: Returns the root development command with subcommands for intercept and skaffold operations
- **Subcommands**: Integrates `intercept` and `skaffold` commands for traffic management and deployment workflows
- **Prerequisites checking**: Validates required tools (Telepresence, Skaffold) before command execution
- **Global flags**: Supports shared configuration options across development commands

## Usage Example

```go
// Get the dev command with all subcommands
devCmd := dev.GetDevCmd()

// The command supports these operations:
// openframe dev intercept my-service    - Intercept traffic to local development
// openframe dev skaffold my-cluster     - Deploy with live reloading
// openframe dev                         - Show help and available commands
```

The command automatically checks prerequisites based on the subcommand being executed and displays the OpenFrame logo for user interactions. It serves as the entry point for all local development workflows in Kubernetes environments.