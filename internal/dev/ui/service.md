<!-- source-hash: 84c966c83fd99fe0f350fc812a5b5af4 -->
This service provides a unified interface for development UI interactions, specifically handling interactive Kubernetes service intercept workflows with user prompts and validation.

## Key Components

- **Service**: Main service struct that orchestrates UI interactions and intercept setup
- **NewService/NewServiceWithExecutor**: Factory functions for creating service instances
- **InteractiveInterceptSetup**: Prompts users for service, port, and local port configuration
- **RunFullInteractiveIntercept**: Complete workflow for setting up service intercepts
- **InterceptSetup**: Configuration struct containing intercept parameters

## Usage Example

```go
// Create service with executor for interactive workflows
executor := &myExecutor{}
service := NewService WithExecutor(executor, true)

// Run full interactive intercept workflow
ctx := context.Background()
err := service.RunFullInteractiveIntercept(ctx)
if err != nil {
    log.Fatal(err)
}

// Or use step-by-step setup
setup, err := service.InteractiveInterceptSetup(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Intercept setup: %s:%d -> %d\n", 
    setup.ServiceName, 
    setup.KubernetesPort.Port, 
    setup.LocalPort)
```

The service handles kubectl validation, cluster connectivity checks, and provides a streamlined interface for setting up development intercepts with proper error handling and optional verbose logging.