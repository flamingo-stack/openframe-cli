<!-- source-hash: bd396ae8c2eeab772669db8c38308c3c -->
This file implements the cluster creation command for a Kubernetes cluster management CLI tool. It provides both interactive and non-interactive modes for creating local development clusters with customizable configurations.

## Key Components

- **`getCreateCmd()`** - Creates and configures the Cobra command with flags, validation, and help text
- **`runCreateCluster()`** - Main execution logic that handles interactive/non-interactive modes and delegates to services
- **Interactive Mode** - Uses UI components for step-by-step cluster configuration
- **Non-Interactive Mode** - Creates clusters directly from CLI flags with sensible defaults
- **Configuration Validation** - Validates cluster names, node counts, and other parameters

## Usage Example

```go
// Create command instance
createCmd := getCreateCmd()

// Interactive mode (default)
// openframe cluster create
// openframe cluster create my-cluster

// Non-interactive mode with flags
// openframe cluster create --skip-wizard --nodes 3 --type k3d

// Dry run to preview configuration
// openframe cluster create --dry-run --nodes 5

// The command handles:
// - Cluster name validation
// - Node count validation (minimum 1, default 3)
// - Cluster type selection (defaults to k3d)
// - Configuration summary display
// - Service delegation for actual cluster creation
```