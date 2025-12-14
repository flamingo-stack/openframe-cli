<!-- source-hash: de57b3ed183049fbaa88378a7289eab6 -->
This test file validates the functionality of a tool installer that manages prerequisite software installation for development environments.

## Key Components

- **TestNewInstaller**: Validates proper initialization of the installer instance and its internal checker component
- **TestInstallTool**: Tests tool installation delegation for supported tools (docker, kubectl, k3d) and error handling for unknown tools
- **TestRunCommand**: Verifies command execution functionality using a simple echo command
- **containsSubstring**: Helper function for string matching in error validation

## Usage Example

```go
func TestCustomInstaller(t *testing.T) {
    installer := NewInstaller()
    
    // Test valid tool installation
    err := installer.installTool("docker")
    if err != nil {
        t.Logf("Installation failed as expected in test env: %v", err)
    }
    
    // Test invalid tool
    err = installer.installTool("nonexistent-tool")
    if err == nil {
        t.Error("Expected error for unknown tool")
    }
    
    // Test command execution
    err = installer.runCommand("ls", "-la")
    if err != nil {
        t.Errorf("Command failed: %v", err)
    }
}
```

The tests focus on validating installer behavior in controlled environments, expecting installation failures but ensuring proper error handling and tool recognition logic work correctly.