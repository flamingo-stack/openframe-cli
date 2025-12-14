<!-- source-hash: 87f70ee901e8953714fc264eedfa4d19 -->
Test file for validating the wizard steps UI component functionality and ensuring proper instantiation and method availability.

## Key Components

- **TestWizardSteps**: Tests basic instantiation of `NewWizardSteps()` function
- **TestWizardSteps_PromptClusterType**: Validates the cluster type prompt method exists
- **TestWizardSteps_PromptK8sVersion**: Verifies the Kubernetes version prompt method is available
- **TestWizardSteps_ConfirmConfiguration**: Tests configuration confirmation method with valid cluster config

## Usage Example

```go
// Run tests to validate wizard steps functionality
func TestExample(t *testing.T) {
    steps := NewWizardSteps()
    
    // Verify methods exist and don't panic
    assert.NotNil(t, steps.PromptClusterType)
    assert.NotNil(t, steps.PromptK8sVersion)
    
    // Test with valid cluster configuration
    config := models.ClusterConfig{
        Name:       "test-cluster",
        Type:       models.ClusterTypeK3d,
        NodeCount:  3,
        K8sVersion: "latest",
    }
    
    assert.NotPanics(t, func() {
        steps.ConfirmConfiguration(config)
    })
}
```

This test suite focuses on structural validation rather than interactive testing, ensuring all wizard step methods are properly defined and handle valid configurations without errors.