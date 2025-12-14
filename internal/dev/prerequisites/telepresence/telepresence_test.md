<!-- source-hash: 1897b5cc8b500984af76f6e0fcb9d532 -->
Test suite for the Telepresence installer functionality, validating installation detection, version checking, and help message generation.

## Key Components

- **TestNewTelepresenceInstaller**: Verifies proper instantiation of the installer
- **TestTelepresenceInstaller_GetInstallHelp**: Validates help message content and URLs
- **TestTelepresenceInstaller_IsInstalled**: Tests installation detection logic
- **TestIsTelepresenceRunning**: Checks runtime status detection
- **TestTelepresenceInstaller_GetVersion**: Validates version retrieval with proper error handling
- **TestCommandExists**: Utility function testing for command availability
- **TestTelepresenceInstallHelp**: Tests standalone help message generation

## Usage Example

```go
func TestYourTelepresenceFeature(t *testing.T) {
    installer := NewTelepresenceInstaller()
    
    // Test installation status
    if installer.IsInstalled() {
        version, err := installer.GetVersion()
        assert.NoError(t, err)
        assert.NotEmpty(t, version)
    } else {
        help := installer.GetInstallHelp()
        assert.Contains(t, help, "telepresence.io")
    }
    
    // Test runtime status
    isRunning := IsTelepresenceRunning()
    assert.IsType(t, true, isRunning)
}
```

The test suite handles different system states gracefully, checking for installation status before testing version-dependent functionality and avoiding actual system modifications during test execution.