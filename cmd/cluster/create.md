This file implements the `create` command for creating new Kubernetes clusters in the OpenFrame CLI, handling both interactive and non-interactive cluster creation workflows.

## Key Components

- **`getCreateCmd()`** - Creates and configures the cobra command for cluster creation with flags and validation
- **`runCreateCluster()`** - Main execution logic that handles interactive UI configuration or flag-based configuration
- **Interactive Mode** - Uses UI configuration handler for step-by-step cluster setup
- **Non-interactive Mode** - Builds cluster config directly from command flags and arguments
- **Validation** - Validates cluster names, node counts, and other configuration parameters

## Usage Example

```go
// Interactive cluster creation with selection menu
openframe cluster create

// Create cluster with custom name (shows selection menu)
openframe cluster create my-cluster

// Non-interactive creation with defaults
openframe cluster create --skip-wizard

// Create with specific configuration
openframe cluster create my-cluster --nodes 3 --type k3d --skip-wizard

// Dry run to preview configuration
openframe cluster create --dry-run --nodes 5
```

The command supports both interactive wizards for guided setup and direct flag-based configuration for automation. It validates inputs, shows configuration summaries, and delegates actual cluster creation to the service layer.