<!-- source-hash: adf0e76343980bd8b6582f63b5924cf7 -->
This package provides automated SSL certificate generation and installation using mkcert for local development environments.

## Key Components

- **CertificateInstaller**: Main struct that handles certificate installation and management
- **NewCertificateInstaller()**: Factory function to create a new installer instance
- **IsInstalled()**: Checks if mkcert is available on the system
- **Install()**: Installs mkcert (if needed) and generates certificates
- **ForceRegenerate()**: Regenerates certificates even if they already exist
- **GetInstallHelp()**: Returns platform-specific installation guidance

## Usage Example

```go
// Create a certificate installer
installer := certificates.NewCertificateInstaller()

// Check if mkcert is already installed
if !installer.IsInstalled() {
    fmt.Println(installer.GetInstallHelp())
    
    // Install mkcert and generate certificates
    if err := installer.Install(); err != nil {
        log.Fatalf("Failed to install certificates: %v", err)
    }
}

// Force regeneration of existing certificates
if err := installer.ForceRegenerate(); err != nil {
    log.Fatalf("Failed to regenerate certificates: %v", err)
}
```

The installer automatically handles platform-specific mkcert installation (macOS via Homebrew, Linux via direct download) and generates localhost certificates in `~/.config/openframe/certs/`. On macOS and Linux, it also configures system trust stores for browser compatibility.