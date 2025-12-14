<!-- source-hash: 1ebdb79f11ad431dc5abf5b36825a179 -->
Implements a Cobra command for checking and displaying Kubernetes cluster status information. Provides interactive cluster selection and detailed health reporting.

## Key Components

- **`getStatusCmd()`** - Creates and configures the Cobra command for cluster status operations
- **`runClusterStatus()`** - Main execution function that handles cluster selection and status display
- **Command Configuration** - Sets up CLI arguments, flags, and validation for the status command
- **Interactive Selection** - Uses UI components to allow users to select clusters when no name is provided

## Usage Example

```go
// Create the status command
statusCmd := getStatusCmd()

// The command supports several usage patterns:
// openframe cluster status my-cluster
// openframe cluster status --detailed
// openframe cluster status  # interactive selection

// Command execution flow:
func runClusterStatus(cmd *cobra.Command, args []string) error {
    service := utils.GetCommandService()
    operationsUI := ui.NewOperationsUI()
    
    clusters, err := service.ListClusters()
    if err != nil {
        return fmt.Errorf("failed to list clusters: %w", err)
    }
    
    clusterName, err := operationsUI.SelectClusterForOperation(clusters, args, "check status")
    // ... status display logic
}
```