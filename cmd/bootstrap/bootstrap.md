<!-- source-hash: f4718e0a909d911174cd1ef80538dbd5 -->
This file provides the CLI command definition for bootstrapping a complete OpenFrame environment, combining cluster creation and chart installation into a single streamlined command.

## Key Components

- **GetBootstrapCmd()** - Returns a configured cobra.Command for the bootstrap operation
- **Command flags**:
  - `deployment-mode` - Specifies deployment type (oss-tenant, saas-tenant, saas-shared)
  - `non-interactive` - Enables CI/CD mode without prompts
  - `verbose` - Shows detailed logging including ArgoCD sync progress

## Usage Example

```go
package main

import (
    "github.com/your-org/cli/cmd/bootstrap"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "openframe"}
    
    // Add bootstrap command to root
    rootCmd.AddCommand(bootstrap.GetBootstrapCmd())
    
    rootCmd.Execute()
}

// Command line usage examples:
// openframe bootstrap                                    # Interactive mode
// openframe bootstrap my-cluster                        # Custom cluster name
// openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
// openframe bootstrap --non-interactive -v             # CI/CD mode with verbose output
```

The command delegates actual execution to the internal bootstrap service while providing a user-friendly CLI interface with comprehensive help text and flexible configuration options.