Provides Kubernetes traffic interception capabilities using Telepresence to route cluster traffic to local development environments.

## Key Components

- **`getInterceptCmd()`** - Creates the `intercept` cobra command with flags for port, namespace, mounting, and traffic filtering
- **`runIntercept()`** - Main entry point that handles both interactive and flag-based intercept modes
- **`runInteractiveIntercept()`** - Implements interactive flow with cluster selection and service discovery
- **`selectClusterForIntercept()`** - Handles cluster selection using the cluster UI service
- **`setKubectlContext()`** - Switches kubectl context to the selected cluster
- **`models.InterceptFlags`** - Configuration structure for intercept parameters

## Usage Example

```go
// Interactive mode - prompts for cluster and service selection
cmd := getInterceptCmd()
err := cmd.Execute() // openframe dev intercept

// Flag-based mode with specific service
flags := &models.InterceptFlags{
    Port:      8080,
    Namespace: "production",
    Global:    false,
    Header:    []string{"x-user-id=123"},
}
err := runIntercept(cmd, []string{"my-service"}, flags)

// Set kubectl context for specific cluster
ctx := context.Background()
err := setKubectlContext(ctx, "my-cluster", true)
```

The command supports both interactive service discovery and direct service targeting with flexible traffic filtering options including header-based routing and volume mounting capabilities.