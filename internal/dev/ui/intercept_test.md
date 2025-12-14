<!-- source-hash: 513afcd0d1aae6659245a5d2d9f09f1f -->
This file contains comprehensive unit tests for the InterceptUI component, which handles user interface logic for service interception functionality in a CLI development tool.

## Key Components

- **TestNewInterceptUI** - Tests the constructor for InterceptUI, verifying proper initialization with Kubernetes and service clients
- **TestInterceptUI_findServiceInCluster** - Tests service discovery functionality across different namespaces with various port configurations
- **TestInterceptUI_validatePort** - Tests port validation logic for user input, covering edge cases and error conditions
- **ServiceInfo struct** - Test data structure containing service metadata (name, namespace, ports, found status)

## Usage Example

```go
// Run the tests
go test ./internal/ui/intercept_test.go

// Test service discovery
func ExampleTestServiceDiscovery() {
    testutil.InitializeTestMode()
    client := devMocks.NewMockKubernetesClient()
    ui := NewInterceptUI(client, client)
    
    serviceInfo, err := ui.findServiceInCluster(context.Background(), "my-api")
    if err == nil && serviceInfo.Found {
        fmt.Printf("Service %s found in namespace %s\n", 
            serviceInfo.Name, serviceInfo.Namespace)
    }
}

// Test port validation
func ExamplePortValidation() {
    ui := NewInterceptUI(client, client)
    if err := ui.validatePort("8080"); err != nil {
        log.Fatal("Invalid port")
    }
}
```