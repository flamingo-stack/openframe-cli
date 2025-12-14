This file defines the main configuration structure and management utilities for cluster command flags in the OpenFrame CLI.

## Key Components

- **`FlagContainer`** - Central structure that holds all cluster command flags and dependencies
- **`NewFlagContainer()`** - Constructor that initializes a new flag container with default values
- **`SyncGlobalFlags()`** - Synchronizes global flags across all command-specific flag structures
- **`Reset()`** - Resets all flags to zero values for testing purposes
- **Interface implementations** - `GetGlobal()` and `GetExecutor()` for command integration

## Usage Example

```go
// Create a new flag container with defaults
flags := NewFlagContainer()

// Access specific command flags
flags.Create.ClusterType = "k3d"
flags.Create.NodeCount = 5

// Set global flags and sync across all commands
flags.Global.Verbose = true
flags.SyncGlobalFlags()

// Reset for testing
flags.Reset()

// Access dependencies
executor := flags.GetExecutor()
global := flags.GetGlobal()
```

The `FlagContainer` serves as a centralized configuration manager, ensuring consistent flag handling across all cluster-related commands while providing testing utilities and dependency injection capabilities.