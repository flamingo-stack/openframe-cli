<!-- source-hash: 944c1f5d9af2483b29568385e468d616 -->
This file defines the main development command for the Openframe CLI, providing local development workflows with Telepresence and Skaffold integration.

## Key Components

- **GetDevCmd()** - Returns the root `dev` command with subcommands for traffic interception and live reloading
- **PersistentPreRunE** - Validates prerequisites for intercept and skaffold operations before execution
- **Subcommands** - Integrates `intercept` and `skaffold` commands for development workflows
- **Global Flags** - Adds shared configuration flags via `models.AddGlobalFlags()`

## Usage Example

```go
package main

import (
    "github.com/flamingo-stack/openframe-cli/internal/dev"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "openframe"}
    
    // Add development command group
    devCmd := dev.GetDevCmd()
    rootCmd.AddCommand(devCmd)
    
    // Execute CLI
    rootCmd.Execute()
}
```

The command supports aliases and provides comprehensive help text with examples for both intercept and skaffold workflows. Prerequisites are automatically checked before command execution to ensure required tools are available.