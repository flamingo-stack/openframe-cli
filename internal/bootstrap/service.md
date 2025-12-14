This service orchestrates the complete OpenFrame bootstrap process by combining cluster creation and chart installation in a single command.

## Key Components

- **Service**: Main service struct providing bootstrap functionality
- **NewService()**: Factory function to create a new bootstrap service instance
- **Execute()**: Command handler that validates flags and coordinates the bootstrap process
- **bootstrap()**: Core orchestration method that sequentially creates cluster and installs charts
- **createClusterSuppressed()**: Creates a Kubernetes cluster with minimal UI output
- **installChartWithMode()**: Installs OpenFrame charts with specified deployment configuration
- **buildClusterConfig()**: Builds cluster configuration with sensible defaults

## Usage Example

```go
// Create and use the bootstrap service
service := NewService()

// Execute bootstrap with command and arguments
cmd := &cobra.Command{}
args := []string{"my-cluster"}
err := service.Execute(cmd, args)
if err != nil {
    log.Fatal(err)
}

// The service will:
// 1. Create a k3d cluster named "my-cluster"
// 2. Install OpenFrame charts with specified deployment mode
// 3. Handle verbose logging and non-interactive modes
```

The service supports three deployment modes: `oss-tenant`, `saas-tenant`, and `saas-shared`, with automatic validation and error handling throughout the process.