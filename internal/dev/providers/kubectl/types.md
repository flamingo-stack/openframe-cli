<!-- source-hash: d9d94f89197243685727b5f846da8812 -->
This file defines JSON data structures for unmarshaling Kubernetes service information from kubectl command output.

## Key Components

- **serviceJSON**: Represents a single Kubernetes service with metadata (name, namespace) and specification details (type, ports configuration)
- **serviceListJSON**: Container structure for multiple services, typically used when parsing `kubectl get services` output

## Usage Example

```go
import (
    "encoding/json"
    "fmt"
)

// Parse a single service
var service serviceJSON
err := json.Unmarshal(serviceData, &service)
if err == nil {
    fmt.Printf("Service: %s in namespace: %s\n", 
        service.Metadata.Name, 
        service.Metadata.Namespace)
}

// Parse multiple services
var services serviceListJSON
err = json.Unmarshal(servicesData, &services)
if err == nil {
    for _, svc := range services.Items {
        fmt.Printf("Found service: %s\n", svc.Metadata.Name)
    }
}
```

These structures handle the JSON output from commands like `kubectl get service -o json` and `kubectl get services -o json`, providing structured access to service metadata and port configurations.