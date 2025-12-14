<!-- source-hash: 8086533d2c49b59f199ffbb62baf496a -->
Test suite for Helm manager functionality that validates chart installation, status checking, and availability detection through command execution mocking.

## Key Components

**MockExecutor**: Test double implementing `CommandExecutor` interface with configurable command results and errors for isolated testing.

**Test Functions**:
- `TestHelmManager_IsHelmInstalled` - Validates Helm binary availability detection
- `TestHelmManager_IsChartInstalled` - Tests chart installation status checking
- `TestHelmManager_InstallArgoCD` - Comprehensive ArgoCD installation testing including dry-run mode
- `TestHelmManager_GetChartStatus` - Tests chart status retrieval and JSON parsing

## Usage Example

```go
// Create a mock executor for testing
mockExec := NewMockExecutor()

// Configure expected command result
mockExec.SetResult("helm version --short", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "v3.12.0+g4f11b4a",
})

// Test the manager
manager := NewHelmManager(mockExec)
err := manager.IsHelmInstalled(context.Background())

// Verify commands were executed
commands := mockExec.GetCommands()
assert.Equal(t, []string{"helm", "version", "--short"}, commands[0])
```

The mock executor captures all executed commands and allows verification of proper Helm CLI interactions, command arguments, and error handling scenarios.