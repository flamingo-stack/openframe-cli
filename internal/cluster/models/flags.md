<!-- source-hash: 22b4235c35a9a1aef48e39b7b28dc9cf -->
This file defines command-line flag structures and utilities for cluster management commands in the OpenFrame CLI tool.

## Key Components

**Flag Structures:**
- `CreateFlags` - Flags for cluster creation (cluster type, node count, K8s version, skip wizard)
- `ListFlags` - Flags for listing clusters (quiet mode)
- `StatusFlags` - Flags for cluster status (detailed view, exclude apps)
- `DeleteFlags` - Flags for cluster deletion (force deletion)
- `CleanupFlags` - Flags for cleanup operations (aggressive cleanup)

**Flag Setup Functions:**
- `AddGlobalFlags()` - Adds common flags to commands
- `AddCreateFlags()`, `AddListFlags()`, etc. - Add command-specific flags

**Validation Functions:**
- `ValidateClusterName()` - Validates cluster names against Kubernetes DNS-1123 rules
- `ValidateCreateFlags()`, `ValidateListFlags()`, etc. - Validate flag combinations

## Usage Example

```go
// Setting up create command flags
var createFlags CreateFlags
cmd := &cobra.Command{Use: "create"}

AddGlobalFlags(cmd, &createFlags.GlobalFlags)
AddCreateFlags(cmd, &createFlags)

// Validating flags before execution
if err := ValidateCreateFlags(&createFlags); err != nil {
    return err
}

// Validating cluster name
if err := ValidateClusterName("my-cluster"); err != nil {
    return err
}
```