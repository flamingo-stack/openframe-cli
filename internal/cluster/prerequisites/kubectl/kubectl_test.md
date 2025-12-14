<!-- source-hash: 169587ea4811fa1acc7e4489186e15b6 -->
Test suite for the kubectl installer package, providing comprehensive validation of kubectl installation functionality across different operating systems.

## Key Components

- **TestNewKubectlInstaller**: Validates that the kubectl installer can be created successfully
- **TestKubectlInstaller_GetInstallHelp**: Tests platform-specific installation help text generation for macOS, Linux, and Windows
- **TestKubectlInstaller_Install**: Verifies error handling for automatic installation across different platforms and scenarios
- **TestCommandExists**: Tests the utility function that checks for command availability
- **containsSubstring**: Helper function for substring matching in test assertions

## Usage Example

```go
func TestKubectlInstaller_GetInstallHelp(t *testing.T) {
    installer := NewKubectlInstaller()
    help := installer.GetInstallHelp()
    
    if help == "" {
        t.Error("Install help should not be empty")
    }
    
    switch runtime.GOOS {
    case "darwin":
        if !containsSubstring(help, "brew") {
            t.Error("macOS help should contain brew reference")
        }
    case "linux":
        if !containsSubstring(help, "package manager") {
            t.Error("Linux help should contain package manager reference")
        }
    }
}
```

The tests validate cross-platform behavior, ensuring appropriate error messages and installation guidance are provided based on the operating system and available tools like Homebrew on macOS.