<!-- source-hash: 2f82ba41f257ba53a9cd4c8f349d62ce -->
This file defines the central flag container and management system for cluster commands, providing a unified structure to handle command-line flags across different cluster operations.

## Key Components

- **FlagContainer**: Main struct that aggregates all cluster command flags and their dependencies
- **NewFlagContainer()**: Constructor that creates a new container with default values (k3d cluster, 3 nodes, K8s v1.31.5)
- **SyncGlobalFlags()**: Synchronizes global flags across all command-specific flag instances
- **Reset()**: Resets all flags to zero values for testing purposes
- **GetGlobal()** and **GetExecutor()**: Interface methods for accessing global flags and command executor

## Usage Example

```go
// Create a new flag container with defaults
flags := NewFlagContainer()

// Access specific command flags
flags.Create.NodeCount = 5
flags.Create.ClusterType = "k3d"

// Sync global flags to all commands
flags.Global.Verbose = true
flags.SyncGlobalFlags()

// Use in testing
flags.TestManager = &k3d.K3dManager{}
flags.Reset() // Clean state for tests
```

The container serves as a central configuration hub that ensures consistent flag handling across create, list, status, delete, and cleanup cluster operations while providing testing capabilities through dependency injection.