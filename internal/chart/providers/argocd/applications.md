<!-- source-hash: ccf7aea70ca5df5489c22627aa045ba8 -->
This file provides ArgoCD application management functionality for monitoring and tracking application deployments in Kubernetes clusters.

## Key Components

**Manager**: Main struct that handles ArgoCD operations using a command executor interface.

**Application**: Struct representing an ArgoCD application with name, health status, and sync status.

**Key Methods**:
- `NewManager(executor.CommandExecutor)`: Creates a new ArgoCD manager instance
- `getTotalExpectedApplications(ctx, config)`: Estimates total number of applications using multiple detection methods
- `parseApplications(ctx, verbose)`: Retrieves current ArgoCD applications and their statuses via kubectl

## Usage Example

```go
import (
    "context"
    "github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// Create ArgoCD manager
exec := executor.NewCommandExecutor()
manager := NewManager(exec)

// Get current applications
ctx := context.Background()
apps, err := manager.parseApplications(ctx, true)
if err != nil {
    log.Fatal(err)
}

// Display application status
for _, app := range apps {
    fmt.Printf("App: %s, Health: %s, Sync: %s\n", 
        app.Name, app.Health, app.Sync)
}

// Get expected total count
config := config.ChartInstallConfig{Verbose: true}
total := manager.getTotalExpectedApplications(ctx, config)
fmt.Printf("Expected applications: %d\n", total)
```

The manager uses multiple fallback strategies to detect applications, making it resilient to various ArgoCD deployment states and configurations.