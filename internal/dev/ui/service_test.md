<!-- source-hash: acfa78312e72d2426fc4104e01eacbdd -->
This file contains unit tests for the UI service components, specifically testing service creation, intercept UI access, and intercept setup data structures.

## Key Components

- **TestNewService**: Tests the creation and initialization of a new UI service with proper client dependencies
- **TestService_GetInterceptUI**: Validates retrieval of the intercept UI component from the service
- **TestInterceptSetup_Structure**: Tests the InterceptSetup struct configuration with service and port details

## Usage Example

```go
// Run tests for the UI service
go test ./ui/service_test.go

// Example test setup pattern used in the file
func TestExample(t *testing.T) {
    testutil.InitializeTestMode()
    client := devMocks.NewMockKubernetesClient()
    
    service := NewService(client, client)
    
    assert.NotNil(t, service)
    assert.NotNil(t, service.interceptUI)
}

// InterceptSetup usage pattern
setup := &InterceptSetup{
    ServiceName:    "my-service",
    Namespace:      "production", 
    LocalPort:      3000,
    KubernetesPort: &intercept.ServicePort{
        Name:     "http",
        Port:     8080,
        Protocol: "TCP",
    },
}
```