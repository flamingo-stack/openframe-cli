<!-- source-hash: dc3d78e5aeea8ccb920a4c0b813bd2d4 -->
Provides interactive wizard steps for cluster configuration, handling user prompts and validation for cluster creation parameters.

## Key Components

- **WizardSteps**: Main struct containing wizard step implementations
- **PromptClusterName()**: Collects and validates cluster name input
- **PromptClusterType()**: Presents cluster type selection (k3d/GKE)
- **PromptNodeCount()**: Prompts for worker node count with range validation
- **PromptK8sVersion()**: Allows Kubernetes version selection from predefined options
- **ConfirmConfiguration()**: Displays configuration summary and requests confirmation
- **renderConfigurationTable()**: Renders styled configuration table with fallback

## Usage Example

```go
// Create wizard steps handler
wizard := NewWizardSteps()

// Prompt for cluster configuration
name, err := wizard.PromptClusterName("my-cluster")
if err != nil {
    return err
}

clusterType, err := wizard.PromptClusterType()
if err != nil {
    return err
}

nodeCount, err := wizard.PromptNodeCount(3)
if err != nil {
    return err
}

k8sVersion, err := wizard.PromptK8sVersion()
if err != nil {
    return err
}

// Create config and confirm
config := models.ClusterConfig{
    Name: name,
    Type: clusterType,
    NodeCount: nodeCount,
    K8sVersion: k8sVersion,
}

confirmed, err := wizard.ConfirmConfiguration(config)
if err != nil || !confirmed {
    return err
}
```