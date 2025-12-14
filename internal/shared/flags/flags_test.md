<!-- source-hash: ddbb92eceee76fa630038a0a155a2b9f -->
This file contains comprehensive unit tests for a command-line flags management system that integrates with the Cobra CLI library. It validates the functionality of common flags like verbose, dry-run, and force across various scenarios.

## Key Components

- **TestCommonFlags**: Tests for `CommonFlags` struct default and set values
- **TestNewFlagManager**: Validates `FlagManager` creation and initialization
- **TestFlagManager_AddCommonFlags**: Tests flag registration with Cobra commands
- **TestValidateCommonFlags**: Validates flag combination logic
- **TestGetFlagDescription**: Tests flag description retrieval functionality
- **Edge case tests**: Covers nil handling, multiple commands, and flag conflicts
- **Benchmark tests**: Performance testing for core functions

## Usage Example

```go
func TestYourFlags(t *testing.T) {
    // Test flag manager creation
    flags := &CommonFlags{}
    manager := NewFlagManager(flags)
    
    // Create and configure a command
    cmd := &cobra.Command{Use: "test"}
    manager.AddCommonFlags(cmd)
    
    // Simulate command execution with flags
    cmd.SetArgs([]string{"--verbose", "--force"})
    err := cmd.Execute()
    
    // Verify flags were set correctly
    assert.NoError(t, err)
    assert.True(t, flags.Verbose)
    assert.True(t, flags.Force)
}
```

The tests cover normal operations, edge cases (nil values, multiple commands), integration scenarios, and include benchmarks for performance validation.