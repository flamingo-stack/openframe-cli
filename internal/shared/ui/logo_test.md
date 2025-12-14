<!-- source-hash: 04de0e53dc65a0fe7e682773a943e560 -->
Test file for the UI logo display functionality, providing comprehensive test coverage for logo rendering modes, text centering utilities, and environment-based configuration.

## Key Components

- **ShowLogo Tests**: Validates logo display behavior in test, plain, and fancy modes
- **centerText Tests**: Comprehensive testing of text centering algorithm with edge cases
- **Environment Tests**: Verifies logo behavior with different environment variable configurations  
- **Performance Tests**: Benchmarks for logo rendering and text centering operations
- **Helper Functions**: `captureOutput()` utility for testing stdout output

## Usage Example

```go
// Run all logo tests
go test -v ./ui -run TestShowLogo

// Run specific test with environment setup
func TestCustomLogo(t *testing.T) {
    os.Setenv("OPENFRAME_FANCY_LOGO", "false")
    defer os.Unsetenv("OPENFRAME_FANCY_LOGO")
    
    output := captureOutput(func() {
        ShowLogo()
    })
    
    assert.Contains(t, output, "OpenFrame Platform")
}

// Run benchmarks
go test -bench=BenchmarkShowPlainLogo -benchmem
```

The test suite covers all logo display modes, validates proper environment variable handling, and includes performance benchmarks. It uses the `captureOutput` helper to test stdout-based logo functions and ensures proper cleanup of test state.