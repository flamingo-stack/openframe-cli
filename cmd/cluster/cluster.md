This file defines the root cluster command for the OpenFrame CLI, providing Kubernetes cluster lifecycle management with support for K3d local development clusters.

## Key Components

- **GetClusterCmd()** - Returns the main cluster command with subcommands for create, delete, list, status, and cleanup operations
- **Global flag initialization** - Sets up shared configuration flags across all cluster subcommands
- **Prerequisites checking** - Validates required dependencies before executing cluster operations
- **UI integration** - Displays OpenFrame logo and provides consistent user experience

## Usage Example

```go
// Get the cluster command and add it to your CLI
clusterCmd := GetClusterCmd()
rootCmd.AddCommand(clusterCmd)

// The command supports various operations:
// openframe cluster create    - Interactive cluster creation
// openframe cluster delete    - Remove clusters and cleanup
// openframe cluster list      - Show managed clusters  
// openframe cluster status    - Display cluster details
// openframe cluster cleanup   - Remove unused resources

// Short alias is also available:
// openframe k create
```

The command includes persistent pre-run hooks that check prerequisites and display the OpenFrame logo, ensuring a consistent experience across all cluster operations.