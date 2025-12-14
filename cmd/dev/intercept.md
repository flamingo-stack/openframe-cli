Provides traffic interception functionality using Telepresence to route Kubernetes service traffic to local development environments.

## Key Components

- **getInterceptCmd()** - Creates the cobra command with flags for port, namespace, mount options, headers, and other intercept settings
- **runIntercept()** - Main handler that routes to interactive or flag-based intercept modes based on arguments
- **runInteractiveIntercept()** - Interactive flow with cluster selection, kubectl context switching, and UI-driven service configuration
- **selectClusterForIntercept()** - Handles cluster discovery and selection using the cluster UI service
- **setKubectlContext()** - Switches kubectl context to the selected k3d cluster

## Usage Example

```go
// Interactive mode - prompts for cluster and service selection
cmd := getInterceptCmd()
cmd.RunE(cmd, []string{}, &models.InterceptFlags{})

// Direct service intercept with flags
flags := &models.InterceptFlags{
    Port:      8080,
    Namespace: "default",
    Global:    false,
    Header:    []string{"x-user=dev"},
}
runIntercept(cmd, []string{"my-service"}, flags)

// Set kubectl context for cluster
ctx := context.Background()
err := setKubectlContext(ctx, "my-cluster", true)
```

The command supports both interactive cluster/service selection and direct flag-based invocation for automated workflows. It integrates with the cluster management system and provides comprehensive Telepresence lifecycle management.