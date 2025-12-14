Test file for the chart install command functionality, validating command structure, flag handling, and integration points.

## Key Components

**Test Functions:**
- `TestInstallCommand` - Validates basic command structure and properties
- `TestInstallCommandFlags` - Tests flag definitions, shortcuts, and defaults
- `TestInstallCommandHelp` - Verifies help text content and usage examples
- `TestInstallCommandFlagHandling` - Tests flag parsing with various combinations
- `TestRunInstallCommand` - Validates command execution structure

**Mock Infrastructure:**
- `MockExecutor` - Mock implementation for testing command execution
- `InstallFlags` struct validation - Tests flag extraction and mapping

## Usage Example

```go
// Test flag handling
func TestCustomFlags(t *testing.T) {
    cmd := getInstallCmd()
    cmd.Flags().Set("dry-run", "true")
    cmd.Flags().Set("force", "true")
    
    flags, err := extractInstallFlags(cmd)
    assert.NoError(t, err)
    assert.True(t, flags.DryRun)
    assert.True(t, flags.Force)
}

// Mock executor for testing
mockExec := NewMockExecutor()
mockExec.results["helm install"] = &executor.CommandResult{
    ExitCode: 0,
    Stdout: "installation complete",
}
```

The test suite covers command validation, flag parsing, help text verification, and provides mocking utilities for integration testing scenarios.