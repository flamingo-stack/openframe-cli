<!-- source-hash: 47160bdd78e3119be122738d21cba061 -->
This file implements a Telepresence provider for managing Kubernetes service intercepts, enabling local development with remote cluster traffic routing.

## Key Components

- **Provider**: Main struct that manages Telepresence operations with configurable command execution and verbose output
- **NewProvider()**: Factory function to create a new provider instance
- **SetupIntercept()**: Orchestrates the full intercept setup process (installation check, cluster connection, intercept creation, and status display)
- **TeardownIntercept()**: Removes an active intercept for a specified service
- **Disconnect()**: Terminates the Telepresence connection
- **Private methods**: Handle individual operations like installation verification, cluster connection, and intercept creation with detailed flag support

## Usage Example

```go
// Create a new Telepresence provider
executor := &executor.RealCommandExecutor{}
provider := NewProvider(executor, true) // verbose mode

// Set up an intercept with custom flags
flags := &models.InterceptFlags{
    Port:      8080,
    Namespace: "production", 
    Mount:     "/tmp/telepresence",
    Header:    []string{"x-user-id=dev"},
    Global:    false,
}

// Create the intercept
err := provider.SetupIntercept("my-service", flags)
if err != nil {
    log.Fatal(err)
}

// Later, remove the intercept
err = provider.TeardownIntercept("my-service", "production")

// Disconnect when done
provider.Disconnect()
```