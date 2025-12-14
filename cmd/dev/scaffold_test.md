This file contains comprehensive unit tests for the Skaffold command functionality in the OpenFrame CLI development tools, validating command structure, flag handling, and argument processing.

## Key Components

- **TestGetScaffoldCmd**: Validates basic command properties including usage string, description, and function assignments
- **TestScaffoldCmd_FlagBinding**: Ensures all required flags (port, namespace, image, sync paths, skip-bootstrap, helm-values) are properly registered
- **TestScaffoldCmd_FlagTypes**: Tests flag parsing and type conversion for all command flags
- **TestScaffoldCmd_FlagDefaults**: Verifies default values for all flags are set correctly
- **TestScaffoldCmd_ArgumentHandling**: Validates argument count restrictions (0-1 cluster names allowed)
- **TestScaffoldCmd_Examples**: Checks that help text contains proper usage examples
- **TestScaffoldCmd_UsageText**: Ensures documentation includes prerequisite validation and live reloading information

## Usage Example

```go
func TestCustomFlag(t *testing.T) {
    cmd := getScaffoldCmd()
    
    // Test flag parsing
    err := cmd.ParseFlags([]string{
        "--port", "9000",
        "--namespace", "dev",
        "--skip-bootstrap",
    })
    require.NoError(t, err)
    
    // Verify parsed values
    port, _ := cmd.Flags().GetInt("port")
    assert.Equal(t, 9000, port)
    
    skipBootstrap, _ := cmd.Flags().GetBool("skip-bootstrap")
    assert.True(t, skipBootstrap)
}
```