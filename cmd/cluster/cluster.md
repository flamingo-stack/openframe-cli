This file defines the main cluster command for the OpenFrame CLI, providing Kubernetes cluster lifecycle management functionality through a structured cobra command interface.

## Key Components

- **GetClusterCmd()** - Main function that returns the configured cluster command with subcommands
- **clusterCmd** - Root cobra command for cluster operations with aliases and help text
- **PersistentPreRunE** - Hook that displays UI logo and checks prerequisites for subcommands
- **Subcommands** - Integrates create, delete, list, status, and cleanup cluster operations
- **Global flags** - Adds shared configuration options across all cluster commands

## Usage Example

```go
// Get the cluster command and add it to your root CLI
rootCmd := &cobra.Command{Use: "openframe"}
clusterCmd := GetClusterCmd()
rootCmd.AddCommand(clusterCmd)

// The command supports various operations:
// openframe cluster create    - Interactive cluster creation
// openframe k list           - List clusters (using alias)
// openframe cluster status   - Show cluster details
// openframe cluster cleanup  - Remove unused resources
```

The command provides a complete cluster management interface with prerequisite checking, UI integration, and support for K3d local development clusters.