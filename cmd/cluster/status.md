This file implements the `status` command for the cluster management CLI, providing detailed cluster information including health, nodes, applications, and resource usage.

## Key Components

- **`getStatusCmd()`** - Creates and configures the Cobra command for cluster status operations
- **`runClusterStatus()`** - Executes the status check logic with cluster selection and service integration
- **Command flags** - Supports `--detailed` and `--no-apps` options for customizing output
- **Interactive selection** - Provides user-friendly cluster selection when no name is specified

## Usage Example

```go
// Register the status command
clusterCmd.AddCommand(getStatusCmd())

// Command usage examples:
// openframe cluster status my-cluster
// openframe cluster status --detailed
// openframe cluster status my-cluster --no-apps
```

The command integrates with the service layer to retrieve cluster information and uses the operations UI for interactive cluster selection when no cluster name is provided as an argument.