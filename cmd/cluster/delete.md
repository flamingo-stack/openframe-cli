This file implements the `delete` subcommand for the cluster management CLI, providing functionality to delete Kubernetes clusters and clean up associated resources.

## Key Components

- **`getDeleteCmd()`** - Creates and configures the cobra command for cluster deletion with flags and validation
- **`runDeleteCluster()`** - Main execution function that handles cluster selection, type detection, and deletion workflow
- **Interactive UI** - Uses `OperationsUI` for user-friendly cluster selection and confirmation dialogs
- **Service Integration** - Leverages cluster service for listing, type detection, and deletion operations

## Usage Example

```go
// Command usage from CLI
// openframe cluster delete my-cluster
// openframe cluster delete my-cluster --force
// openframe cluster delete  // interactive selection

// Internal function usage
cmd := getDeleteCmd()

// The command handles:
// 1. Global flag validation and synchronization
// 2. Interactive cluster selection if no name provided
// 3. Confirmation prompts (unless --force flag used)
// 4. Cluster type auto-detection
// 5. Resource cleanup and deletion
// 6. User-friendly progress and error messages

// Typical workflow:
// - Lists available clusters
// - Shows selection UI for user choice
// - Confirms deletion operation
// - Detects cluster type (kind, minikube, etc.)
// - Executes deletion with cleanup
// - Reports success/failure status
```