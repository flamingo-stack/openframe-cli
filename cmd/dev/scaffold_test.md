This file contains comprehensive unit tests for the Skaffold command functionality in the OpenFrame CLI development module.

## Key Components

- **TestGetScaffoldCmd**: Tests basic command properties, argument validation, and RunE function setup
- **TestScaffoldCmd_FlagBinding**: Verifies all expected flags are present and properly configured
- **TestScaffoldCmd_Examples**: Validates command examples and usage documentation
- **TestScaffoldCmd_FlagTypes**: Tests flag parsing with various data types (int, string, bool)
- **TestScaffoldCmd_FlagDefaults**: Ensures default values are correctly set for all flags
- **TestScaffoldCmd_FlagToModelMapping**: Tests mapping between CLI flags and ScaffoldFlags model
- **TestScaffoldCmd_ArgumentHandling**: Validates argument validation logic (0-1 cluster names allowed)
- **TestScaffoldCmd_UsageText**: Checks command documentation and help text content

## Usage Example

```go
// Run tests for the scaffold command
func TestExample(t *testing.T) {
    cmd := getScaffoldCmd()
    
    // Test flag parsing
    err := cmd.ParseFlags([]string{
        "--port", "9090",
        "--namespace", "my-namespace",
        "--skip-bootstrap",
    })
    require.NoError(t, err)
    
    // Verify flag values
    port, _ := cmd.Flags().GetInt("port")
    assert.Equal(t, 9090, port)
}
```

The tests ensure the Skaffold command properly handles development deployment with live reloading, validates prerequisites, and supports cluster bootstrapping with configurable options.