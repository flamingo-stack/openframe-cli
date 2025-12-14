<!-- source-hash: d57818b349c7aec4474ea7e3443c50c5 -->
Provides functionality for detecting, installing, and managing k3d (a lightweight Kubernetes distribution) across different operating systems. The package includes cross-platform installation support with automatic detection of package managers and fallback mechanisms.

## Key Components

- **K3dInstaller**: Main struct that provides k3d installation and management functionality
- **NewK3dInstaller()**: Creates a new K3dInstaller instance
- **IsInstalled()**: Checks if k3d is already installed on the system
- **GetInstallHelp()**: Returns platform-specific installation instructions
- **Install()**: Automatically installs k3d based on the detected operating system
- **commandExists()**: Utility function to check if a command is available in PATH
- **Platform-specific installers**: Separate methods for macOS, Linux distributions, and Windows

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    installer := k3d.NewK3dInstaller()
    
    // Check if k3d is already installed
    if installer.IsInstalled() {
        fmt.Println("k3d is already installed")
        return
    }
    
    // Get installation help
    fmt.Println(installer.GetInstallHelp())
    
    // Attempt automatic installation
    if err := installer.Install(); err != nil {
        log.Printf("Failed to install k3d: %v", err)
        fmt.Println("Please install manually using:", installer.GetInstallHelp())
    } else {
        fmt.Println("k3d installed successfully")
    }
}
```