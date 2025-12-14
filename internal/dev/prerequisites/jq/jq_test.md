<!-- source-hash: f0616558d7e87371eb01f834ad59facf -->
This file contains comprehensive unit tests for the jq installer functionality, validating command-line tool detection, installation checks, and platform-specific behaviors.

## Key Components

- **TestNewJqInstaller** - Tests constructor for jq installer instance
- **TestJqInstaller_IsInstalled** - Validates jq installation detection without system dependencies
- **TestJqInstaller_GetVersion** - Tests version retrieval with conditional assertions based on installation status
- **TestIsJqRunning** - Tests standalone function for checking if jq process is active
- **TestCommandExists** - Validates generic command existence checking with known and unknown commands
- **TestJqInstaller_GetInstallHelp** - Ensures installation help text contains relevant information
- **TestJqInstaller_OSDetection** - Tests platform detection methods (Debian/RHEL-like systems)
- **TestJqInstaller_Install** - Verifies install method structure without executing system changes

## Usage Example

```go
func TestCustomJqBehavior(t *testing.T) {
    installer := NewJqInstaller()
    
    // Test installation status
    if installer.IsInstalled() {
        version, err := installer.GetVersion()
        assert.NoError(t, err)
        assert.NotEmpty(t, version)
    }
    
    // Test help functionality
    help := installer.GetInstallHelp()
    assert.Contains(t, help, "jq")
    
    // Test command detection
    assert.True(t, commandExists("echo"))
}
```

The tests are designed to be non-destructive, avoiding actual system modifications while thoroughly validating the installer's detection and informational capabilities.