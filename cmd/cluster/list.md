Provides the `list` subcommand for the cluster management CLI, allowing users to view all Kubernetes clusters managed by OpenFrame CLI in a formatted table.

## Key Components

- **`getListCmd()`** - Creates and configures the cobra command for listing clusters with flags and validation
- **`runListClusters()`** - Core execution function that retrieves and displays cluster information
- **Flag integration** - Supports global flags like `--verbose` and `--quiet` for output customization
- **Pre-run validation** - Validates flags before command execution

## Usage Example

```go
// Register the list command with the cluster command group
clusterCmd.AddCommand(getListCmd())

// Example CLI usage:
// openframe cluster list
// openframe cluster list --verbose
// openframe cluster list --quiet
```

The command retrieves cluster data from all registered providers through the command service and displays it in a table format showing cluster name, type, status, and node count. Output verbosity can be controlled through global flags, with quiet mode for minimal output and verbose mode for detailed information.