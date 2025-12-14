Defines the CLI command structure for the OpenFrame bootstrap functionality, which provides a streamlined way to set up a complete OpenFrame environment by combining cluster creation and chart installation in a single operation.

## Key Components

- **GetBootstrapCmd()** - Returns a configured cobra.Command for the bootstrap operation
- **Command flags**:
  - `deployment-mode` - Specifies deployment type (oss-tenant, saas-tenant, saas-shared)
  - `non-interactive` - Enables CI/CD mode by skipping prompts
  - `verbose` - Shows detailed logging including ArgoCD sync progress

## Usage Example

```go
// Get the bootstrap command for CLI registration
bootstrapCmd := GetBootstrapCmd()

// The command supports various usage patterns:
// openframe bootstrap                                    # Interactive mode
// openframe bootstrap my-cluster                        # Custom cluster name
// openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
// openframe bootstrap --non-interactive --verbose      # Full automation with detailed logs
```

The bootstrap command internally delegates execution to `bootstrap.NewService().Execute()`, which handles the actual bootstrapping logic including cluster creation and chart installation.