<!-- source-hash: a7b831da12b6bdad16de69f28b4b5db1 -->
Test suite for the ui package's time formatting utilities. Contains comprehensive tests for the `FormatAge` function to ensure proper time duration formatting across different time units.

## Key Components

- **TestFormatAge**: Main test function validating age formatting for seconds, minutes, hours, and days
- **TestFormatAgeEdgeCases**: Additional tests covering edge cases like recent times and future dates
- **Test cases**: Covers zero time, various duration units (s, m, h, d), and boundary conditions

## Usage Example

```go
// Run the tests
go test -v ./ui

// Run specific test
go test -run TestFormatAge ./ui

// Run with coverage
go test -cover ./ui

// The tests validate that FormatAge works correctly:
// FormatAge(time.Time{}) -> "unknown"
// FormatAge(30 seconds ago) -> "30s" 
// FormatAge(5 minutes ago) -> "5m"
// FormatAge(2 hours ago) -> "2h"
// FormatAge(3 days ago) -> "3d"
```

The test suite uses table-driven tests for the main scenarios and separate functions for edge cases, with time-tolerance handling to account for test execution delays.