<!-- source-hash: ba50d774958bd631aec56e2b4c5c72bd -->
This file provides a complete Skaffold development workflow service for Kubernetes applications, handling everything from prerequisite validation to running development sessions with proper cluster integration.

## Key Components

- **Service**: Main service struct managing Skaffold workflows with command execution, kubectl integration, and signal handling
- **NewService()**: Constructor creating a new scaffold service with executor and verbose options
- **RunScaffoldWorkflow()**: Core method orchestrating the complete development workflow including service selection, cluster setup, chart installation, and Skaffold execution
- **checkPrerequisites()**: Validates Skaffold installation with automatic installation prompts
- **runSkaffoldDev()**: Executes Skaffold dev commands with retry logic and graceful shutdown handling

## Usage Example

```go
// Create executor and service
executor := executor.NewCommandExecutor()
service := NewService(executor, true) // verbose mode enabled

// Define scaffold flags
flags := &models.ScaffoldFlags{
    SkipBootstrap:   false,
    HelmValuesFile:  "values.yaml",
    Namespace:       "development",
}

// Run complete workflow
ctx := context.Background()
args := []string{"my-cluster"} // optional cluster name
err := service.RunScaffoldWorkflow(ctx, args, flags)
if err != nil {
    log.Fatal(err)
}

// Check if service is running
if service.IsRunning() {
    service.Stop()
}
```

The service integrates cluster management, chart installation, and Skaffold development workflows with comprehensive error handling and user interaction.