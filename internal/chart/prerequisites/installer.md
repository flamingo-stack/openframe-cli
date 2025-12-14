<!-- source-hash: 1bf48c5468aa0373b40bd513f89dcacf -->
Manages the installation of required prerequisites for the OpenFrame CLI tool. Handles automated installation of missing tools like git, Helm, and certificates, with support for both interactive and non-interactive modes.

## Key Components

- **`Installer`**: Main installer struct that coordinates prerequisite installation
- **`NewInstaller()`**: Creates a new installer instance with a prerequisite checker
- **`CheckAndInstall()`**: Main entry point that checks and installs missing prerequisites
- **`CheckAndInstallNonInteractive()`**: Installation with CI/CD support (non-interactive mode)
- **`RegenerateCertificatesOnly()`**: Standalone certificate regeneration for install command

## Usage Example

```go
// Create installer and check/install prerequisites
installer := prerequisites.NewInstaller()

// Interactive mode (default)
err := installer.CheckAndInstall()
if err != nil {
    log.Fatal(err)
}

// Non-interactive mode (for CI/CD)
err = installer.CheckAndInstallNonInteractive(true)
if err != nil {
    log.Fatal(err)
}

// Regenerate certificates only
err = installer.RegenerateCertificatesOnly()
if err != nil {
    log.Fatal(err)
}
```

The installer handles memory warnings gracefully, provides user confirmation for installations, and continues with available tools in non-interactive mode even if some installations fail.