This file implements the scaffold command for the OpenFrame CLI's development environment, providing hot-reloading development workflows with Skaffold integration.

## Key Components

- **`getScaffoldCmd()`** - Returns a configured Cobra command for scaffolding development environments
- **`runScaffold()`** - Executes the scaffold workflow with proper service initialization
- **Configuration flags** - Port, namespace, image, sync directories, and bootstrap options
- **Service integration** - Uses scaffold service for workflow execution with command executor

## Usage Example

```go
// Add the scaffold command to the dev command group
devCmd.AddCommand(getScaffoldCmd())

// Command usage examples:
// openframe dev skaffold                    # Interactive setup
// openframe dev skaffold my-dev-cluster    # Named cluster
// openframe dev skaffold --port 8080       # Custom port
// openframe dev skaffold --skip-bootstrap  # Skip cluster setup
// openframe dev skaffold --namespace dev   # Specific namespace
```

The scaffold command provides a complete development environment setup by validating Skaffold prerequisites, bootstrapping clusters with development-friendly settings, and enabling live code reloading capabilities. It supports both interactive and scripted workflows through flexible flag configuration.