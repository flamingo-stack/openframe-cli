<!-- source-hash: 01d629272a856f40fdc8a176cb678bef -->
A test suite that validates the proper initialization and structure of the `ConfigurationWizard` constructor and its internal components.

## Key Components

- **TestNewConfigurationWizard**: Tests that `NewConfigurationWizard()` creates a wizard with all required components properly initialized
- **TestConfigurationWizard_Structure**: Validates the internal structure and non-nil state of all wizard components
- **TestConfigurationWizard_Components**: Ensures multiple wizard instances are independent with separate component instances

## Usage Example

```go
// Run the tests to validate wizard initialization
func TestWizardCreation(t *testing.T) {
    wizard := NewConfigurationWizard()
    
    // The wizard and its components should be properly initialized
    assert.NotNil(t, wizard)
    assert.NotNil(t, wizard.modifier)
    assert.NotNil(t, wizard.branchConfig)
    assert.NotNil(t, wizard.dockerConfig)
    assert.NotNil(t, wizard.ingressConfig)
}

// Test multiple wizard instances are independent
func TestMultipleWizards(t *testing.T) {
    wizard1 := NewConfigurationWizard()
    wizard2 := NewConfigurationWizard()
    
    // Each wizard should be a separate instance
    assert.NotSame(t, wizard1, wizard2)
    assert.NotSame(t, wizard1.modifier, wizard2.modifier)
}
```