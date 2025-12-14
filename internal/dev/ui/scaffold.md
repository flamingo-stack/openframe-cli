<!-- source-hash: 45ab00c9b0945c9fb5515d45b82a3ed9 -->
Provides UI functionality for discovering and selecting Skaffold configuration files in a project directory. The package handles recursive file discovery, categorization, and user selection with search capabilities.

## Key Components

- **SkaffoldUI**: Main struct that handles Skaffold file discovery and user interaction
- **ServiceSelection**: Represents a selected Skaffold service with name, file path, and directory
- **SkaffoldCategory**: Groups Skaffold files by type (OpenFrame Services, Integrated Tools, Client Apps, etc.)
- **DiscoverAndSelectService()**: Primary method that finds all `skaffold.yaml` files and prompts user selection
- **findSkaffoldYamlFiles()**: Recursively searches for Skaffold configuration files
- **categorizeSkaffoldFiles()**: Organizes discovered files into logical categories
- **extractServiceName()**: Derives clean service names from file paths

## Usage Example

```go
package main

import (
    "log"
    "github.com/yourproject/ui"
)

func main() {
    // Create new SkaffoldUI instance with verbose logging
    skaffoldUI := ui.NewSkaffoldUI(true)
    
    // Discover and select a service
    selection, err := skaffoldUI.DiscoverAndSelectService()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the selected service
    fmt.Printf("Selected service: %s\n", selection.ServiceName)
    fmt.Printf("Config file: %s\n", selection.FilePath)
    fmt.Printf("Directory: %s\n", selection.Directory)
}
```