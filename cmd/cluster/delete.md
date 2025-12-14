<!-- source-hash: 70656a4035625e3283a92b44787b8622 -->
Provides CLI command functionality for deleting Kubernetes clusters with interactive selection and comprehensive cleanup of associated resources.

## Key Components

- **`getDeleteCmd()`** - Creates and configures the cobra command for cluster deletion with flags and validation
- **`runDeleteCluster()`** - Main execution function that handles cluster selection, type detection, and deletion workflow
- **Interactive UI** - Uses `OperationsUI` for user-friendly cluster selection and operation feedback
- **Service Integration** - Leverages cluster service for listing, type detection, and deletion operations
- **Global Flags** - Supports force deletion and other global configuration options

## Usage Example

```go
// Command usage from CLI
// openframe cluster delete my-cluster
// openframe cluster delete my-cluster --force
// openframe cluster delete  # interactive selection

// Function integration
deleteCmd := getDeleteCmd()
parentCmd.AddCommand(deleteCmd)

// The command will:
// 1. List available clusters
// 2. Allow interactive selection if no name provided
// 3. Detect cluster type automatically
// 4. Perform cleanup of intercepts, Docker resources, and config
// 5. Show progress and success/error messages
```

The delete command provides a safe, interactive way to remove clusters with proper cleanup of all associated resources including intercepts, Docker containers, and configuration files.