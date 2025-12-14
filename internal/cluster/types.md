This file defines the core type system for managing cluster command flags and dependencies in a centralized container structure.

## Key Components

**FlagContainer**
- Central structure holding all cluster command flags and dependencies
- Contains flag instances for global, create, list, status, delete, and cleanup operations
- Includes executor and test manager dependencies for command execution
- Implements `models.CommandFlags` and `models.CommandExecutor` interfaces

**Key Methods**
- `NewFlagContainer()` - Creates initialized container with default values
- `SyncGlobalFlags()` - Propagates global flags to all command-specific flags
- `Reset()` - Resets all flags to zero values for testing
- `GetGlobal()` - Interface method returning global flags
- `GetExecutor()` - Interface method returning command executor

## Usage Example

```go
// Create a new flag container with defaults
container := NewFlagContainer()

// Access specific command flags
container.Create.ClusterType = "k3d"
container.Create.NodeCount = 5

// Sync global flags across all commands
container.Global.Debug = true
container.SyncGlobalFlags()

// Use in command execution
executor := container.GetExecutor()
globalFlags := container.GetGlobal()

// Reset for testing
container.Reset()
```

The container serves as the primary interface between CLI commands and the underlying cluster management system.