<!-- source-hash: 91438d79c9fc91a85ed6d143b241aa12 -->
Test file for the configuration package focusing on GitHub Container Registry (GHCR) credential extraction and validation functionality.

## Key Components

- **TestConfigurationWizard_ExtractGHCRCredentials**: Comprehensive test for extracting GHCR credentials from existing chart values, covering various scenarios including missing, default, partial, and valid credentials
- **TestConfigurationWizard_GHCRCredentialsInConfig**: Test ensuring GHCR credentials are properly stored in the configuration structure
- **Test scenarios**: No credentials, existing credentials, default values, empty values, and partial registry structures

## Usage Example

```go
// Run the GHCR credential extraction tests
func TestYourGHCRLogic(t *testing.T) {
    config := &types.ChartConfiguration{
        ExistingValues: map[string]interface{}{
            "registry": map[string]interface{}{
                "ghcr": map[string]interface{}{
                    "username": "myuser",
                    "email":    "user@example.com",
                    "password": "token123",
                },
            },
        },
    }
    
    // Test credential extraction logic
    // Verify username, email, and credential existence flags
    assert.Equal(t, "myuser", extractedUsername)
    assert.True(t, hasCredentials)
}
```

The tests validate the configuration wizard's ability to handle GHCR authentication across different input states and ensure proper credential management for container registry operations.