<!-- source-hash: c3e91a824b00bb96fd9a5ad5683eeba0 -->
Defines data models for Helm chart management, providing structured representations of chart information and types.

## Key Components

- **ChartInfo**: Struct containing metadata about an installed Helm chart including name, namespace, status, version, and application version
- **ChartType**: String type for categorizing different chart types
- **Chart Type Constants**:
  - `ChartTypeArgoCD`: Represents ArgoCD charts
  - `ChartTypeAppOfApps`: Represents app-of-apps pattern charts

## Usage Example

```go
// Create chart information
chartInfo := models.ChartInfo{
    Name:       "my-app",
    Namespace:  "default",
    Status:     "deployed",
    Version:    "1.0.0",
    AppVersion: "2.3.1",
}

// Use chart types
var chartType models.ChartType = models.ChartTypeArgoCD

// Check chart type
if chartType == models.ChartTypeAppOfApps {
    // Handle app-of-apps pattern
}
```