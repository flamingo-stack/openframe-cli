Implements the `delete` subcommand for the cluster management CLI, providing functionality to delete Kubernetes clusters with interactive selection and cleanup of associated resources.

## Key Components

- **`getDeleteCmd()`** - Creates and configures the Cobra command for cluster deletion with flags and validation
- **`runDeleteCluster()`** - Main execution function that handles cluster selection, deletion, and user feedback
- **Interactive Selection** - Uses UI components to allow users to select clusters for deletion with confirmation prompts
- **Force Flag Support** - Provides `--force` option to skip confirmation dialogs
- **Resource Cleanup** - Stops intercepts, deletes clusters, and cleans up Docker resources

## Usage Example

```go
// Basic cluster deletion with interactive selection
cmd := exec.Command("openframe", "cluster", "delete")

// Delete specific cluster by name
cmd := exec.Command("openframe", "cluster", "delete", "my-cluster")

// Force deletion without confirmation
cmd := exec.Command("openframe", "cluster", "delete", "my-cluster", "--force")

// Integrating the command in CLI
rootCmd.AddCommand(getDeleteCmd())
```

The command supports both interactive cluster selection (when no name provided) and direct deletion by name, with built-in error handling and user-friendly operation feedback throughout the deletion process.