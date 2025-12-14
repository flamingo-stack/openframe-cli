<!-- source-hash: 23eda37396f9fc1812fa101ab054a0c6 -->
This service provides a clean abstraction layer for displaying cluster-related information and UI operations in the OpenFrame CLI. It separates presentation logic from business logic by handling all cluster UI display functionality.

## Key Components

- **ClusterDisplayInfo** - Struct containing cluster information for display (name, type, status, nodes, creation time)
- **NodeDisplayInfo** - Struct for individual node display data (name, role, status)
- **ClusterConfigDisplay** - Struct for cluster configuration summary display
- **DisplayService** - Main service struct that handles all UI display operations
- **ShowClusterList()** - Renders a formatted table of clusters with status colors
- **ShowClusterStatus()** - Displays detailed status information for a single cluster
- **ShowConfigurationSummary()** - Shows cluster configuration summary with dry-run support

## Usage Example

```go
// Create display service
displayService := ui.NewDisplayService()

// Show cluster creation progress
displayService.ShowClusterCreationStart("my-cluster", "kind", os.Stdout)

// Display cluster list
clusters := []ClusterDisplayInfo{
    {
        Name: "dev-cluster",
        Type: "kind", 
        Status: "running",
        NodeCount: 3,
        CreatedAt: time.Now(),
    },
}
displayService.ShowClusterList(clusters, os.Stdout)

// Show detailed status
displayService.ShowClusterStatus(&clusters[0], os.Stdout)

// Display configuration summary
config := &ClusterConfigDisplay{
    Name: "prod-cluster",
    Type: "kind",
    K8sVersion: "1.28.0",
    NodeCount: 5,
}
displayService.ShowConfigurationSummary(config, false, false, os.Stdout)
```