<!-- source-hash: 819e40cb98e1a51d54bd7f86618e4ed8 -->
Defines the Cobra CLI command for scaffolding development environments with live reloading capabilities using Skaffold. This command provides a complete development workflow including cluster bootstrapping and hot deployment setup.

## Key Components

- **`getScaffoldCmd()`** - Creates the main scaffold command with comprehensive flags for development configuration
- **`runScaffold()`** - Executes the scaffold workflow by initializing the scaffold service and running the development environment setup
- **`models.ScaffoldFlags`** - Configuration structure containing port, namespace, image, sync paths, and bootstrap options
- **CLI Flags** - Port configuration, Kubernetes namespace, Docker image, file sync paths, bootstrap control, and Helm values

## Usage Example

```go
// Basic scaffold command with default settings
openframe dev skaffold

// Scaffold with specific cluster name and custom port
openframe dev skaffold my-dev-cluster --port 3000

// Advanced configuration with file sync and custom namespace
openframe dev skaffold \
  --namespace dev-env \
  --sync-local ./src \
  --sync-remote /app/src \
  --helm-values custom-values.yaml

// Skip cluster bootstrap for existing clusters
openframe dev skaffold existing-cluster --skip-bootstrap
```

The command integrates with Skaffold for live reloading, manages cluster prerequisites, and provides flexible configuration options for various development scenarios.