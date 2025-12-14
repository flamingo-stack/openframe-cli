<!-- source-hash: 02d3f567d6d445c0eb4ffc09dcc897ce -->
Provides user-friendly interfaces for cluster operations including selection, confirmation dialogs, and status messages. This UI service handles interactive workflows for cluster management operations like creation, deletion, and cleanup.

## Key Components

- **OperationsUI**: Main service struct for cluster operation interfaces
- **NewOperationsUI()**: Constructor function
- **SelectClusterForOperation()**: Interactive cluster selection for general operations
- **SelectClusterForDelete()**: Cluster selection with deletion confirmation
- **SelectClusterForCleanup()**: Cluster selection with cleanup confirmation
- **ShowOperationStart/Success/Error()**: Status message display methods
- **ShowConfigurationSummary()**: Displays cluster configuration before creation
- **ShowNoResourcesMessage()**: Helper for empty resource states

## Usage Example

```go
// Create operations UI service
ui := NewOperationsUI()

// Select cluster for an operation
clusterName, err := ui.SelectClusterForOperation(clusters, args, "start")
if err != nil {
    return err
}

// Show operation progress
ui.ShowOperationStart("start", clusterName)

// Perform operation...

// Show success message
ui.ShowOperationSuccess("start", clusterName)

// For deletion with confirmation
clusterName, err = ui.SelectClusterForDelete(clusters, args, force)
if err != nil {
    return err
}
```