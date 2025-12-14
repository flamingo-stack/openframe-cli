<!-- source-hash: 3bee8cdfb5e8d27b7bd5098b40f6987f -->
This file defines flag structures and utilities for CLI commands in a Kubernetes development tool, providing configuration options for traffic interception and service scaffolding operations.

## Key Components

**InterceptFlags**
- Configures traffic interception parameters including port forwarding, namespace targeting, volume mounting, and header-based filtering
- Supports both global and selective traffic interception modes

**ScaffoldFlags**  
- Defines service scaffolding options for deploying containerized applications
- Includes container configuration, file synchronization, and Kubernetes resource mounting

**AddGlobalFlags()**
- Adds common CLI flags (verbose, silent, dry-run) to cobra commands

## Usage Example

```go
// Setting up intercept command with flags
var interceptFlags InterceptFlags
interceptCmd := &cobra.Command{
    Use: "intercept",
    Run: func(cmd *cobra.Command, args []string) {
        // Use interceptFlags for traffic interception
        setupIntercept(interceptFlags)
    },
}

// Setting up scaffold command with flags
var scaffoldFlags ScaffoldFlags
scaffoldCmd := &cobra.Command{
    Use: "scaffold",
    Run: func(cmd *cobra.Command, args []string) {
        // Use scaffoldFlags for service deployment
        deployService(scaffoldFlags)
    },
}

// Add global flags to root command
rootCmd := &cobra.Command{Use: "dev"}
AddGlobalFlags(rootCmd)
```