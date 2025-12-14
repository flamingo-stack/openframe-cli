<!-- source-hash: 3eecc2e0721c49b7cab4729603f66246 -->
Handles graceful cleanup and signal management for Telepresence intercept operations, ensuring proper resource cleanup on interrupt signals.

## Key Components

- **setupCleanupHandler()** - Configures signal handlers for SIGINT and SIGTERM to trigger cleanup
- **cleanup()** - Performs orderly shutdown including leaving intercepts, stopping daemon, and restoring namespace

## Usage Example

The cleanup functionality is automatically integrated into the service:

```go
// Signal handling is set up during service initialization
service := &Service{
    signalChannel: make(chan os.Signal, 1),
    verbose: true,
}

// Setup cleanup handler for graceful shutdown
service.setupCleanupHandler("my-service")

// When Ctrl+C is pressed or SIGTERM received:
// 1. Stops current intercept
// 2. Quits telepresence daemon  
// 3. Restores original namespace
// 4. Exits cleanly
```

The cleanup process runs with a 30-second timeout and handles errors gracefully, displaying appropriate success/warning messages based on the verbose flag setting.