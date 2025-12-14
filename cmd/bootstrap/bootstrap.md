This file defines the CLI command interface for the OpenFrame bootstrap functionality, providing a streamlined way to set up a complete OpenFrame environment with a single command.

## Key Components

- **GetBootstrapCmd()** - Returns a configured Cobra command that handles the bootstrap process
- **Command flags**:
  - `--deployment-mode` - Specifies deployment type (oss-tenant, saas-tenant, saas-shared)
  - `--non-interactive` - Enables CI/CD mode without prompts
  - `--verbose` - Shows detailed ArgoCD sync progress
- **bootstrap.NewService()** - Delegates execution to the internal bootstrap service

## Usage Example

```go
// Register the bootstrap command with the root CLI
rootCmd := &cobra.Command{Use: "openframe"}
rootCmd.AddCommand(GetBootstrapCmd())

// Command usage examples:
// openframe bootstrap                                    # Interactive mode
// openframe bootstrap my-cluster                        # Custom cluster name
// openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
// openframe bootstrap --non-interactive --verbose      # CI/CD mode with logs
```

The bootstrap command combines cluster creation and chart installation into a single operation, making it the primary entry point for new OpenFrame deployments. It supports both interactive and automated (CI/CD) workflows through its flag options.