<!-- source-hash: 8395cc490bf8b055e26e78bd05e1f510 -->
A cross-platform memory checker that validates system memory requirements and provides memory information across Windows, macOS, and Linux systems.

## Key Components

- **MemoryChecker**: Main struct implementing memory validation interface
- **NewMemoryChecker()**: Constructor function returning a new MemoryChecker instance
- **IsInstalled()**: Checks if system has sufficient memory (≥15GB)
- **HasSufficientMemory()**: Validates memory against recommended threshold
- **GetMemoryInfo()**: Returns current memory, recommended memory, and sufficiency status
- **GetInstallHelp()**: Provides memory upgrade guidance
- **RecommendedMemoryMB**: Constant defining 15GB memory requirement

## Usage Example

```go
checker := NewMemoryChecker()

// Check if memory requirements are met
if checker.IsInstalled() {
    fmt.Println("System has sufficient memory")
} else {
    fmt.Println(checker.GetInstallHelp())
}

// Get detailed memory information
current, recommended, sufficient := checker.GetMemoryInfo()
fmt.Printf("Current: %d MB, Recommended: %d MB, Sufficient: %t\n", 
    current, recommended, sufficient)

// Attempt installation (will return error)
err := checker.Install()
if err != nil {
    fmt.Printf("Cannot install memory: %v\n", err)
}
```

The checker uses platform-specific commands (sysctl for macOS, /proc/meminfo for Linux, PowerShell for Windows) to retrieve accurate memory information.