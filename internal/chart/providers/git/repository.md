<!-- source-hash: 699e3ee5c67a48f7f80386f780e0fd99 -->
This package provides Git repository operations for cloning and managing chart repositories, specifically designed for handling GitHub repositories containing Helm charts.

## Key Components

- **Repository**: Main struct that handles git operations using a command executor
- **NewRepository**: Constructor function that creates a new Repository instance
- **CloneChartRepository**: Clones a GitHub repository to a temporary directory with optimizations for speed (depth 1, single branch)
- **Cleanup**: Utility method for removing temporary directories
- **CloneResult**: Return type containing temporary directory path and chart path information

## Usage Example

```go
package main

import (
    "context"
    "github.com/flamingo-stack/openframe-cli/internal/git"
    "github.com/flamingo-stack/openframe-cli/internal/chart/models"
    "github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

func main() {
    // Create repository handler
    exec := executor.NewCommandExecutor()
    repo := git.NewRepository(exec)
    
    // Configure repository settings
    config := &models.AppOfAppsConfig{
        GitHubRepo:   "https://github.com/user/charts-repo",
        GitHubBranch: "main",
        ChartPath:    "charts/my-app",
    }
    
    // Clone repository
    ctx := context.Background()
    result, err := repo.CloneChartRepository(ctx, config)
    if err != nil {
        panic(err)
    }
    
    // Clean up when done
    defer repo.Cleanup(result.TempDir)
    
    // Use result.ChartPath for chart operations
}
```