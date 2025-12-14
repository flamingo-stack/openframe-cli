This file defines the CLI command for bootstrapping a complete OpenFrame environment, combining cluster creation and chart installation into a single streamlined operation.

## Key Components

- **GetBootstrapCmd()**: Returns a configured Cobra command that orchestrates the complete OpenFrame setup process
- **Command flags**: Supports deployment mode selection, non-interactive mode, and verbose logging
- **Arguments**: Accepts an optional cluster name parameter

## Usage Example

```go
// Get the bootstrap command and add it to your CLI
bootstrapCmd := GetBootstrapCmd()
rootCmd.AddCommand(bootstrapCmd)

// The command supports various execution modes:
// Interactive mode (default)
// openframe bootstrap

// With custom cluster name
// openframe bootstrap my-cluster

// Non-interactive with specific deployment mode
// openframe bootstrap --deployment-mode=oss-tenant --non-interactive

// Verbose logging for detailed progress
// openframe bootstrap -v --deployment-mode=saas-shared
```

The bootstrap command internally delegates to `bootstrap.NewService().Execute()` which handles the actual orchestration of cluster creation and chart installation steps.