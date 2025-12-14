<!-- source-hash: 04f239e56dbb0a7e420b9217a7a8f672 -->
A utility for reading, modifying, and writing Helm values files with support for different deployment modes (OSS, SaaS, SaaS Shared) and configuration management.

## Key Components

- **HelmValuesModifier**: Main struct for managing Helm values operations
- **LoadExistingValues()**: Reads and parses existing Helm values from file
- **LoadOrCreateBaseValues()**: Loads values from current directory or creates empty defaults
- **ApplyConfiguration()**: Applies deployment mode, branch, and registry configurations
- **CreateTemporaryValuesFile()**: Creates temporary values file for operations
- **WriteValues()**: Serializes and writes values back to YAML file
- **Getter methods**: Extract current settings (branch, docker, ingress, deployment mode)

## Usage Example

```go
// Create modifier and load existing values
modifier := NewHelmValuesModifier()
values, err := modifier.LoadOrCreateBaseValues()
if err != nil {
    log.Fatal(err)
}

// Apply configuration changes
config := &types.ChartConfiguration{
    DeploymentMode: &types.DeploymentModeOSS,
    Branch:         stringPtr("develop"),
    DockerRegistry: &types.DockerRegistryConfig{
        Username: "myuser",
        Password: "mypass",
        Email:    "user@example.com",
    },
}

err = modifier.ApplyConfiguration(values, config)
if err != nil {
    log.Fatal(err)
}

// Write updated values
err = modifier.WriteValues(values, "helm-values.yaml")
```