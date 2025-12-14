<!-- source-hash: 1e81d00a3a140eb6cc4dac2171e67fa5 -->
Provides a prerequisite checking system for CLI development tools. The checker validates and provides installation guidance for required tools like Telepresence, jq, and Skaffold.

## Key Components

- **PrerequisiteChecker**: Main struct that manages multiple tool requirements
- **Requirement**: Defines a tool requirement with name, command, installation check, and help functions
- **NewPrerequisiteChecker()**: Factory function that initializes checker with predefined requirements
- **CheckAll()**: Validates all prerequisites and returns status with missing tools list
- **GetInstallInstructions()**: Retrieves installation help for specific missing tools
- **CheckPrerequisites()**: Convenience function for checking and installing prerequisites

## Usage Example

```go
// Check all prerequisites
checker := NewPrerequisiteChecker()
allInstalled, missing := checker.CheckAll()

if !allInstalled {
    fmt.Printf("Missing tools: %v\n", missing)
    
    // Get installation instructions for missing tools
    instructions := checker.GetInstallInstructions(missing)
    for _, instruction := range instructions {
        fmt.Println(instruction)
    }
}

// Simple prerequisite check and install
err := CheckPrerequisites()
if err != nil {
    log.Fatal("Prerequisites check failed:", err)
}
```