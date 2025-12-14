<!-- source-hash: 15f3586f94abe4223cc7da82de314f75 -->
Defines core interfaces for the scaffold package, establishing contracts for bootstrap operations, prerequisite validation, and Skaffold command execution.

## Key Components

- **BootstrapService**: Interface for executing bootstrap commands with arguments
- **PrerequisiteChecker**: Interface for validating tool installation status and providing installation assistance
- **ScaffoldRunner**: Interface for running Skaffold operations (dev, build, deploy) with context support

## Usage Example

```go
// Implementing a prerequisite checker
type DockerChecker struct{}

func (d *DockerChecker) IsInstalled() bool {
    // Check if Docker is installed
    return true
}

func (d *DockerChecker) GetInstallHelp() string {
    return "Install Docker from https://docker.com"
}

func (d *DockerChecker) Install() error {
    // Installation logic
    return nil
}

func (d *DockerChecker) GetVersion() (string, error) {
    return "20.10.0", nil
}

// Using the scaffold runner
var runner ScaffoldRunner
ctx := context.Background()
err := runner.RunDev(ctx, []string{"--port-forward"})
```