<!-- source-hash: a294c9f3548f62c5a2a8f711d1d211be -->
This service provides a streamlined bootstrap command that combines cluster creation and chart installation into a single operation for setting up OpenFrame environments.

## Key Components

- **Service**: Main bootstrap service struct
- **NewService()**: Constructor for creating a new bootstrap service instance
- **Execute()**: Primary command handler that processes CLI arguments and flags
- **bootstrap()**: Core orchestration method that sequences cluster creation and chart installation
- **createClusterSuppressed()**: Creates a Kubernetes cluster with minimal UI output
- **installChartWithMode()**: Installs charts with specified deployment mode configuration
- **buildClusterConfig()**: Constructs cluster configuration with sensible defaults

## Usage Example

```go
// Create and use the bootstrap service
service := NewService()

// Execute bootstrap command (typically called by cobra)
cmd := &cobra.Command{
    Use: "bootstrap",
}
cmd.Flags().String("deployment-mode", "", "Deployment mode")
cmd.Flags().Bool("non-interactive", false, "Non-interactive mode")
cmd.Flags().Bool("verbose", false, "Verbose output")

err := service.Execute(cmd, []string{"my-cluster"})
if err != nil {
    log.Fatal(err)
}

// Or use the core bootstrap functionality directly
err = service.bootstrap("my-cluster", "oss-tenant", false, true)
```

The service validates deployment modes (oss-tenant, saas-tenant, saas-shared) and enforces that non-interactive mode requires a deployment mode specification.