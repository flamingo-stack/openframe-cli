<!-- source-hash: 9586e851899a1db8b85cd539f9bc0e6f -->
Test suite for table rendering and status formatting functions in the UI package. Validates color mapping, table display functions, and success message boxes.

## Key Components

- **`TestGetStatusColor`** - Tests status-to-color mapping for various cluster states (running, stopped, error, etc.)
- **`TestRenderTableWithFallback`** - Tests table rendering with and without headers using fallback display
- **`TestRenderKeyValueTable`** - Tests key-value pair table formatting 
- **`TestRenderNodeTable`** - Tests node-specific table rendering for cluster node information
- **`TestShowSuccessBox`** - Tests success message box display functionality

## Usage Example

```go
// Run tests
go test ./ui -v

// Test specific function
go test ./ui -run TestGetStatusColor

// The tests validate that UI functions handle various inputs correctly:
// - Status colors map properly (running->green, error->red, etc.)
// - Table rendering doesn't panic with valid pterm.TableData
// - Success boxes display without errors
```

These tests ensure the UI components gracefully handle different data formats and don't crash during table rendering or status display operations. All tests focus on preventing panics rather than asserting specific output formatting.