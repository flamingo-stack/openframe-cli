<!-- source-hash: f9d3fee8b5cc44b8d1a4b3feb24e8018 -->
Defines the core interfaces for the cluster management domain, establishing contracts for cluster providers, business logic services, and provider registry management.

## Key Components

**ClusterProvider Interface**
- `Create()` - Creates new clusters with specified configuration
- `Delete()` - Removes clusters with optional force deletion
- `Start()` - Starts stopped clusters
- `List()` - Lists all managed clusters
- `Status()` - Gets detailed cluster status information
- `DetectType()` - Identifies cluster type for provider selection
- `GetKubeconfig()` - Retrieves cluster access configuration

**ClusterService Interface**
- Business logic operations for cluster lifecycle management
- Wraps provider operations with application-specific logic
- Methods mirror ClusterProvider but include cluster type parameters

**ProviderRegistry Interface**
- `RegisterProvider()` - Associates cluster types with specific providers
- `GetProvider()` - Retrieves provider by cluster type
- `GetAllProviders()` - Returns all registered providers

## Usage Example

```go
// Register a provider for kind clusters
registry.RegisterProvider(ClusterTypeKind, kindProvider)

// Create a cluster using the service
config := ClusterConfig{Name: "dev-cluster", Version: "1.28"}
service.CreateCluster(ctx, config)

// Get cluster status
info, err := service.GetClusterStatus(ctx, "dev-cluster")
if err != nil {
    log.Fatal(err)
}
```