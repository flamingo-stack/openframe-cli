<!-- source-hash: c74918819ef6c7ca9e973b22c9d0293e -->
This file provides namespace-related operations for the kubectl provider, including retrieving, validating, and automatically discovering namespaces for Kubernetes resources.

## Key Components

- **`GetNamespaces()`** - Retrieves all namespaces in the current cluster
- **`ValidateNamespace()`** - Verifies if a specified namespace exists
- **`FindResourceNamespace()`** - Automatically discovers the namespace containing resources matching a service name
- **`searchResourceByType()`** - Internal helper that searches for specific resource types across all namespaces

## Usage Example

```go
// Get all available namespaces
namespaces, err := provider.GetNamespaces(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Available namespaces:", namespaces)

// Validate a specific namespace exists
err = provider.ValidateNamespace(ctx, "production")
if err != nil {
    log.Printf("Namespace validation failed: %v", err)
}

// Auto-discover namespace for a service
namespace, err := provider.FindResourceNamespace(ctx, "my-service")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Service found in namespace: %s", namespace)
```

The `FindResourceNamespace()` function is particularly useful for automatically locating services across namespaces by searching deployments, statefulsets, and pods with matching names or prefixes.