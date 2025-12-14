<!-- source-hash: 3fd65dbd1a1735d357012bac15f105ad -->
Provides comprehensive Docker installation and management utilities with cross-platform support for detecting, installing, and starting Docker on macOS, Linux, and Windows systems.

## Key Components

**Core Functions:**
- `IsDockerRunning()` - Checks if Docker daemon is accessible
- `IsDockerInstalledButNotRunning()` - Detects Docker installation without running daemon
- `StartDocker()` - Attempts to start Docker service/application
- `WaitForDocker()` - Waits for Docker daemon to become available

**DockerInstaller Struct:**
- `IsInstalled()` - Checks if Docker command exists
- `GetInstallHelp()` - Returns platform-specific installation instructions
- `Install()` - Automatically installs Docker using platform package managers

**Platform-Specific Installation:**
- macOS: Homebrew with Docker Desktop
- Linux: Support for apt, yum, dnf, and pacman package managers
- Windows: Installation guidance (manual process)

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Check Docker status
    if !docker.IsDockerRunning() {
        fmt.Println("Docker is not running")
        
        // Try to start Docker
        if err := docker.StartDocker(); err != nil {
            log.Printf("Failed to start Docker: %v", err)
        }
        
        // Wait for Docker to be ready
        if err := docker.WaitForDocker(); err != nil {
            log.Printf("Docker failed to start: %v", err)
        }
    }
    
    // Install Docker if not present
    installer := docker.NewDockerInstaller()
    if !installer.IsInstalled() {
        fmt.Println(installer.GetInstallHelp())
        if err := installer.Install(); err != nil {
            log.Printf("Installation failed: %v", err)
        }
    }
}
```