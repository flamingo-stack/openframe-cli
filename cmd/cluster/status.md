This file implements the cluster status command for the OpenFrame CLI, providing detailed information about Kubernetes cluster health, nodes, and applications.

## Key Components

- **`getStatusCmd()`** - Creates and configures the cobra command for cluster status operations
- **`runClusterStatus()`** - Main execution function that handles cluster selection and displays status information
- **Command configuration** - Sets up CLI arguments, flags, and validation for the status command

## Usage Example

```go
// Create the status command
statusCmd := getStatusCmd()

// The command supports various usage patterns:
// openframe cluster status my-cluster
// openframe cluster status --detailed
// openframe cluster status  # interactive selection

// Execute with specific cluster
args := []string{"my-cluster"}
err := runClusterStatus(statusCmd, args)
if err != nil {
    log.Fatal(err)
}
```

The command integrates with the cluster service layer to retrieve and display comprehensive cluster information including health metrics, node status, installed applications, and resource usage. It supports both direct cluster specification and interactive selection when no cluster name is provided.