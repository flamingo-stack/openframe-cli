<!-- source-hash: 23c8f919488469166c92d602599ca2ae -->
This test file provides comprehensive test coverage for Kubernetes client interfaces and data structures used in the intercept package. It includes mock implementations for testing and validation of service discovery functionality.

## Key Components

- **MockKubernetesClient**: Mock implementation of KubernetesClient interface with methods for namespace operations
- **MockServiceClient**: Mock implementation of ServiceClient interface with methods for service operations  
- **ServicePort**: Data structure representing Kubernetes service port configuration
- **ServiceInfo**: Data structure representing complete Kubernetes service information
- **Interface compliance tests**: Verification that mock implementations satisfy the expected interfaces

## Usage Example

```go
// Testing with mock clients
func TestServiceDiscovery(t *testing.T) {
    ctx := context.Background()
    mockK8s := new(MockKubernetesClient)
    mockSvc := new(MockServiceClient)
    
    // Setup mock expectations
    mockK8s.On("GetNamespaces", ctx).Return([]string{"default", "production"}, nil)
    mockSvc.On("GetServices", ctx, "default").Return([]ServiceInfo{
        {
            Name:      "web-service",
            Namespace: "default", 
            Type:      "ClusterIP",
            Ports: []ServicePort{
                {Name: "http", Port: 80, TargetPort: "8080", Protocol: "TCP"},
            },
        },
    }, nil)
    
    // Test your code using the mocks
    namespaces, err := mockK8s.GetNamespaces(ctx)
    assert.NoError(t, err)
    assert.Contains(t, namespaces, "default")
    
    services, err := mockSvc.GetServices(ctx, "default")
    assert.NoError(t, err)
    assert.Len(t, services, 1)
    
    mockK8s.AssertExpectations(t)
    mockSvc.AssertExpectations(t)
}
```