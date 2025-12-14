<!-- source-hash: 0f7b26ad415328189f8b8fa37e2ece52 -->
Provides standardized message templating and formatting functionality for CLI applications. Uses predefined templates with consistent styling and emoji indicators for different message types.

## Key Components

- **MessageType**: Enum constants for different message categories (Info, Success, Warning, Error, Progress, Completion)
- **Templates**: Main struct containing predefined message templates organized by type and operation
- **NewTemplates()**: Factory function creating a new Templates instance with standard message formats
- **FormatMessage()**: Core method for formatting messages using template strings and arguments
- **Show* methods**: High-level methods for displaying different message types (ShowInfo, ShowSuccess, ShowError, etc.)
- **CustomTemplates**: Extended struct allowing custom template additions
- **Formatter**: Convenience wrapper providing specialized formatters for installations and clusters

## Usage Example

```go
package main

import (
    "errors"
    "time"
)

func main() {
    templates := NewTemplates()
    
    // Show operation start
    templates.ShowOperationStart("deployment", "production-cluster")
    
    // Show progress
    templates.ShowProgress("health_check", "nginx-service")
    
    // Show completion with duration
    templates.ShowStepComplete("database migration", 2*time.Minute)
    
    // Show error with troubleshooting
    err := errors.New("connection timeout")
    templates.ShowOperationFailed("database backup", err)
    
    // Use specialized formatters
    formatter := NewFormatter()
    formatter.Installation().Starting("prometheus", "dev-cluster")
    formatter.Installation().Complete("prometheus", []string{
        "Access dashboard at http://localhost:9090",
        "Configure alerts in config/alerts.yml",
    })
}
```