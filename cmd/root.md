<!-- source-hash: 6ee9a475e46368b4169e18c499d0fae2 -->
Root command setup for the OpenFrame CLI application, providing the main entry point and command structure for Kubernetes cluster management and development tools.

## Key Components

- **VersionInfo**: Struct holding version, commit, and build date information
- **GetRootCmd()**: Creates the main cobra command with version info
- **Execute()/ExecuteWithVersion()**: Entry points for running the CLI
- **Subcommand getters**: Functions returning cluster, chart, bootstrap, and dev commands

## Usage Example

```go
package main

import (
    "github.com/flamingo-stack/openframe-cli/cmd"
)

func main() {
    // Use default version info
    if err := cmd.Execute(); err != nil {
        log.Fatal(err)
    }
    
    // Or with custom version info
    versionInfo := cmd.VersionInfo{
        Version: "1.0.0",
        Commit:  "abc123",
        Date:    "2024-01-01",
    }
    
    if err := cmd.ExecuteWithVersion(versionInfo); err != nil {
        log.Fatal(err)
    }
}
```

The root command provides global flags (`--verbose`, `--silent`) and initializes the configuration system before executing subcommands for cluster management, chart operations, bootstrapping, and development workflows.