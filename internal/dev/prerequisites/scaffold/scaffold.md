<!-- source-hash: 5e7c1d65e474df994da31252b5d6497d -->
This package provides a cross-platform installer for Skaffold, a tool that facilitates continuous development for Kubernetes applications. It handles detection, installation, and version management of Skaffold across different operating systems.

## Key Components

- **ScaffoldInstaller**: Main struct that provides installation and management functionality
- **IsScaffoldRunning()**: Global function to check if Skaffold is available and working
- **NewScaffoldInstaller()**: Constructor for creating a new installer instance
- **IsInstalled()**: Checks if Skaffold is installed on the system
- **Install()**: Automatically installs Skaffold using platform-specific methods
- **GetInstallHelp()**: Provides OS-specific installation instructions
- **GetVersion()**: Retrieves the installed Skaffold version

## Usage Example

```go
package main

import (
    "fmt"
    "scaffold"
)

func main() {
    installer := scaffold.NewScaffoldInstaller()
    
    if !installer.IsInstalled() {
        fmt.Println("Installing Skaffold...")
        if err := installer.Install(); err != nil {
            fmt.Printf("Installation failed: %v\n", err)
            fmt.Println("Manual installation:", installer.GetInstallHelp())
            return
        }
    }
    
    version, err := installer.GetVersion()
    if err != nil {
        fmt.Printf("Error getting version: %v\n", err)
        return
    }
    
    fmt.Printf("Skaffold version: %s\n", version)
}
```