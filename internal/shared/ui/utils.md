<!-- source-hash: d3aca3edce11b7b08cff2599a0918a78 -->
Utility functions for formatting time-based data in user interfaces. Provides human-readable time duration formatting similar to common CLI tools.

## Key Components

- **FormatAge(createdAt time.Time) string** - Converts a timestamp to a human-readable age string (e.g., "2d", "3h", "45m", "30s")

## Usage Example

```go
import (
    "time"
    "yourproject/ui"
)

// Format recent timestamps
now := time.Now()
oneHourAgo := now.Add(-1 * time.Hour)
twoDaysAgo := now.Add(-48 * time.Hour)

fmt.Println(ui.FormatAge(oneHourAgo))  // "1h"
fmt.Println(ui.FormatAge(twoDaysAgo))  // "2d"

// Handle zero timestamps
var zeroTime time.Time
fmt.Println(ui.FormatAge(zeroTime))    // "unknown"

// Display in table or list format
for _, item := range items {
    fmt.Printf("%-20s %s\n", item.Name, ui.FormatAge(item.CreatedAt))
}
```

The function returns the largest applicable unit (days > hours > minutes > seconds) for concise display, making it ideal for status tables, logs, and resource listings.