<!-- source-hash: 6931cab1ba3cd5a75d7fc0ce0232d7b9 -->
Test suite for the Skaffold command functionality in the OpenFrame CLI development tools. Validates command configuration, flag handling, argument validation, and proper integration with the scaffold model.

## Key Components

- **TestGetScaffoldCmd**: Validates basic command properties (usage, description, arguments)
- **TestScaffoldCmd_FlagBinding**: Verifies all expected flags are present and configured
- **TestScaffoldCmd_FlagTypes**: Tests flag parsing with various data types (int, string, bool)
- **TestScaffoldCmd_FlagDefaults**: Validates default values for all flags
- **TestScaffoldCmd_ArgumentHandling**: Tests argument validation (0-1 cluster names allowed)
- **TestScaffoldCmd_UsageText**: Verifies help text and examples content

## Usage Example

```go
// Run the complete test suite
func TestScaffoldCommand(t *testing.T) {
    cmd := getScaffoldCmd()
    
    // Test flag parsing
    err := cmd.ParseFlags([]string{
        "--port", "9090",
        "--namespace", "dev",
        "--skip-bootstrap",
    })
    require.NoError(t, err)
    
    // Verify parsed values
    port, _ := cmd.Flags().GetInt("port")
    assert.Equal(t, 9090, port)
    
    // Test argument validation
    err = cmd.Args(cmd, []string{"my-cluster"})
    assert.NoError(t, err)
}
```

The tests ensure the scaffold command properly handles development workflow flags like port forwarding, namespace configuration, and Docker image settings for live reloading functionality.