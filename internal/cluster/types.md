Defines the central flag container structure for managing command-line flags and dependencies across all cluster operations in the OpenFrame CLI tool.

## Key Components

- **`FlagContainer`** - Main struct that holds all cluster command flags and execution dependencies
- **`NewFlagContainer()`** - Constructor that creates a new container with default values (k3d cluster, 3 nodes, K8s v1.31.5-k3s1)
- **`SyncGlobalFlags()`** - Propagates global flags to all command-specific flag structures
- **`Reset()`** - Resets all flags to zero values for testing purposes
- **Interface implementations** - `GetGlobal()` and `GetExecutor()` for dependency injection

## Usage Example

```go
// Create a new flag container with defaults
flags := NewFlagContainer()

// Access specific command flags
flags.Create.NodeCount = 5
flags.Create.K8sVersion = "v1.30.0-k3s1"

// Sync global flags across all commands
flags.Global.Verbose = true
flags.SyncGlobalFlags()

// Reset for testing
flags.Reset()

// Set custom executor for testing
flags.Executor = mockExecutor
flags.TestManager = mockK3dManager
```

The container centralizes flag management and provides a clean interface for dependency injection, making cluster commands testable and maintainable.