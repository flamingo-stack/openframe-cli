<!-- source-hash: a2211d3c9c2214455f7bb571b9bf2ddc -->
This file provides cross-platform Helm installation capabilities with automatic detection and installation methods for different operating systems.

## Key Components

- **HelmInstaller**: Main struct that handles Helm installation operations
- **NewHelmInstaller()**: Constructor function that returns a new HelmInstaller instance
- **IsInstalled()**: Checks if Helm is already installed on the system
- **GetInstallHelp()**: Returns platform-specific installation instructions
- **Install()**: Automatically installs Helm using the appropriate method for the current OS
- **commandExists()**: Utility function to check if a command is available in PATH

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    installer := NewHelmInstaller()
    
    // Check if Helm is already installed
    if installer.IsInstalled() {
        fmt.Println("Helm is already installed")
        return
    }
    
    // Get installation help for manual installation
    fmt.Println(installer.GetInstallHelp())
    
    // Attempt automatic installation
    if err := installer.Install(); err != nil {
        log.Printf("Automatic installation failed: %v", err)
        fmt.Println("Please install manually using the instructions above")
    } else {
        fmt.Println("Helm installed successfully")
    }
}
```

The installer supports macOS (via Homebrew), various Linux distributions (Ubuntu, Red Hat, Fedora, Arch), and provides installation guidance for Windows.