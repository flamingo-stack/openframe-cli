This file contains comprehensive unit tests for the Helm chart installation command functionality, validating command structure, flag handling, and behavior configurations.

## Key Components

- **TestInstallCommand** - Tests basic command structure (name, descriptions, RunE function)
- **TestInstallCommandFlags** - Validates flag existence, shortcuts, and default values
- **TestInstallCommandHelp** - Checks help text content and usage information
- **TestInstallCommandFlagHandling** - Tests flag parsing with various combinations
- **MockExecutor** - Test double for command execution without actual system calls
- **TestRunInstallCommand** - Validates command execution structure and flag extraction

## Usage Example

```go
// Run flag handling tests
func TestCustomFlags(t *testing.T) {
    cmd := getInstallCmd()
    
    // Set flags programmatically
    cmd.Flags().Set("dry-run", "true")
    cmd.Flags().Set("github-branch", "develop")
    
    // Extract and validate
    flags, err := extractInstallFlags(cmd)
    require.NoError(t, err)
    assert.True(t, flags.DryRun)
    assert.Equal(t, "develop", flags.GitHubBranch)
}

// Use mock executor for testing
mockExec := NewMockExecutor()
mockExec.results["helm install"] = &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "Success",
}
```

The tests focus on command structure validation and flag parsing rather than actual installation execution, which is handled separately in integration tests.