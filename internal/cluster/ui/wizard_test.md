<!-- source-hash: 20faf6cfbd240cc840a3331083b9ec76 -->
Test file for the configuration wizard component that validates cluster setup functionality. Contains comprehensive unit tests for wizard creation, cluster configuration, and utility functions.

## Key Components

- **TestNewConfigWizard**: Tests creation of config wizard with default values (name: "openframe-dev", type: K3d, nodes: 3)
- **TestClusterConfig**: Validates `ClusterConfig` struct creation with different cluster types (K3d, GKE)
- **TestFormatClusterOption**: Tests formatting of cluster information for display (name and status)
- **TestGetClusterNameOrDefault**: Tests cluster name resolution from arguments with fallback logic
- **TestSelectCluster**: Validates cluster selection functionality and error handling for empty lists
- **TestWizardValidation**: Tests validation rules for cluster names and node counts
- **TestConfigWizardState**: Verifies wizard state management and instance independence

## Usage Example

```go
// Test wizard creation with defaults
func TestCustomWizard(t *testing.T) {
    wizard := NewConfigWizard()
    
    assert.Equal(t, "openframe-dev", wizard.config.Name)
    assert.Equal(t, ClusterTypeK3d, wizard.config.Type)
    assert.Equal(t, 3, wizard.config.NodeCount)
}

// Test cluster name resolution
func TestClusterName(t *testing.T) {
    name := GetClusterNameOrDefault([]string{"my-cluster"}, "default")
    assert.Equal(t, "my-cluster", name)
}
```