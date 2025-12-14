<!-- source-hash: c37bfa46d7ac735690b6980c85d2db8c -->
This file provides helper utilities for managing global flags, command services, and common CLI setup patterns across cluster commands in the OpenFrame CLI tool.

## Key Components

- **Global Flag Management**: `InitGlobalFlags()`, `GetGlobalFlags()` - Thread-safe initialization and access to global flag container
- **Command Services**: `GetCommandService()`, `GetSuppressedCommandService()` - Factory functions for creating cluster services with different UI modes
- **Command Wrapper**: `WrapCommandWithCommonSetup()` - Wraps commands with consistent error handling and setup logic
- **Flag Operations**: `SyncGlobalFlags()`, `ValidateGlobalFlags()` - Synchronization and validation utilities
- **Testing Support**: `SetTestExecutor()`, `ResetGlobalFlags()`, `SetVerboseForIntegrationTesting()` - Test isolation and mock injection

## Usage Example

```go
// Initialize and use global flags
utils.InitGlobalFlags()
flags := utils.GetGlobalFlags()

// Create a command service
service := utils.GetCommandService()

// Wrap a command with common setup
cmd.RunE = utils.WrapCommandWithCommonSetup(func(cmd *cobra.Command, args []string) error {
    // Your command logic here
    return service.CreateCluster(clusterName)
})

// For testing - inject mock executor
utils.SetTestExecutor(mockExecutor)
defer utils.ResetGlobalFlags()
```

The file centralizes flag management and command setup patterns, ensuring consistent behavior across all CLI commands while providing proper test isolation mechanisms.