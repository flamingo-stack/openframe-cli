<!-- source-hash: de21251f40fd2f62eee811dfe517d487 -->
This file contains unit tests for the `CloneResult` struct, verifying its initialization behavior and field assignments.

## Key Components

- `TestCloneResult_DefaultValues` - Tests that a new CloneResult instance has empty default values
- `TestCloneResult_WithValues` - Tests that CloneResult fields can be properly set and retrieved

## Usage Example

```go
// Run tests using go test
go test -v ./git

// The tests validate CloneResult behavior:
func TestCloneResult_DefaultValues(t *testing.T) {
    result := &CloneResult{}
    
    assert.Empty(t, result.TempDir)
    assert.Empty(t, result.ChartPath)
}

func TestCloneResult_WithValues(t *testing.T) {
    result := &CloneResult{
        TempDir:   "/tmp/clone-12345",
        ChartPath: "/tmp/clone-12345/manifests/app-of-apps",
    }
    
    assert.Equal(t, "/tmp/clone-12345", result.TempDir)
    assert.Equal(t, "/tmp/clone-12345/manifests/app-of-apps", result.ChartPath)
}
```