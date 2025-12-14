<!-- source-hash: e84232c5782aada64fc53eb1c13aa4be -->
This package provides a centralized flag management system for CLI commands using the Cobra framework, ensuring consistent flag handling across different command types.

## Key Components

- **`CommonFlags`** - Struct containing standard flags (Verbose, DryRun, Force) used across non-cluster commands
- **`FlagManager`** - Handles consistent flag setup and binding for Cobra commands
- **`NewFlagManager()`** - Constructor for creating a flag manager instance
- **`AddCommonFlags()`** - Adds standard flags to any Cobra command with consistent naming and descriptions
- **`ValidateCommonFlags()`** - Validates flag combinations (placeholder for future validation logic)
- **`GetFlagDescription()`** - Returns standardized descriptions for common flag types

## Usage Example

```go
package main

import (
    "github.com/spf13/cobra"
    "your-app/flags"
)

func main() {
    commonFlags := &flags.CommonFlags{}
    flagManager := flags.NewFlagManager(commonFlags)
    
    rootCmd := &cobra.Command{
        Use: "myapp",
        Run: func(cmd *cobra.Command, args []string) {
            if commonFlags.Verbose {
                fmt.Println("Verbose mode enabled")
            }
            if commonFlags.DryRun {
                fmt.Println("Dry run mode - no changes will be made")
            }
        },
    }
    
    flagManager.AddCommonFlags(rootCmd)
    rootCmd.Execute()
}
```