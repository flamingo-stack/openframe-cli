Implements the cluster cleanup command for the OpenFrame CLI, allowing users to remove unused Docker images and resources from cluster nodes to free up disk space.

## Key Components

- **`getCleanupCmd()`** - Creates and configures the cobra command with flags, aliases, and validation
- **`runCleanupCluster()`** - Main execution logic that handles cluster selection, type detection, and cleanup operation
- **`GetCleanupCmdForTesting()`** - Exported function for testing the cleanup command

## Usage Example

```go
// Basic cleanup command setup
cleanupCmd := getCleanupCmd()

// The command supports various usage patterns:
// openframe cluster cleanup                    // Interactive cluster selection
// openframe cluster cleanup my-cluster        // Cleanup specific cluster
// openframe cluster cleanup my-cluster --force // Skip confirmation prompts

// Programmatic execution example
err := runCleanupCluster(cmd, []string{"my-cluster"})
if err != nil {
    log.Fatalf("Cleanup failed: %v", err)
}
```

The command includes comprehensive error handling, user-friendly UI interactions for cluster selection, automatic cluster type detection, and integrates with the service layer for the actual cleanup operations. It supports both interactive and non-interactive modes through the force flag.