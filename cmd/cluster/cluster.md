<!-- source-hash: 1c793277ed7d465a129a1fdda5d5ff32 -->
This file defines the main cluster command for the OpenFrame CLI, providing Kubernetes cluster lifecycle management functionality including create, delete, list, status, and cleanup operations.

## Key Components

- **GetClusterCmd()** - Returns the root cluster command with all subcommands configured
- **Global flags initialization** - Sets up shared configuration flags across all cluster commands
- **Prerequisite checking** - Validates system requirements before executing cluster operations
- **UI integration** - Handles logo display and user interface elements

## Usage Example

```go
// Get the cluster command with all subcommands
clusterCmd := GetClusterCmd()

// Add to root command
rootCmd.AddCommand(clusterCmd)

// The command supports these operations:
// openframe cluster create    - Create new K3d cluster
// openframe cluster delete    - Remove cluster
// openframe cluster list      - Show all clusters  
// openframe cluster status    - Display cluster info
// openframe cluster cleanup   - Clean unused resources
// openframe cluster          - Show help (no subcommand)
```

The command includes persistent pre-run hooks for prerequisite validation and conditional logo display, with support for the alias `k` for shorter command usage.