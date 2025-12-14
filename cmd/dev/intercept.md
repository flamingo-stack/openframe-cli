Provides traffic interception functionality to route Kubernetes service traffic to local development environments using Telepresence. This module enables developers to debug and test services locally while maintaining connection to the cluster.

## Key Components

- `getInterceptCmd()` - Creates the Cobra command for traffic interception with various configuration flags
- `runIntercept()` - Main entry point handling both interactive and flag-based intercept modes
- `runInteractiveIntercept()` - Interactive flow for cluster selection and service configuration
- `selectClusterForIntercept()` - Cluster selection UI for intercept operations
- `setKubectlContext()` - Switches kubectl context to the selected cluster

## Usage Example

```go
// Flag-based intercept
flags := &models.InterceptFlags{
    Port:      8080,
    Namespace: "default",
    Global:    false,
}
err := runIntercept(cmd, []string{"my-service"}, flags)

// Interactive intercept setup
ctx := context.Background()
err := runInteractiveIntercept(ctx, true, false)

// Set kubectl context for cluster
err := setKubectlContext(ctx, "my-cluster", true)
```

The command supports both interactive mode (no service name) and direct mode with service name and flags for port forwarding, namespace selection, volume mounting, and environment variable loading.