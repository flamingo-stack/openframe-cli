Implements the delete command functionality for Kubernetes cluster management, providing cluster selection and deletion with cleanup of associated resources.

## Key Components

- **`getDeleteCmd()`** - Creates and configures the cobra command for cluster deletion with flags and validation
- **`runDeleteCluster()`** - Main execution function that handles cluster selection, type detection, and deletion process
- **Interactive cluster selection** - Uses UI components to provide user-friendly cluster selection with confirmation
- **Force deletion support** - Allows bypassing confirmation prompts with `--force` flag
- **Resource cleanup** - Stops intercepts, deletes clusters, and cleans up Docker resources

## Usage Example

```go
// Register the delete command
deleteCmd := getDeleteCmd()
rootCmd.AddCommand(deleteCmd)

// Command usage examples:
// openframe cluster delete my-cluster
// openframe cluster delete my-cluster --force
// openframe cluster delete  # interactive selection

// The command will:
// 1. List available clusters
// 2. Allow interactive selection if no name provided
// 3. Detect cluster type (kind, k3d, etc.)
// 4. Confirm deletion (unless --force flag used)
// 5. Clean up all associated resources
```