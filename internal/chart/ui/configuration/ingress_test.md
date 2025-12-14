<!-- source-hash: b84cb7d0acd99e74123e597c516071d5 -->
Test suite for the IngressConfigurator component, validating localhost and ngrok ingress configuration functionality for OpenFrame CLI's Helm chart management.

## Key Components

- **Constructor Tests**: Validates `NewIngressConfigurator` initialization with HelmValuesModifier
- **Localhost Configuration**: Tests local ingress setup with proper YAML structure
- **Ngrok Configuration**: Validates tunnel configuration with credentials, domains, and IP allowlists
- **Configuration Switching**: Tests transitions between localhost and ngrok ingress types
- **Settings Detection**: Validates current ingress type detection from Helm values
- **Validation Tests**: Ensures proper credential validation and error handling

## Usage Example

```go
// Test localhost configuration
func TestLocalhostConfig(t *testing.T) {
    modifier := templates.NewHelmValuesModifier()
    configurator := NewIngressConfigurator(modifier)
    
    values := map[string]interface{}{}
    err := configurator.applyLocalhostConfig(values)
    assert.NoError(t, err)
    
    // Verify localhost ingress is enabled
    ingress := values["deployment"].(map[string]interface{})["oss"].(map[string]interface{})["ingress"]
    localhost := ingress.(map[string]interface{})["localhost"].(map[string]interface{})
    assert.True(t, localhost["enabled"].(bool))
}

// Test ngrok configuration with IP restrictions
func TestNgrokWithIPs(t *testing.T) {
    config := &types.NgrokConfig{
        AuthToken:     "token_123",
        APIKey:        "key_456", 
        Domain:        "example.ngrok-free.app",
        UseAllowedIPs: true,
        AllowedIPs:    []string{"192.168.1.1"},
    }
    
    values := map[string]interface{}{}
    err := configurator.applyNgrokConfig(values, config)
    assert.NoError(t, err)
}
```