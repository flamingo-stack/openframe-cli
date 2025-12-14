<!-- source-hash: 284ee50153ba4104e9ffff4cb2220e35 -->
Manages prerequisite validation for OpenFrame CLI by checking required system dependencies and providing installation guidance. It validates the presence of Git, Helm, system memory, and certificates before allowing chart operations.

## Key Components

- **PrerequisiteChecker**: Main checker that validates multiple system requirements
- **Requirement**: Struct defining individual prerequisites with validation and help functions
- **NewPrerequisiteChecker()**: Factory function that initializes checker with built-in requirements (Git, Helm, Memory, Certificates)
- **CheckAll()**: Validates all prerequisites and returns status with missing tools
- **GetInstallInstructions()**: Provides installation help for missing dependencies
- **CheckPrerequisites()**: Convenience function for quick validation

## Usage Example

```go
// Basic prerequisite checking
checker := NewPrerequisiteChecker()
allPresent, missing := checker.CheckAll()

if !allPresent {
    fmt.Printf("Missing tools: %v\n", missing)
    
    // Get installation instructions
    instructions := checker.GetInstallInstructions(missing)
    for _, help := range instructions {
        fmt.Println(help)
    }
}

// Quick check using convenience function
if err := CheckPrerequisites(); err != nil {
    log.Fatal("Prerequisites not met:", err)
}
```