<!-- source-hash: 7a6724730ad9d0b7b91b67dd36e8cf00 -->
This file provides pre-configured Helm chart values for deploying ArgoCD with custom health checks and timeout settings.

## Key Components

- **GetArgoCDValues()** - Returns a YAML string containing ArgoCD Helm chart configuration values

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/yourorg/argocd"
)

func main() {
    // Get ArgoCD Helm values for deployment
    values := argocd.GetArgoCDValues()
    
    // Use with Helm deployment
    fmt.Println("ArgoCD Values:")
    fmt.Println(values)
    
    // Could be written to values.yaml file
    // or passed directly to helm install/upgrade commands
}
```

The returned YAML includes:
- Custom health check configuration for ArgoCD Application resources
- Repository server timeout settings (180s)
- Fullname override for consistent naming
- Commented example for image pull secrets

This is typically used when programmatically deploying ArgoCD via Helm, allowing consistent configuration across environments.