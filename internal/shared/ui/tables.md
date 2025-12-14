<!-- source-hash: 3ba816996027a02fb60ad5db1dc01ee3 -->
Provides utility functions for rendering tables and formatted output in CLI applications using the pterm library, with fallback mechanisms for environments where styled output may not be supported.

## Key Components

- **RenderTableWithFallback()** - Renders a 5-column table with styled output, falls back to plain text formatting
- **RenderKeyValueTable()** - Displays data in key-value pairs with header handling
- **RenderNodeTable()** - Specialized table renderer for node information with fixed column layout
- **GetStatusColor()** - Returns appropriate color functions based on status values (running/ready = green, stopped/pending = yellow, error/failed = red)
- **ShowSuccessBox()** - Displays formatted success messages in a styled box

## Usage Example

```go
// Render a cluster status table
data := pterm.TableData{
    {"Name", "Status", "Type", "Nodes", "Age"},
    {"my-cluster", "Running", "kind", "3", "2h"},
}
ui.RenderTableWithFallback(data, true)

// Display key-value information
info := pterm.TableData{
    {"Key", "Value"},
    {"Version", "1.0.0"},
    {"Status", "Active"},
}
ui.RenderKeyValueTable(info)

// Color status text
colorFunc := ui.GetStatusColor("running")
coloredStatus := colorFunc("Running") // Returns green text

// Show success message
ui.ShowSuccessBox("Operation Complete", "Cluster created successfully!")
```