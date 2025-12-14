<!-- source-hash: da6133b6d613cd0de62ab6d759615639 -->
This file implements the core intercept creation logic for Telepresence, handling the construction and execution of intercept commands with various configuration options.

## Key Components

- **`createIntercept`** - Main function that builds and executes a Telepresence intercept command with port mapping, mount settings, and optional flags
- **`getRemotePortName`** - Helper function that determines the remote port name, defaulting to the port number if not explicitly specified

## Usage Example

```go
// Create an intercept with basic port mapping
flags := &models.InterceptFlags{
    Port: 8080,
    RemotePortName: "http",
}
err := service.createIntercept(ctx, "my-service", flags)

// Create a global intercept with custom headers and env file
flags := &models.InterceptFlags{
    Port: 3000,
    Global: true,
    EnvFile: ".env.local",
    Header: []string{"x-user-id: 123", "x-debug: true"},
    Replace: true,
}
err := service.createIntercept(ctx, "api-service", flags)
```

The function automatically disables volume mounting and formats port mappings in the `local_port:remote_port_name` format required by Telepresence.