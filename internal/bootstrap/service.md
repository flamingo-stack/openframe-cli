Provides a complete bootstrap service that orchestrates cluster creation and chart installation in a single command, simplifying the development environment setup process.

## Key Components

- **Service**: Main bootstrap service struct
- **NewService()**: Constructor for creating a new bootstrap service instance
- **Execute()**: Command handler that validates flags and orchestrates the bootstrap process
- **bootstrap()**: Core method that executes cluster creation followed by chart installation
- **createClusterSuppressed()**: Creates a Kubernetes cluster with minimal UI output
- **buildClusterConfig()**: Builds cluster configuration with sensible defaults
- **installChartWithMode()**: Installs charts with specified deployment mode settings

## Usage Example

```go
// Create and use bootstrap service
service := bootstrap.NewService()

// Execute bootstrap with cobra command
cmd := &cobra.Command{
    Use: "bootstrap [cluster-name]",
    RunE: service.Execute,
}

// Bootstrap with deployment mode
err := service.Execute(cmd, []string{"my-cluster"})
if err != nil {
    log.Fatal(err)
}

// Non-interactive bootstrap
cmd.Flags().String("deployment-mode", "oss-tenant", "Deployment mode")
cmd.Flags().Bool("non-interactive", true, "Non-interactive mode")
```

The service validates deployment modes (oss-tenant, saas-tenant, saas-shared) and supports both interactive and non-interactive execution modes.