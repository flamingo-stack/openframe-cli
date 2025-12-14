<!-- source-hash: d2917b4157e2e135d745903b9e72a9eb -->
A cross-platform utility for managing jq (JSON processor) installation and verification across different operating systems.

## Key Components

- **`JqInstaller`** - Main struct providing installation and verification methods
- **`NewJqInstaller()`** - Constructor function returning a new installer instance
- **`IsJqRunning()`** - Global function to check if jq is available and functional
- **Core Methods:**
  - `IsInstalled()` - Checks if jq is installed on the system
  - `Install()` - Automatically installs jq based on the detected OS
  - `GetInstallHelp()` - Returns platform-specific installation instructions
  - `GetVersion()` - Retrieves the installed jq version

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    installer := jq.NewJqInstaller()
    
    // Check if jq is installed
    if !installer.IsInstalled() {
        fmt.Println("jq is not installed")
        fmt.Println(installer.GetInstallHelp())
        
        // Attempt automatic installation
        if err := installer.Install(); err != nil {
            log.Fatal(err)
        }
    }
    
    // Verify jq is running
    if jq.IsJqRunning() {
        version, _ := installer.GetVersion()
        fmt.Printf("jq is running: %s\n", version)
    }
}
```

The package supports automatic installation on macOS (via Homebrew), various Linux distributions (apt, yum, dnf, pacman), and provides manual installation guidance for Windows.