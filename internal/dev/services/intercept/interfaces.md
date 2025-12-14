<!-- source-hash: 2b96e793c35a28a9beeb5d15e7fab7f5 -->
Defines core interfaces and data structures for Kubernetes service interception operations, providing abstractions for namespace and service management.

## Key Components

**Data Types:**
- `ServicePort` - Configuration for individual service ports including name, port numbers, and protocol
- `ServiceInfo` - Complete service metadata with name, namespace, type, and port configurations

**Interfaces:**
- `KubernetesClient` - Namespace operations including listing and validation
- `ServiceClient` - Service operations for retrieval and validation across namespaces

## Usage Example

```go
// Implementing the ServiceClient interface
type MyServiceClient struct{}

func (c *MyServiceClient) GetServices(ctx context.Context, namespace string) ([]ServiceInfo, error) {
    return []ServiceInfo{
        {
            Name:      "web-service",
            Namespace: namespace,
            Type:      "ClusterIP",
            Ports: []ServicePort{
                {Name: "http", Port: 80, TargetPort: "8080", Protocol: "TCP"},
            },
        },
    }, nil
}

// Using the interfaces
func validateAndGetService(sc ServiceClient, namespace, serviceName string) (*ServiceInfo, error) {
    ctx := context.Background()
    if err := sc.ValidateService(ctx, namespace, serviceName); err != nil {
        return nil, err
    }
    return sc.GetService(ctx, namespace, serviceName)
}
```