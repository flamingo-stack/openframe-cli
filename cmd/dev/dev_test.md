This file contains comprehensive unit tests for the development command module, validating command structure, subcommands, flags, and behavior.

## Key Components

- **TestGetDevCmd**: Main test function that validates the dev command's basic properties, aliases, description, and verifies the existence of intercept and skaffold subcommands
- **TestDevCmd_Examples**: Tests that usage examples are properly included in the command's long description
- **TestDevCmd_RunE**: Validates the command's execution behavior when called without arguments (should display help)
- **TestDevCmd_GlobalFlags**: Tests default values for global flags (verbose, silent, dry-run)
- **TestDevCmd_FlagInheritance**: Verifies that subcommands properly inherit global flags from the parent command

## Usage Example

```go
func TestYourDevCommand(t *testing.T) {
    // Get the dev command for testing
    cmd := GetDevCmd()
    
    // Test command properties
    assert.Equal(t, "dev", cmd.Use)
    assert.Equal(t, []string{"d"}, cmd.Aliases)
    
    // Verify subcommands exist
    subcommands := cmd.Commands()
    assert.Len(t, subcommands, 2)
    
    // Test global flags are present
    _, err := cmd.PersistentFlags().GetBool("verbose")
    assert.NoError(t, err)
}
```

The tests ensure the dev command is properly configured with correct metadata, subcommands, and flag inheritance patterns.