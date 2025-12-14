<!-- source-hash: 4715ded5ad318c605c52a66b5e4b0e60 -->
This file implements functionality to wait for ArgoCD applications to reach a healthy and synced state during chart installation. It provides robust cancellation handling, progress monitoring, and detailed logging capabilities.

## Key Components

- **`WaitForApplications()`** - Main function that waits for all ArgoCD applications to become healthy and synced
- **Context management** - Handles cancellation via context, interrupt signals, and timeouts
- **Progress monitoring** - Shows spinner with timer and verbose status updates
- **Error handling** - Graceful handling of timeouts, cancellation, and parsing errors
- **Debug functionality** - Detailed pod diagnostics for stuck applications after extended wait times

## Usage Example

```go
// Basic usage within a Manager instance
manager := &Manager{executor: executor}
config := config.ChartInstallConfig{
    Verbose: true,
    Silent:  false,
    DryRun:  false,
}

ctx := context.Background()
err := manager.WaitForApplications(ctx, config)
if err != nil {
    log.Printf("Failed to wait for applications: %v", err)
}

// With timeout context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()

err = manager.WaitForApplications(ctx, config)
```

The function includes a 30-second bootstrap phase, monitors applications every 2 seconds, supports graceful cancellation via Ctrl+C, and provides detailed debugging information for stuck applications after 7 minutes.