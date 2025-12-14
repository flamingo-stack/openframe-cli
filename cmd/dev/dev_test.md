<!-- source-hash: d86e8c80d9027d457640856bb25b5540 -->
Test suite for the dev command module, validating command structure, subcommands, flags, and behavior of the development tools CLI interface.

## Key Components

- **TestGetDevCmd**: Main test verifying the dev command structure, aliases, description, and presence of intercept/skaffold subcommands
- **TestDevCmd_Examples**: Validates that usage examples are included in command documentation
- **TestDevCmd_RunE**: Tests the command's execution behavior when called without subcommands
- **TestDevCmd_GlobalFlags**: Verifies global flag defaults (verbose, silent, dry-run)
- **TestDevCmd_FlagInheritance**: Ensures subcommands properly inherit global flags from parent command

## Usage Example

```go
// Run the main dev command test
func TestDevCommand(t *testing.T) {
    cmd := GetDevCmd()
    
    // Verify command structure
    assert.Equal(t, "dev", cmd.Use)
    assert.Equal(t, []string{"d"}, cmd.Aliases)
    
    // Check subcommands
    subcommands := cmd.Commands()
    assert.Len(t, subcommands, 2)
    
    // Test flag inheritance
    interceptCmd := findSubcommand(cmd, "intercept")
    _, err := interceptCmd.InheritedFlags().GetBool("verbose")
    assert.NoError(t, err)
}
```

The test suite ensures the dev command correctly integrates Telepresence and Skaffold functionality with proper CLI structure and flag management.