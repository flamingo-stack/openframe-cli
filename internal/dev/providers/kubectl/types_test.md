<!-- source-hash: 7a34b7c8baa7abb6dd49bc01aa4bedbb -->
Test file that validates JSON unmarshaling for Kubernetes service data structures used by kubectl operations. Ensures compatibility with real kubectl JSON output formats.

## Key Components

- **TestServiceJSON_UnmarshalJSON**: Tests unmarshaling of individual Kubernetes service JSON objects with various configurations (single/multiple ports, numeric target ports, unnamed ports)
- **TestServiceListJSON_UnmarshalJSON**: Tests unmarshaling of service list JSON containing multiple services, including empty lists and malformed JSON handling  
- **TestServiceJSON_RealWorldExamples**: Tests realistic service configurations like web applications, databases, and headless services

## Usage Example

```go
// Example test structure being validated
type serviceJSON struct {
    Metadata struct {
        Name      string `json:"name"`
        Namespace string `json:"namespace"`
    } `json:"metadata"`
    Spec struct {
        Type  string `json:"type"`
        Ports []struct {
            Name       string      `json:"name"`
            Port       int32       `json:"port"`
            Protocol   string      `json:"protocol"`
            TargetPort interface{} `json:"targetPort"`
        } `json:"ports"`
    } `json:"spec"`
}

// Run tests
go test -v ./kubectl/types_test.go
```

The tests ensure that JSON structures can properly handle kubectl output variations like different service types (ClusterIP, NodePort, LoadBalancer), multiple ports, and both string and numeric target port values.