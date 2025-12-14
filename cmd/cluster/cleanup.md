<!-- source-hash: f2150bc054aaad7a08066538927964c0 -->
This file implements the cluster cleanup command for the OpenFrame CLI, providing functionality to remove unused Docker images and resources from cluster nodes to free disk space.

## Key Components

- **`getCleanupCmd()`** - Creates and configures the Cobra command with flags, validation, and aliases
- **`runCleanupCluster()`** - Core execution logic that handles cluster selection, type detection, and cleanup operations
- **`GetCleanupCmdForTesting()`** - Test-friendly export of the cleanup command

## Usage Example

```go
// Register the cleanup command with the cluster command group
clusterCmd.AddCommand(getCleanupCmd())

// Command usage examples:
// openframe cluster cleanup
// openframe cluster cleanup my-cluster
// openframe cluster cleanup my-cluster --force
```

The command supports interactive cluster selection when no name is provided, includes confirmation prompts (unless `--force` is used), and provides user-friendly status messages throughout the cleanup process. It integrates with the service layer for actual cleanup operations and supports both verbose output and forced execution modes.