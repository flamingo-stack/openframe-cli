<!-- source-hash: 28c77d29ead94ef48049dc9c41704780 -->
Provides a user interface for setting up Kubernetes service intercepts, guiding users through service selection, port configuration, and validation.

## Key Components

- **InterceptUI**: Main struct that handles user interactions for intercept setup
- **ServiceInfo**: Contains service details including name, namespace, ports, and discovery status
- **PromptForService()**: Interactive service name input with cluster-wide validation
- **PromptForKubernetesPort()**: Port selection from available service ports
- **PromptForLocalPort()**: Local port configuration for traffic forwarding
- **findServiceInCluster()**: Searches for services across all namespaces
- **validatePort()**: Port number validation (1-65535 range)

## Usage Example

```go
// Initialize the intercept UI
kubernetesClient := // ... your kubernetes client
serviceClient := // ... your service client
interceptUI := NewInterceptUI(kubernetesClient, serviceClient)

// Guide user through intercept setup
ctx := context.Background()

// Get service information
serviceInfo, err := interceptUI.PromptForService(ctx)
if err != nil {
    log.Fatal(err)
}

// Select Kubernetes port to intercept
kubernetesPort, err := interceptUI.PromptForKubernetesPort(serviceInfo.Ports)
if err != nil {
    log.Fatal(err)
}

// Get local port for forwarding
localPort, err := interceptUI.PromptForLocalPort(kubernetesPort)
if err != nil {
    log.Fatal(err)
}

// Now you have: serviceInfo.Name, serviceInfo.Namespace, kubernetesPort, localPort
```