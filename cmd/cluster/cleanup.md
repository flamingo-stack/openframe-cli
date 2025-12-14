This file implements the cluster cleanup command for the OpenFrame CLI, providing functionality to remove unused Docker images and resources from cluster nodes to free up disk space.

## Key Components

- **`getCleanupCmd()`** - Creates and configures the cobra command with flags, aliases, and validation
- **`runCleanupCluster()`** - Main execution logic that handles cluster selection, type detection, and cleanup operations
- **`GetCleanupCmdForTesting()`** - Test-friendly wrapper that exposes the command for unit testing

## Usage Example

```go
// Register the cleanup command
cleanupCmd := getCleanupCmd()
rootCmd.AddCommand(cleanupCmd)

// The command supports various usage patterns:
// openframe cluster cleanup                    # Interactive cluster selection
// openframe cluster cleanup my-cluster        # Cleanup specific cluster
// openframe cluster cleanup my-cluster --force # Skip confirmation prompts
```

The command integrates with the broader cluster management system, using shared utilities for flag management, UI operations, and service layer interactions. It provides both interactive and non-interactive modes, with proper error handling and user feedback throughout the cleanup process.