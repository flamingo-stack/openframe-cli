<!-- source-hash: 80638bc9b630d3ad28e05fac2b6e0c02 -->
This file provides system-level configuration and initialization services for the OpenFrame deployment system, handling log directory setup and system-wide settings management.

## Key Components

- **`SystemService`** - Main service struct that manages system configuration and log directory paths
- **`NewSystemService()`** - Constructor that creates a service instance with default temp directory logging
- **`NewSystemServiceWithOptions()`** - Constructor allowing custom log directory specification
- **`Initialize()`** - Performs system setup tasks including log directory creation
- **`GetLogDirectory()`** - Returns the configured log directory path

## Usage Example

```go
// Create system service with default settings
sysService := config.NewSystemService()

// Initialize the system (creates log directories)
if err := sysService.Initialize(); err != nil {
    log.Fatal("System initialization failed:", err)
}

// Get the log directory for use in other components
logPath := sysService.GetLogDirectory()
fmt.Println("Logs will be written to:", logPath)

// Create with custom log directory
customService := config.NewSystemServiceWithOptions("/var/log/openframe")
if err := customService.Initialize(); err != nil {
    log.Fatal("Custom system setup failed:", err)
}
```