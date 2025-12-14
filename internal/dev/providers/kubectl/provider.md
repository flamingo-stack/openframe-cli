<!-- source-hash: a2f191f843f98f22e96bc4e71f564ea1 -->
This file implements a kubectl provider that wraps kubectl command execution for Kubernetes operations, serving as a bridge between the application and kubectl CLI.

## Key Components

- **Provider**: Main struct that implements Kubernetes and Service client interfaces using kubectl commands
- **NewProvider()**: Constructor function that creates a new kubectl provider instance
- **CheckConnection()**: Validates kubectl cluster connectivity
- **GetCurrentContext()**: Retrieves the active kubectl context
- **SetContext()**: Switches between different kubectl contexts
- **parseTargetPort()**: Helper method for parsing port values from different data types

## Usage Example

```go
// Create a new kubectl provider
exec := executor.NewCommandExecutor()
provider := kubectl.NewProvider(exec, true)

// Check cluster connectivity
ctx := context.Background()
if err := provider.CheckConnection(ctx); err != nil {
    log.Fatal("No cluster connection:", err)
}

// Get current context
context, err := provider.GetCurrentContext(ctx)
if err != nil {
    log.Fatal("Failed to get context:", err)
}
fmt.Printf("Current context: %s\n", context)

// Switch to different context
err = provider.SetContext(ctx, "production")
if err != nil {
    log.Fatal("Failed to switch context:", err)
}
```

The provider abstracts kubectl operations and implements both `KubernetesClient` and `ServiceClient` interfaces for use with the intercept service.