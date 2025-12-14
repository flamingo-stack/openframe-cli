<!-- source-hash: e3ed36297c9d2baea2d4dfa2c006d88a -->
This Go package provides a cross-platform installer and manager for Telepresence, a tool for connecting local development environments to remote Kubernetes clusters. It handles installation, version checking, and daemon initialization across macOS, Linux, and Windows platforms.

## Key Components

- **TelepresenceInstaller**: Main struct providing installation and management functionality
- **IsTelepresenceRunning()**: Global function to check if Telepresence daemon is active
- **NewTelepresenceInstaller()**: Constructor for creating installer instances
- **Install()**: Platform-specific installation with automatic dependency detection
- **IsInstalled()**: Checks if Telepresence is available in system PATH
- **GetVersion()**: Retrieves installed Telepresence version
- **EnsureConfig()**: Initializes Telepresence daemon to handle sudo requirements

## Usage Example

```go
// Create installer and check status
installer := telepresence.NewTelepresenceInstaller()

// Check if already installed
if !installer.IsInstalled() {
    // Get platform-specific installation instructions
    fmt.Println(installer.GetInstallHelp())
    
    // Or install automatically (macOS/Linux only)
    if err := installer.Install(); err != nil {
        log.Fatal(err)
    }
}

// Check version and daemon status
version, _ := installer.GetVersion()
isRunning := telepresence.IsTelepresenceRunning()
fmt.Printf("Telepresence %s, running: %v\n", version, isRunning)

// Initialize daemon if needed
installer.EnsureConfig()
```