<!-- source-hash: a21fd1e56e3049574a9c9bdb34ac036e -->
Test suite for the `HelmValuesModifier` component that validates YAML file operations, configuration parsing, and value manipulation for Helm charts in the OpenFrame CLI.

## Key Components

- **`NewHelmValuesModifier()`** - Tests constructor for Helm values modifier
- **`LoadExistingValues()`** - Tests loading and parsing YAML files with validation for file not found, invalid YAML, and empty files
- **`GetCurrentOSSBranch()`** - Tests extraction of OSS branch configuration with fallback to defaults
- **`GetCurrentDockerSettings()`** - Tests retrieval of Docker registry configuration
- **`ApplyConfiguration()`** - Tests applying chart configuration changes for branch and Docker settings
- **`WriteValues()`** - Tests writing modified values back to YAML files

## Usage Example

```go
func TestHelmValuesModifier_LoadExistingValues(t *testing.T) {
    modifier := NewHelmValuesModifier()
    
    // Test loading existing YAML values
    values, err := modifier.LoadExistingValues(testFile)
    assert.NoError(t, err)
    
    // Verify parsed structure
    global := values["global"].(map[string]interface{})
    assert.Equal(t, "main", global["repoBranch"])
}

func TestHelmValuesModifier_ApplyConfiguration_Branch(t *testing.T) {
    // Apply new branch configuration
    config := &types.ChartConfiguration{
        Branch: &newBranch,
        DeploymentMode: &deploymentMode,
        ModifiedSections: []string{"branch"},
    }
    
    err := modifier.ApplyConfiguration(values, config)
    assert.NoError(t, err)
}
```