<!-- source-hash: 0d045822685aa15b78f7ffdaabdaa6bd -->
Test suite for the configuration builder component that validates proper instantiation and configuration building functionality for OpenFrame CLI installations.

## Key Components

- **TestNewBuilder**: Tests builder instantiation with and without UI components
- **TestBuilder_ImplementsConfigBuilderInterface**: Validates interface compliance  
- **TestBuilder_BuildInstallConfig_***: Comprehensive test suite covering various configuration scenarios including basic setup, flag combinations, GitHub repository integration, and certificate directory handling
- **TestBuilder_ComponentsInitialized**: Verifies proper component initialization
- **TestBuilder_MultipleBuilds**: Ensures builder statelessness across multiple configurations

## Usage Example

```go
func TestBasicBuilderUsage(t *testing.T) {
    // Create builder with operations UI
    operationsUI := chartUI.NewOperationsUI()
    builder := NewBuilder(operationsUI)
    
    // Build install configuration
    config, err := builder.BuildInstallConfig(
        false, false, false,      // force, dryRun, verbose flags
        "my-cluster",             // cluster name
        "https://github.com/my/repo", // GitHub repo
        "main",                   // branch
        "/certs",                // certificate directory
    )
    
    assert.NoError(t, err)
    assert.Equal(t, "my-cluster", config.ClusterName)
    assert.True(t, config.HasAppOfApps())
}
```

The tests validate builder functionality across different scenarios including flag combinations, GitHub integration, certificate handling, and ensure the builder remains stateless between multiple configuration builds.