This file implements the Skaffold development command for the OpenFrame CLI, providing hot reloading and live development capabilities for Kubernetes services.

## Key Components

- `getScaffoldCmd()` - Creates the Cobra command with comprehensive flags for development configuration
- `runScaffold()` - Command execution handler that orchestrates the scaffold workflow
- `models.ScaffoldFlags` - Configuration structure for scaffold parameters (port, namespace, image, sync paths)
- `scaffoldService.Service` - Core service handling the development environment setup

## Usage Example

```go
// Basic scaffold command setup
cmd := getScaffoldCmd()

// Example flag configuration
flags := &models.ScaffoldFlags{
    Port: 8080,
    Namespace: "dev",
    SyncLocal: "./src",
    SyncRemote: "/app",
    SkipBootstrap: false,
}

// Run scaffold workflow
ctx := context.Background()
exec := executor.NewRealCommandExecutor(false, true)
service := scaffoldService.NewService(exec, true)
err := service.RunScaffoldWorkflow(ctx, []string{"my-cluster"}, flags)
```

The command supports various development scenarios from basic cluster scaffolding to advanced file synchronization with custom Docker images and Helm configurations.