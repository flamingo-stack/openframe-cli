<!-- source-hash: e218bf9bf93107e21d026e54fffe2de5 -->
Handles traffic interception setup for local development by connecting to Kubernetes clusters and routing service traffic to local environments using Telepresence.

## Key Components

- **getInterceptCmd()** - Creates the cobra command with flags for port, namespace, mount options, and traffic filtering
- **runIntercept()** - Main entry point that handles both interactive and flag-based intercept modes
- **runInteractiveIntercept()** - Provides interactive flow with cluster selection and service configuration
- **selectClusterForIntercept()** - Lists and allows selection of available clusters
- **setKubectlContext()** - Switches kubectl context to the selected cluster

## Usage Example

```go
// Interactive mode - prompts for cluster and service selection
err := runInteractiveIntercept(ctx, verbose, dryRun)

// Direct intercept with flags
flags := &models.InterceptFlags{
    Port:      8080,
    Namespace: "default",
    Global:    false,
}
err := runIntercept(cmd, []string{"my-service"}, flags)

// Cluster selection for intercept
clusterName, err := selectClusterForIntercept(verbose)
if err == nil && clusterName != "" {
    err = setKubectlContext(ctx, clusterName, verbose)
}
```

The command supports various intercept modes including header-based filtering, volume mounting, and environment variable loading for comprehensive local development workflows.