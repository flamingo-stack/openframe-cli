<!-- source-hash: 1d577759c57f6301fd396758db5136a5 -->
This file provides interactive configuration for Docker registry credentials in a Helm chart deployment system, allowing users to input or skip Docker authentication settings.

## Key Components

- **DockerConfigurator**: Main struct that handles Docker registry configuration through interactive prompts
- **NewDockerConfigurator()**: Constructor that takes a `HelmValuesModifier` for managing chart values
- **Configure()**: Primary method that presents options to configure Docker credentials or skip them
- **promptForDockerSettings()**: Helper method that collects Docker username, password, and email through interactive input

## Usage Example

```go
// Create a Docker configurator
modifier := templates.NewHelmValuesModifier()
dockerConfig := NewDockerConfigurator(modifier)

// Configure Docker settings interactively
config := &types.ChartConfiguration{
    ExistingValues: existingHelmValues,
}

err := dockerConfig.Configure(config)
if err != nil {
    log.Fatalf("Docker configuration failed: %v", err)
}

// Check if Docker settings were modified
if slices.Contains(config.ModifiedSections, "docker") {
    fmt.Printf("Docker registry configured for user: %s", config.DockerRegistry.Username)
}
```

The configurator presents a choice between skipping Docker credentials or entering custom ones, only updating the configuration when values actually change from existing settings.