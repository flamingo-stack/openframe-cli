<!-- source-hash: baf75431fef54d0eaf65ffa772ef4436 -->
Handles ingress configuration for OpenFrame CLI deployments, supporting localhost, Ngrok, and GCP ingress types with interactive setup flows.

## Key Components

- **IngressConfigurator**: Main struct that handles ingress configuration including Ngrok, localhost, and GCP setups
- **NewIngressConfigurator()**: Constructor that creates a new ingress configurator with a Helm values modifier
- **Configure()**: Primary method that prompts users to choose and configure ingress options based on deployment mode
- **configureNgrok()**: Complete Ngrok setup flow including credential collection and IP allowlist configuration
- **applyLocalhostConfig()**: Applies localhost-specific configuration to Helm values
- **applyGCPConfig()**: Applies GCP ingress configuration with tenant ID setup
- **collectNgrokCredentials()**: Interactive credential collection for Ngrok domain, API key, and auth token

## Usage Example

```go
// Create ingress configurator
modifier := templates.NewHelmValuesModifier()
configurator := NewIngressConfigurator(modifier)

// Configure ingress for a chart
config := &types.ChartConfiguration{
    DeploymentMode: &types.DeploymentModeOSS,
    ExistingValues: make(map[string]interface{}),
}

err := configurator.Configure(config)
if err != nil {
    log.Fatal("Failed to configure ingress:", err)
}

// Access the configured ingress settings
fmt.Printf("Ingress type: %s\n", config.IngressConfig.Type)
if config.IngressConfig.NgrokConfig != nil {
    fmt.Printf("Ngrok domain: %s\n", config.IngressConfig.NgrokConfig.Domain)
}
```