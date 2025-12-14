<!-- source-hash: bfeec34d7421e870281c6f6524f90b54 -->
This file provides interactive wizards and configuration handlers for Kubernetes cluster creation in the OpenFrame CLI tool. It manages both quick default setups and step-by-step interactive configuration flows.

## Key Components

**ClusterConfig** - Configuration structure holding cluster name, type, node count, and Kubernetes version

**ConfigWizard** - Interactive wizard that guides users through cluster configuration with customizable defaults

**ConfigurationHandler** - Main handler that orchestrates the configuration flow, offering both quick and interactive modes

**SelectCluster()** - Utility function for interactive cluster selection from a list of available clusters

**GetClusterNameOrDefault()** - Helper function that resolves cluster names from command arguments with fallback defaults

## Usage Example

```go
// Create and run interactive configuration wizard
wizard := NewConfigWizard()
wizard.SetDefaults("my-cluster", models.ClusterTypeK3d, 3, "v1.28")
config, err := wizard.Run()
if err != nil {
    log.Fatal(err)
}

// Use configuration handler for complete flow
handler := NewConfigurationHandler()
clusterConfig, err := handler.GetClusterConfig("my-cluster")
if err != nil {
    log.Fatal(err)
}

// Select from existing clusters
clusters := []models.ClusterInfo{{Name: "dev", Status: "running"}}
selected, err := SelectCluster(clusters, "Choose cluster")
if err != nil {
    log.Fatal(err)
}
```