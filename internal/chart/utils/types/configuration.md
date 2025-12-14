<!-- source-hash: 54b8cd756ead72ef0c60481d9863602f -->
Defines configuration types and constants for managing deployment modes, ingress settings, and chart configurations across different OpenFrame environments.

## Key Components

**Types:**
- `DockerRegistryConfig` - Docker registry authentication settings
- `DeploymentMode` - Enum for OSS, SaaS, or SaaS Shared deployment types
- `IngressType` - Enum for localhost, ngrok, or GCP ingress options
- `NgrokConfig` - Comprehensive ngrok tunnel configuration with IP filtering
- `SaaSConfig` - SaaS-specific repository and branch settings
- `IngressConfig` - Combined ingress configuration wrapper
- `ChartConfiguration` - Complete Helm chart installation configuration

**Constants:**
- Deployment mode constants (`DeploymentModeOSS`, `DeploymentModeSaaS`, etc.)
- Ingress type constants (`IngressTypeLocalhost`, `IngressTypeNgrok`, etc.)
- `NgrokRegistrationURLs` - Static URLs for ngrok service registration

**Functions:**
- `GetRepositoryURL()` - Returns appropriate Git repository URL based on deployment mode

## Usage Example

```go
// Configure for SaaS deployment with ngrok ingress
config := &ChartConfiguration{
    DeploymentMode: &DeploymentModeSaaS,
    IngressConfig: &IngressConfig{
        Type: IngressTypeNgrok,
        NgrokConfig: &NgrokConfig{
            AuthToken: "your-token",
            Domain:    "your-domain.ngrok.io",
            UseAllowedIPs: true,
            AllowedIPs: []string{"192.168.1.0/24"},
        },
    },
}

// Get repository URL for deployment
repoURL := GetRepositoryURL(DeploymentModeSaaS)
```