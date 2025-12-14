<!-- source-hash: b8b8871634348c35bec26710d5f0a24f -->
This file contains comprehensive unit tests for cluster-related data models and configuration structures in a Kubernetes cluster management system.

## Key Components

- **ClusterType Tests** - Validates cluster type constants (k3d, gke) and string conversions
- **ClusterConfig Tests** - Tests cluster configuration structure with name, type, node count, and Kubernetes version
- **ClusterInfo Tests** - Validates cluster information including status, creation time, and node details
- **NodeInfo Tests** - Tests node information structure with name, status, and role fields
- **ProviderOptions Tests** - Validates provider-specific options for K3d and GKE clusters
- **JSON Serialization Tests** - Ensures proper struct tag configuration for JSON marshaling

## Usage Example

```go
// Test cluster configuration creation
func TestClusterSetup(t *testing.T) {
    config := ClusterConfig{
        Name:       "dev-cluster",
        Type:       ClusterTypeK3d,
        NodeCount:  3,
        K8sVersion: "v1.25.0-k3s1",
    }
    
    // Test with K3d-specific options
    options := ProviderOptions{
        K3d: &K3dOptions{
            PortMappings: []string{"8080:80@loadbalancer"},
        },
        Verbose: true,
    }
    
    assert.Equal(t, "dev-cluster", config.Name)
    assert.NotNil(t, options.K3d)
}
```

The tests cover various scenarios including minimal configurations, zero values, different cluster types and statuses, ensuring robust validation of the cluster model structures.