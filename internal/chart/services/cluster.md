<!-- source-hash: 1490e6fba55df1c5c27e4e18964adb6b -->
This file provides cluster selection functionality for chart operations, handling the logic to list, filter, and select Kubernetes clusters through an interactive UI.

## Key Components

- **ClusterSelector**: Main service struct that coordinates cluster listing and selection operations
- **NewClusterSelector()**: Constructor function that initializes a new cluster selector with required dependencies
- **SelectCluster()**: Core method that manages the complete cluster selection workflow, including error handling and verbose logging

## Usage Example

```go
// Initialize cluster selector with dependencies
clusterLister := &myClusterService{}
ui := chartUI.NewOperationsUI()
selector := services.NewClusterSelector(clusterLister, ui)

// Select a cluster for chart operations
args := []string{"my-cluster"} // optional cluster name filter
verbose := true

selectedCluster, err := selector.SelectCluster(args, verbose)
if err != nil {
    log.Fatal(err)
}

if selectedCluster != "" {
    fmt.Printf("Selected cluster: %s\n", selectedCluster)
} else {
    fmt.Println("No cluster selected or available")
}
```

The service handles edge cases like empty cluster lists and provides detailed logging when verbose mode is enabled. It integrates with the chart UI system to provide interactive cluster selection when multiple options are available.