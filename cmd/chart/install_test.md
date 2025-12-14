This file contains comprehensive unit tests for the Helm chart installation command, validating command structure, flags, and argument handling.

## Key Components

- **TestInstallCommand**: Tests basic command structure (name, descriptions, RunE function)
- **TestInstallCommandFlags**: Validates command flags exist with correct defaults and shorthand options
- **TestInstallCommandHelp**: Ensures help text contains expected ArgoCD installation information
- **TestInstallCommandUsage**: Verifies command usage pattern
- **TestInstallCommandWithDryRun**: Tests dry-run flag parsing and extraction
- **TestInstallCommandFlagHandling**: Table-driven tests for various flag combinations
- **MockExecutor**: Test double for mocking command execution
- **TestRunInstallCommand**: Validates the main command runner function

## Usage Example

```go
// Run specific test
func TestInstallCommand(t *testing.T) {
    cmd := getInstallCmd()
    
    assert.Equal(t, "install", cmd.Name())
    assert.NotEmpty(t, cmd.Short)
    assert.NotNil(t, cmd.RunE)
}

// Test flag handling
func TestFlags(t *testing.T) {
    cmd := getInstallCmd()
    cmd.Flags().Set("dry-run", "true")
    
    flags, err := extractInstallFlags(cmd)
    assert.NoError(t, err)
    assert.True(t, flags.DryRun)
}

// Use mock executor
mockExec := NewMockExecutor()
mockExec.results["helm install"] = &executor.CommandResult{ExitCode: 0}
```