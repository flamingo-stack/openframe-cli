This file implements the `list` command for the OpenFrame CLI cluster management functionality, allowing users to display all registered Kubernetes clusters in a formatted table.

## Key Components

- **`getListCmd()`** - Creates and configures the cobra command for listing clusters with validation and flag setup
- **`runListClusters()`** - Core execution function that retrieves clusters from the service and displays them using configured output formatting

## Usage Example

```go
// Command registration in parent cluster command
clusterCmd.AddCommand(getListCmd())

// Command line usage examples:
// openframe cluster list
// openframe cluster list --verbose  
// openframe cluster list --quiet
```

The command supports global flags for verbose and quiet output modes, validates flags during pre-execution, and leverages the command service to fetch and display cluster information from all registered providers. The implementation follows the standard cobra command pattern with proper error handling and flag synchronization.