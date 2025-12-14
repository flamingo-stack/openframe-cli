<!-- source-hash: 23852bf80f0ef72be26bf63ac9dd808d -->
This file contains comprehensive unit tests for command-line flag structures and their associated helper functions in a CLI application.

## Key Components

**Test Functions:**
- `TestGlobalFlags` - Tests default and custom values for global flags (Verbose, DryRun, Force)
- `TestCreateFlags` - Tests cluster creation flags including inheritance of global flags
- `TestListFlags`, `TestStatusFlags`, `TestDeleteFlags`, `TestCleanupFlags` - Test command-specific flag structures
- `TestAdd*Flags` - Test functions that add flags to Cobra commands
- `TestFlagValidation` - Tests validation logic for all flag types

**Flag Structures Tested:**
- `GlobalFlags` - Common flags shared across commands
- `CreateFlags` - Cluster creation options (type, node count, K8s version)
- `ListFlags`, `StatusFlags`, `DeleteFlags`, `CleanupFlags` - Command-specific options

## Usage Example

```go
func TestYourFlags(t *testing.T) {
    // Test default flag values
    flags := &CreateFlags{}
    assert.Equal(t, 0, flags.NodeCount)
    assert.Empty(t, flags.ClusterType)
    
    // Test flag validation
    flags.NodeCount = -1
    err := ValidateCreateFlags(flags)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "node count must be at least 1")
    
    // Test flag registration with Cobra
    cmd := &cobra.Command{}
    AddCreateFlags(cmd, flags)
    typeFlag := cmd.Flags().Lookup("type")
    assert.NotNil(t, typeFlag)
}
```

The tests ensure flag structures work correctly with default values, custom values, inheritance patterns, and validation rules.