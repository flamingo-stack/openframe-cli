This file implements the "list" subcommand for the OpenFrame CLI cluster management functionality, providing a way to display all managed Kubernetes clusters in a formatted table.

## Key Components

- **`getListCmd()`** - Creates and configures the cobra command for listing clusters with validation and flag setup
- **`runListClusters()`** - Executes the cluster listing operation and handles output formatting
- **Command validation** - Pre-run validation of global and list-specific flags
- **Service integration** - Uses cluster service for data retrieval and display formatting

## Usage Example

```go
// Create the list command
listCmd := getListCmd()

// The command supports various flags and options:
// openframe cluster list
// openframe cluster list --verbose
// openframe cluster list --quiet

// Internal usage in command execution
func runListClusters(cmd *cobra.Command, args []string) error {
    service := utils.GetCommandService()
    clusters, err := service.ListClusters()
    if err != nil {
        return fmt.Errorf("failed to list clusters: %w", err)
    }
    
    globalFlags := utils.GetGlobalFlags()
    return service.DisplayClusterList(clusters, globalFlags.List.Quiet, globalFlags.Global.Verbose)
}
```

The command integrates with the global flag system and provides both quiet and verbose output modes for different use cases.