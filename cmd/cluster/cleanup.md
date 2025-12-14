Implements the cluster cleanup command for removing unused Docker images and resources from cluster nodes to free disk space.

## Key Components

- **`getCleanupCmd()`** - Creates and configures the cleanup cobra command with flags and validation
- **`runCleanupCluster()`** - Main execution function that handles cluster selection, type detection, and cleanup operations
- **`GetCleanupCmdForTesting()`** - Exported wrapper for testing the cleanup command

## Usage Example

```go
// Create cleanup command
cleanupCmd := getCleanupCmd()

// Command supports multiple usage patterns:
// openframe cluster cleanup                    // Interactive cluster selection
// openframe cluster cleanup my-cluster        // Cleanup specific cluster
// openframe cluster cleanup my-cluster --force // Skip confirmation prompts

// The command flow:
// 1. Lists available clusters
// 2. Prompts user for cluster selection (if not specified)
// 3. Detects cluster type (k3d, kind, etc.)
// 4. Executes cleanup operations
// 5. Shows progress and completion status
```

The command integrates with the UI layer for user-friendly cluster selection and operation feedback, while leveraging the service layer for actual cluster management operations.