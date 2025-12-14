<!-- source-hash: bbb8abffe2e40cbfa13110f6467d72a6 -->
This file contains comprehensive unit tests for configuration-related types and structures in a Go application. It validates the behavior of ingress configurations, Docker registry settings, ngrok configurations, and Helm chart configurations.

## Key Components

- **IngressType Constants Tests** - Validates localhost and ngrok ingress type constants
- **DockerRegistryConfig Tests** - Tests Docker registry configuration creation and validation
- **NgrokConfig Tests** - Comprehensive testing of ngrok tunnel configuration including IP allowlists and registration fields
- **IngressConfig Tests** - Tests ingress configuration for both localhost and ngrok types
- **ChartConfiguration Tests** - Validates Helm chart configuration with minimal and full setups
- **NgrokRegistrationURLs Tests** - Verifies ngrok registration URL constants

## Usage Example

```go
// Testing Docker registry configuration
func TestDockerRegistryConfig_Creation(t *testing.T) {
    config := &DockerRegistryConfig{
        Username: "testuser",
        Password: "testpass",
        Email:    "test@example.com",
    }
    
    assert.Equal(t, "testuser", config.Username)
    assert.Equal(t, "testpass", config.Password)
    assert.Equal(t, "test@example.com", config.Email)
}

// Testing ngrok configuration with IP allowlist
func TestNgrokConfig_WithAllowedIPs(t *testing.T) {
    config := &NgrokConfig{
        AuthToken:     "auth_token_123",
        UseAllowedIPs: true,
        AllowedIPs:    []string{"192.168.1.1", "10.0.0.1"},
    }
    
    assert.True(t, config.UseAllowedIPs)
    assert.Len(t, config.AllowedIPs, 2)
}
```