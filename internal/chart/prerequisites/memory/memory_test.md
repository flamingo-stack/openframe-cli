<!-- source-hash: d0fe6d43828955090e40f8e1538904b2 -->
Test suite for the memory checker component that validates system memory detection and requirement checking functionality.

## Key Components

- **TestNewMemoryChecker**: Validates memory checker instantiation
- **TestMemoryChecker_GetMemoryInfo**: Tests current, recommended, and sufficient memory detection
- **TestMemoryChecker_GetInstallHelp**: Verifies help message content and formatting
- **TestMemoryChecker_Install**: Confirms installation attempt returns expected error
- **TestGetTotalMemoryMB**: Tests platform-specific memory detection methods
- **TestHasSufficientMemory**: Validates memory sufficiency logic
- **containsSubstring**: Helper function for string matching in tests

## Usage Example

```go
// Run specific test
go test -run TestMemoryChecker_GetMemoryInfo

// Run all memory tests
go test ./memory

// Test with verbose output
go test -v ./memory

// Example assertion pattern used in tests
checker := NewMemoryChecker()
current, recommended, sufficient := checker.GetMemoryInfo()
if current <= 0 {
    t.Error("Current memory should be greater than 0")
}
```

The tests cover cross-platform memory detection, error handling for installation attempts, and validate that memory information includes proper units and recommendations.