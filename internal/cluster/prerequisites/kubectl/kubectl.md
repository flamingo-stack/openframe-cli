<!-- source-hash: 8ebc6760390746304f19c01e3b8492a7 -->
A cross-platform kubectl installer that detects the operating system and package manager to automatically install kubectl using the appropriate method for each system.

## Key Components

- **KubectlInstaller**: Main struct providing kubectl installation functionality
- **IsInstalled()**: Checks if kubectl is already installed and functional
- **Install()**: Automatically installs kubectl based on the detected OS and package manager
- **GetInstallHelp()**: Returns platform-specific installation instructions
- **commandExists()**: Utility function to check if a command is available in PATH

The installer supports macOS (via Homebrew), Linux (apt, yum, dnf, pacman, or direct binary), and provides guidance for Windows.

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    installer := kubectl.NewKubectlInstaller()
    
    if installer.IsInstalled() {
        fmt.Println("kubectl is already installed")
        return
    }
    
    fmt.Println("kubectl not found, attempting installation...")
    if err := installer.Install(); err != nil {
        // Fall back to manual instructions
        fmt.Println("Automatic installation failed:")
        fmt.Println(installer.GetInstallHelp())
        log.Fatal(err)
    }
    
    fmt.Println("kubectl installed successfully!")
}
```