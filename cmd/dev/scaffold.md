Provides a Cobra CLI command for scaffolding development environments with live reloading capabilities using Skaffold. The command handles the complete development lifecycle from prerequisite validation to hot deployment.

## Key Components

- **getScaffoldCmd()**: Creates the scaffold command with comprehensive flag configuration
- **runScaffold()**: Command execution handler that orchestrates the scaffold workflow
- **ScaffoldFlags**: Configuration struct for development environment settings including port, namespace, image, and sync directories

## Usage Example

```go
// Register the scaffold command in your CLI
scaffoldCmd := getScaffoldCmd()
rootCmd.AddCommand(scaffoldCmd)

// Example command usage:
// openframe dev skaffold my-dev-cluster --port 8080 --namespace dev
// openframe dev skaffold --sync-local ./src --sync-remote /app/src
// openframe dev skaffold --skip-bootstrap --helm-values custom-values.yaml
```

The command supports interactive cluster creation, custom ports, namespace targeting, Docker image specification, and file synchronization between local and remote directories. It integrates with the scaffold service to provide a complete development environment with hot reloading capabilities.