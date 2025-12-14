This service handles the bootstrap command that combines cluster creation and chart installation into a single streamlined operation.

## Key Components

- **Service**: Main struct providing bootstrap functionality
- **NewService()**: Constructor function that creates a new bootstrap service instance
- **Execute()**: Primary command handler that processes CLI arguments and flags
- **bootstrap()**: Core method that orchestrates cluster creation followed by chart installation
- **createClusterSuppressed()**: Creates a Kubernetes cluster with minimal UI output
- **buildClusterConfig()**: Constructs cluster configuration with defaults
- **installChartWithMode()**: Installs charts with specified deployment mode settings

## Usage Example

```go
// Create a new bootstrap service
service := NewService()

// Execute bootstrap with cobra command
err := service.Execute(cmd, []string{"my-cluster"})
if err != nil {
    log.Fatal(err)
}

// Or use the bootstrap method directly
err = service.bootstrap("openframe-dev", "oss-tenant", false, true)
if err != nil {
    log.Fatal(err)
}
```

The service validates deployment modes (oss-tenant, saas-tenant, saas-shared) and enforces that non-interactive mode requires a deployment mode. It defaults to "openframe-dev" cluster name and uses the openframe-oss-tenant repository for chart installation.