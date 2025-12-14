<!-- source-hash: ef576e7b3ee03365926f8ea4658eea0b -->
This file provides a comprehensive Git prerequisite checker for validating Git installation and version requirements in Go applications.

## Key Components

- **GitChecker**: Main struct that encapsulates all Git validation functionality
- **NewGitChecker()**: Constructor function for creating new GitChecker instances
- **IsInstalled()**: Checks if Git is available in the system PATH
- **GetVersion()**: Retrieves the installed Git version string
- **GetInstallInstructions()**: Returns platform-specific installation guidance
- **Validate()**: Performs comprehensive validation including minimum version checks (Git 2.0+)

## Usage Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    checker := NewGitChecker()
    
    // Basic installation check
    if !checker.IsInstalled() {
        fmt.Println(checker.GetInstallInstructions())
        return
    }
    
    // Get version information
    version, err := checker.GetVersion()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Git version: %s\n", version)
    
    // Comprehensive validation
    if err := checker.Validate(); err != nil {
        log.Fatalf("Git validation failed: %v", err)
    }
    
    fmt.Println("Git is properly installed and ready to use")
}
```

This checker is particularly useful for applications that depend on Git for operations like cloning chart repositories or other version control tasks.