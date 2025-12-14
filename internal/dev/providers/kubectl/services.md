<!-- source-hash: 14b9a73d457be224483ab742ac932f42 -->
A Go package that provides Kubernetes service management operations through kubectl commands, with JSON parsing and fallback mechanisms for reliable service discovery.

## Key Components

- **GetServices()** - Retrieves all services in a namespace with JSON parsing and simple fallback
- **GetService()** - Fetches details for a specific service by name
- **ValidateService()** - Checks if a service exists in a given namespace
- **getServicesSimple()** - Fallback method using basic kubectl commands when JSON parsing fails
- **convertJSONToServiceInfo()** - Converts kubectl JSON output to structured ServiceInfo objects

## Usage Example

```go
// Get all services in a namespace
services, err := provider.GetServices(ctx, "default")
if err != nil {
    log.Fatal(err)
}

// Get specific service details
service, err := provider.GetService(ctx, "default", "my-service")
if err != nil {
    log.Fatal(err)
}

// Validate service exists before operations
err = provider.ValidateService(ctx, "default", "my-service")
if err != nil {
    log.Printf("Service not found: %v", err)
}
```

The package automatically handles kubectl JSON output parsing with graceful fallback to simpler commands if parsing fails, ensuring robust service discovery across different Kubernetes environments.