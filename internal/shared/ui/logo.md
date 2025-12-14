<!-- source-hash: 3178efc67608bb4c523f75c1552b5c76 -->
Provides ASCII logo display functionality for the OpenFrame Platform Bootstrapper with support for both fancy terminal output and plain text fallback modes.

## Key Components

- **ShowLogo()** - Displays the OpenFrame ASCII logo using default settings
- **ShowLogoConditional(suppress bool)** - Displays logo with optional suppression flag
- **ShowLogoWithContext(ctx context.Context)** - Context-aware logo display that respects suppression settings
- **WithSuppressedLogo(ctx context.Context)** - Returns a context with logo suppression enabled
- **TestMode** - Global variable to suppress logo output during testing
- **logoArt** - Unicode art array containing the OpenFrame logo design

The package automatically detects terminal capabilities and uses either fancy colored output (via pterm) or plain text with ASCII borders.

## Usage Example

```go
package main

import (
    "context"
    "your-project/ui"
)

func main() {
    // Simple logo display
    ui.ShowLogo()
    
    // Conditional display
    ui.ShowLogoConditional(false) // shows logo
    ui.ShowLogoConditional(true)  // suppresses logo
    
    // Context-based suppression
    ctx := context.Background()
    ui.ShowLogoWithContext(ctx) // shows logo
    
    suppressedCtx := ui.WithSuppressedLogo(ctx)
    ui.ShowLogoWithContext(suppressedCtx) // suppresses logo
    
    // Disable during testing
    ui.TestMode = true
    ui.ShowLogo() // no output
}
```