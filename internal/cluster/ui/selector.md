<!-- source-hash: 530fdf9e5013b2a813fce353f795ab01 -->
A utility for handling cluster selection logic with support for both command-line arguments and interactive selection. It provides consistent cluster selection behavior across different CLI operations.

## Key Components

- **`Selector`**: Main struct that encapsulates cluster selection logic for a specific operation
- **`NewSelector(operation string)`**: Creates a new selector instance for the given operation
- **`SelectCluster(clusters, args)`**: Selects a single cluster via argument or interactive prompt
- **`SelectMultipleClusters(clusters, args)`**: Handles multi-cluster selection with validation
- **`ValidateClusterExists(clusters, name)`**: Checks if a cluster exists in the provided list
- **`GetClusterByName(clusters, name)`**: Retrieves cluster info by name
- **`FilterClusters(clusters, predicate)`**: Filters clusters based on a predicate function

## Usage Example

```go
// Create selector for deployment operation
selector := NewSelector("deployment")

// Single cluster selection
clusterName, err := selector.SelectCluster(clusters, args)
if err != nil {
    return err
}

// Multi-cluster selection
selectedClusters, err := selector.SelectMultipleClusters(clusters, args)
if err != nil {
    return err
}

// Validate cluster exists
if !selector.ValidateClusterExists(clusters, "my-cluster") {
    return errors.New("cluster not found")
}

// Filter active clusters
activeOnly := selector.FilterClusters(clusters, func(c models.ClusterInfo) bool {
    return c.Status == "active"
})
```