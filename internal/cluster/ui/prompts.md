<!-- source-hash: 8151ca51ab5e4c069c7aaaacb5219f05 -->
This file provides UI components and interactive prompts specifically for cluster operations, handling user selection, confirmation, and display formatting for cluster management commands.

## Key Components

- **SelectClusterByName**: Interactive cluster selection from a list of available clusters
- **HandleClusterSelection**: Common pattern for getting cluster names from CLI args or interactive selection
- **ConfirmClusterDeletion**: User confirmation prompt for cluster deletion operations
- **ShowClusterOperationCancelled**: Standardized cancellation message display
- **FormatClusterSuccessMessage**: Formatted success message with cluster details
- **Type aliases**: Re-exports `ClusterType` and `ClusterInfo` from domain models for UI convenience

## Usage Example

```go
// Interactive cluster selection
clusters := []ClusterInfo{
    {Name: "dev-cluster", Type: ClusterTypeK3d},
    {Name: "prod-cluster", Type: ClusterTypeGKE},
}

// Select from available clusters
clusterName, err := SelectClusterByName(clusters, "Select a cluster to manage:")
if err != nil {
    return err
}

// Handle selection from CLI args or prompt
selectedCluster, err := HandleClusterSelection(clusters, os.Args[1:], "Choose cluster:")

// Confirm deletion with safety check
confirmed, err := ConfirmClusterDeletion("my-cluster", false)
if confirmed {
    // Proceed with deletion
    fmt.Println(FormatClusterSuccessMessage("my-cluster", "k3d", "deleted"))
}
```