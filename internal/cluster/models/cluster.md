<!-- source-hash: 522e91fd4fa9512746eea2a2815f35c4 -->
This file defines data models for managing Kubernetes clusters across different providers, supporting both local k3d clusters and Google GKE clusters.

## Key Components

- **ClusterType**: Enum for supported cluster types (`k3d`, `gke`)
- **ClusterConfig**: Configuration structure for creating clusters
- **ClusterInfo**: Complete cluster information including status and nodes
- **NodeInfo**: Individual node details within a cluster
- **ProviderOptions**: Provider-specific configuration options
- **K3dOptions**: Local k3d cluster settings with port mappings
- **GKEOptions**: Google Cloud GKE settings with zone and project

## Usage Example

```go
// Create a local k3d cluster configuration
config := &ClusterConfig{
    Name:       "dev-cluster",
    Type:       ClusterTypeK3d,
    NodeCount:  3,
    K8sVersion: "1.28.0",
}

// Configure provider-specific options
options := &ProviderOptions{
    K3d: &K3dOptions{
        PortMappings: []string{"80:80@loadbalancer", "443:443@loadbalancer"},
    },
    Verbose: true,
}

// Create GKE cluster configuration
gkeConfig := &ClusterConfig{
    Name:       "prod-cluster",
    Type:       ClusterTypeGKE,
    NodeCount:  5,
    K8sVersion: "1.28.0",
}

gkeOptions := &ProviderOptions{
    GKE: &GKEOptions{
        Zone:    "us-central1-a",
        Project: "my-project-id",
    },
}
```