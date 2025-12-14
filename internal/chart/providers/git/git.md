<!-- source-hash: e2b269d483da6f40ce954628aa522d87 -->
This file defines data structures for Git operations, specifically for handling the results of Git clone operations in the context of Helm chart management.

## Key Components

- **`CloneResult`** - A struct that encapsulates the outcome of a git clone operation, containing paths to the temporary directory and the chart location within the cloned repository.

## Usage Example

```go
package main

import (
    "fmt"
    "path/filepath"
)

func handleCloneResult(result *git.CloneResult) {
    // Access the temporary directory where repo was cloned
    fmt.Printf("Repository cloned to: %s\n", result.TempDir)
    
    // Access the specific chart path within the repo
    chartName := filepath.Base(result.ChartPath)
    fmt.Printf("Chart found at: %s (chart: %s)\n", result.ChartPath, chartName)
    
    // Cleanup would typically involve removing TempDir
    // when done processing the chart
}

func processChart() {
    result := &git.CloneResult{
        TempDir:   "/tmp/helm-clone-12345",
        ChartPath: "/tmp/helm-clone-12345/charts/my-app",
    }
    
    handleCloneResult(result)
}
```