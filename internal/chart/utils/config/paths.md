<!-- source-hash: f41c6ea161bee247ef6cdb1eb826af62 -->
Provides centralized path resolution for OpenFrame chart-related files, certificates, and manifests across different execution contexts.

## Key Components

- **PathResolver**: Main struct that handles all path resolution operations
- **NewPathResolver()**: Factory function to create a new path resolver instance
- **GetCertificateDirectory()**: Returns the directory for SSL certificates (`~/.config/openframe/certs`)
- **GetManifestsDirectory()**: Returns path to Kubernetes manifests directory
- **GetHelmValuesFile()**: Returns path to the main Helm values configuration file
- **GetArgocdValuesFile()**: Returns path to ArgoCD-specific values file
- **GetCertificateFiles()**: Returns both certificate and key file paths

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/openframe/config"
)

func main() {
    resolver := config.NewPathResolver()
    
    // Get certificate directory
    certDir := resolver.GetCertificateDirectory()
    fmt.Printf("Certificates stored in: %s\n", certDir)
    
    // Get certificate files
    certFile, keyFile := resolver.GetCertificateFiles()
    fmt.Printf("Cert: %s, Key: %s\n", certFile, keyFile)
    
    // Get Helm values file
    valuesFile := resolver.GetHelmValuesFile()
    fmt.Printf("Helm values: %s\n", valuesFile)
}
```

The resolver automatically creates the certificate directory if it doesn't exist and provides fallback paths for different execution contexts.