Test file for the dev command module that validates the dev command structure, subcommands, and flag inheritance in the OpenFrame CLI.

## Key Components

- **TestGetDevCmd**: Main test function that validates the dev command properties, aliases, description, and presence of required subcommands (intercept and skaffold)
- **TestDevCmd_Examples**: Verifies that usage examples are included in the command's long description
- **TestDevCmd_RunE**: Tests the command's execution function when called without arguments
- **TestDevCmd_GlobalFlags**: Validates default values for global flags (verbose, silent, dry-run)
- **TestDevCmd_FlagInheritance**: Ensures subcommands properly inherit global flags from the parent dev command

## Usage Example

```go
func TestMyDevCommand(t *testing.T) {
    // Get the dev command for testing
    cmd := GetDevCmd()
    
    // Verify command structure
    assert.Equal(t, "dev", cmd.Use)
    assert.Equal(t, []string{"d"}, cmd.Aliases)
    
    // Check subcommands exist
    subcommands := cmd.Commands()
    assert.Len(t, subcommands, 2)
    
    // Test global flags
    verbose, err := cmd.PersistentFlags().GetBool("verbose")
    assert.NoError(t, err)
    assert.False(t, verbose)
}
```