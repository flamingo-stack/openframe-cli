<!-- source-hash: bac5d977361f7636fa8e3e86986e72bd -->
Unit test suite for the prerequisites installer package, validating installer creation, tool installation behavior, and command execution.

## Key Components

- **TestNewInstaller** - Verifies proper initialization of installer instance and its checker component
- **TestInstallTool** - Tests tool installation logic including error handling for memory constraints and unknown tools
- **TestRunCommand** - Validates command execution functionality using basic system commands
- **containsSubstring** - Helper function for substring matching in test assertions

## Usage Example

```go
// Run all tests
go test -v

// Test installer creation
func TestNewInstaller(t *testing.T) {
    installer := NewInstaller()
    if installer == nil {
        t.Error("Expected installer to be created")
    }
}

// Test error handling for unsupported operations
func TestInstallTool(t *testing.T) {
    installer := NewInstaller()
    
    // Memory tool should fail with specific error
    err := installer.installTool("memory")
    expectedError := "memory cannot be automatically increased"
    if !containsSubstring(err.Error(), expectedError) {
        t.Errorf("Expected error containing '%s'", expectedError)
    }
}
```