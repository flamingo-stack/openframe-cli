<!-- source-hash: 883dc6ed7b733c44fbe3818fa6822980 -->
This file provides a centralized system for checking and validating required development tools (Docker, kubectl, k3d) before running cluster operations.

## Key Components

- **PrerequisiteChecker**: Main struct that manages a collection of requirements and provides checking functionality
- **Requirement**: Struct defining a prerequisite with name, command, installation check, and help instructions
- **NewPrerequisiteChecker()**: Factory function that creates a checker with predefined requirements for Docker, kubectl, and k3d
- **CheckAll()**: Returns whether all prerequisites are met and lists missing tools
- **GetInstallInstructions()**: Provides installation help for specific missing tools
- **CheckPrerequisites()**: Convenience function for checking and installing prerequisites

## Usage Example

```go
// Check all prerequisites
checker := NewPrerequisiteChecker()
allPresent, missing := checker.CheckAll()

if !allPresent {
    fmt.Printf("Missing tools: %v\n", missing)
    
    // Get installation instructions
    instructions := checker.GetInstallInstructions(missing)
    for _, instruction := range instructions {
        fmt.Println(instruction)
    }
}

// Quick check using convenience function
if err := CheckPrerequisites(); err != nil {
    log.Fatal("Prerequisites not met:", err)
}
```