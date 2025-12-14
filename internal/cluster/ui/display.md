<!-- source-hash: a3578862cf9d36e9ec862d6fa237051a -->
This file provides UI display functions for the OpenFrame CLI, primarily wrapping shared UI components and offering specialized cluster-related display functionality.

## Key Components

**Deprecated Functions (wrap shared UI):**
- `GetStatusColor()` - Returns color function for status strings
- `RenderTableWithFallback()` - Renders tables with fallback to simple output
- `ShowSuccessBox()` - Displays formatted success messages
- `FormatAge()` - Formats time durations as human-readable age strings

**Active Functions:**
- `RenderOverviewTable()` - Renders cluster overview information
- `RenderNodeTable()` - Renders node information tables
- `ShowClusterCreationNextSteps()` - Displays post-cluster-creation guidance

## Usage Example

```go
// Display cluster creation next steps
ShowClusterCreationNextSteps("my-cluster")

// Render cluster overview
overviewData := pterm.TableData{
    {"Cluster Name", "my-cluster"},
    {"Status", "Running"},
    {"Nodes", "3"},
}
err := RenderOverviewTable(overviewData)

// Render node table
nodeData := pterm.TableData{
    {"NAME", "STATUS", "ROLES", "AGE"},
    {"master-1", "Ready", "control-plane", "5m"},
    {"worker-1", "Ready", "worker", "4m"},
}
err = RenderNodeTable(nodeData)
```

Most functions are deprecated and redirect to shared UI components. The primary active functionality focuses on cluster-specific display operations with graceful fallbacks for rendering issues.