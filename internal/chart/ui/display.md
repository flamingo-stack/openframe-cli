<!-- source-hash: 377fc7e65234def6684c10076cb6cec5 -->
Provides UI display functionality for chart installation operations with progress indicators, status messages, and formatted output using the pterm library.

## Key Components

- **DisplayService**: Main service struct for handling chart-related UI operations
- **NewDisplayService()**: Constructor function that returns a new DisplayService instance
- **ShowInstallProgress()**: Displays installation progress with chart-type-specific icons
- **ShowInstallSuccess()**: Handles successful installation display (placeholder implementation)
- **ShowInstallError()**: Shows formatted error messages for failed installations
- **ShowSkippedInstallation()**: Displays when component installation is skipped
- **ShowPreInstallCheck()**: Shows pre-installation validation messages
- **ShowDryRunResults()**: Outputs dry-run results to specified writer
- **getChartDisplayName()**: Returns user-friendly names for chart types

## Usage Example

```go
import (
    "github.com/flamingo-stack/openframe-cli/internal/chart/models"
    "github.com/flamingo-stack/openframe-cli/internal/ui"
)

// Create display service
display := ui.NewDisplayService()

// Show installation progress
display.ShowInstallProgress(models.ChartTypeArgoCD, "Installing ArgoCD chart...")

// Handle installation error
if err != nil {
    display.ShowInstallError(models.ChartTypeArgoCD, err)
}

// Show skipped installation
display.ShowSkippedInstallation("ArgoCD", "already exists")

// Display dry-run results
results := []string{"Chart validated", "Dependencies checked"}
display.ShowDryRunResults(os.Stdout, results)
```