Test suite for the dev command module that validates the development tools command structure and behavior. Contains comprehensive tests for command properties, subcommands, flag inheritance, and global configuration.

## Key Components

- **TestGetDevCmd**: Validates the main dev command properties, aliases, description, and verifies presence of intercept and skaffold subcommands
- **TestDevCmd_Examples**: Ensures command examples are properly included in help text
- **TestDevCmd_RunE**: Tests the command's execution behavior when no subcommand is provided
- **TestDevCmd_GlobalFlags**: Validates default values for global flags (verbose, silent, dry-run)
- **TestDevCmd_FlagInheritance**: Confirms that subcommands properly inherit global flags from the parent command

## Usage Example

```go
func TestExample(t *testing.T) {
    // Get the dev command for testing
    cmd := GetDevCmd()
    
    // Validate command structure
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