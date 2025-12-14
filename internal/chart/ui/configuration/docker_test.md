<!-- source-hash: c9c11b62910172b6ce23999af257fe22 -->
This file contains comprehensive unit tests for the Docker configuration functionality, ensuring proper handling of Docker registry credentials in Helm chart configurations.

## Key Components

- **TestNewDockerConfigurator**: Tests the constructor for creating a new Docker configurator instance
- **TestDockerConfigurator_Configure_DefaultCredentials**: Verifies handling of default Docker credentials
- **TestDockerConfigurator_Configure_CustomCredentials**: Tests application of custom Docker registry settings
- **TestDockerConfigurator_Configure_WithEmptyValues**: Ensures proper behavior when no existing registry configuration exists
- **TestDockerConfigurator_promptForDockerSettings_Validation**: Validates Docker configuration structure and data integrity
- **TestDockerConfigurator_Configure_NoChangesWhenSameValues**: Tests behavior when user enters identical values
- **TestDockerConfigurator_Configure_SpecialCharacters**: Verifies handling of special characters in credentials
- **TestDockerConfigurator_Configure_EdgeCases**: Tests edge cases like very long strings, single characters, and Unicode

## Usage Example

```go
func TestDockerConfiguration(t *testing.T) {
    // Create modifier and configurator
    modifier := templates.NewHelmValuesModifier()
    configurator := NewDockerConfigurator(modifier)
    
    // Test custom Docker credentials
    config := &types.ChartConfiguration{
        DockerRegistry: &types.DockerRegistryConfig{
            Username: "myuser",
            Password: "mypass",
            Email:    "user@example.com",
        },
        ModifiedSections: []string{"docker"},
        ExistingValues:   map[string]interface{}{},
    }
    
    err := modifier.ApplyConfiguration(config.ExistingValues, config)
    assert.NoError(t, err)
}
```