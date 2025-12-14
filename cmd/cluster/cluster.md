This file defines the root cluster command for the OpenFrame CLI, providing Kubernetes cluster management functionality with subcommands for lifecycle operations.

## Key Components

- **GetClusterCmd()** - Main function that returns the configured cluster command with all subcommands
- **Subcommands** - create, delete, list, status, and cleanup operations for K3d clusters
- **Global flags initialization** - Sets up shared configuration flags across all cluster commands
- **Prerequisites checking** - Validates required dependencies before command execution
- **UI integration** - Displays logo and provides consistent user experience

## Usage Example

```go
// Get the cluster command for CLI integration
clusterCmd := GetClusterCmd()

// The command supports various operations:
// openframe cluster create     - Interactive cluster creation
// openframe cluster delete     - Remove existing cluster  
// openframe cluster list       - Show all managed clusters
// openframe cluster status     - Display cluster details
// openframe cluster cleanup    - Remove unused resources

// Can also be used with alias:
// openframe k create

// Add to root command
rootCmd.AddCommand(clusterCmd)
```

The command includes persistent pre-run hooks that check prerequisites and display the OpenFrame logo, ensuring a consistent experience across all cluster management operations.