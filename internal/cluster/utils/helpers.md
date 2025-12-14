<!-- source-hash: 1488d052895b0db2cac1e587c7f676fd -->
This file provides utility functions for cluster management operations, including validation, type parsing, and error handling.

## Key Components

- **ClusterSelectionResult** - Deprecated struct for cluster selection results
- **ValidateClusterName()** - Validates cluster name format using domain validation
- **ParseClusterType()** - Converts string input to ClusterType enum (k3d, gke)
- **GetNodeCount()** - Returns validated node count with default fallback of 3
- **CreateClusterError()** - Creates standardized cluster operation errors

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/flamingo-stack/openframe-cli/internal/utils"
    "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

func main() {
    // Validate cluster name
    if err := utils.ValidateClusterName("my-cluster"); err != nil {
        fmt.Printf("Invalid cluster name: %v\n", err)
    }

    // Parse cluster type from user input
    clusterType := utils.ParseClusterType("gke") // Returns models.ClusterTypeGKE

    // Get validated node count
    nodeCount := utils.GetNodeCount(0) // Returns 3 (default)
    nodeCount = utils.GetNodeCount(5)  // Returns 5

    // Create standardized error
    err := utils.CreateClusterError("create", "my-cluster", clusterType, fmt.Errorf("connection failed"))
    fmt.Printf("Error: %v\n", err)
}
```