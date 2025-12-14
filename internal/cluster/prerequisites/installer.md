<!-- source-hash: 9308ea254d41325ac8690317ac2b0fa6 -->
Manages the installation and setup of required prerequisites (Docker, kubectl, k3d) for cluster operations. Provides both interactive and non-interactive installation modes with user prompts and progress feedback.

## Key Components

- **Installer**: Main struct that orchestrates prerequisite installation
- **NewInstaller()**: Creates a new installer instance with prerequisite checker
- **InstallMissingPrerequisites()**: Installs all missing tools automatically
- **CheckAndInstall()**: Interactive mode - checks and prompts user for installation
- **CheckAndInstallNonInteractive()**: Non-interactive mode for automated environments
- **installTool()**: Installs specific tools (docker, kubectl, k3d)
- **showManualInstructions()**: Displays manual installation help when auto-install is declined

## Usage Example

```go
// Interactive installation
installer := prerequisites.NewInstaller()
err := installer.CheckAndInstall()
if err != nil {
    log.Fatal(err)
}

// Non-interactive installation (for CI/automation)
err = installer.CheckAndInstallNonInteractive(true)
if err != nil {
    log.Fatal(err)
}

// Install only missing prerequisites
err = installer.InstallMissingPrerequisites()
if err != nil {
    log.Fatal(err)
}
```

The installer handles Docker startup detection separately from installation and provides platform-specific manual instructions when automated installation fails or is declined.