<!-- source-hash: 8e2725b0e344a8a3b6416b54909d4ba4 -->
Test suite for the Helm chart installation command functionality, validating command structure, flag handling, and flag extraction logic.

## Key Components

- **TestInstallCommand**: Validates basic command structure including name, descriptions, and RunE function
- **TestInstallCommandFlags**: Tests flag existence, shortcuts, and default values for `--force` and `--dry-run`
- **TestInstallCommandHelp**: Verifies help text contains expected ArgoCD installation content
- **TestInstallCommandUsage**: Confirms command usage format
- **TestInstallCommandFlagHandling**: Parameterized tests for flag parsing and extraction
- **MockExecutor**: Test double for command execution with configurable results and errors
- **TestRunInstallCommand**: Validates the main command execution structure

## Usage Example

```go
// Test flag extraction
cmd := getInstallCmd()
cmd.Flags().Set("dry-run", "true")
cmd.Flags().Set("force", "true")

flags, err := extractInstallFlags(cmd)
assert.NoError(t, err)
assert.True(t, flags.DryRun)

// Use mock executor for testing
mockExec := NewMockExecutor()
mockExec.results["kubectl get nodes"] = &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "node1 Ready",
}
result, err := mockExec.Execute(ctx, "kubectl", "get", "nodes")
```

The tests focus on command structure validation rather than full execution to avoid interactive UI dependencies, with integration tests handling complete command flows.