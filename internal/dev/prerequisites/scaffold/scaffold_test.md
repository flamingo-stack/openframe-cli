<!-- source-hash: 47378d4a350ddb23d5eb4cf095517116 -->
Test suite for the scaffold package that validates Skaffold installation detection and management functionality.

## Key Components

- **TestNewScaffoldInstaller**: Verifies the `ScaffoldInstaller` constructor creates a valid instance
- **TestScaffoldInstaller_GetInstallHelp**: Tests installation help text generation and validates it contains platform-specific installation methods
- **TestScaffoldInstaller_IsInstalled**: Checks the installation detection functionality without making assumptions about the test environment
- **TestIsScaffoldRunning**: Validates the standalone function for detecting running Skaffold processes
- **TestScaffoldInstaller_GetVersion**: Tests version retrieval with conditional assertions based on installation status
- **TestCommandExists**: Utility test for command existence detection using known commands
- **TestScaffoldInstaller_Install**: Structure validation for the install method (without executing actual installation)

## Usage Example

```go
func TestScaffoldDetection(t *testing.T) {
    installer := NewScaffoldInstaller()
    
    // Check if Skaffold is available
    if installer.IsInstalled() {
        version, err := installer.GetVersion()
        assert.NoError(t, err)
        assert.NotEmpty(t, version)
    }
    
    // Verify help contains installation instructions
    help := installer.GetInstallHelp()
    assert.Contains(t, help, "skaffold")
}
```

The tests use conditional logic to handle different system states, ensuring they pass regardless of whether Skaffold is actually installed on the test machine.