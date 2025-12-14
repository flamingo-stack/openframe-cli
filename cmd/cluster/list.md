<!-- source-hash: 7ba72a9cb27fb4b17e6592adb604e41b -->
Implements the `list` subcommand for the OpenFrame CLI cluster management functionality. This command retrieves and displays information about all Kubernetes clusters managed by the CLI.

## Key Components

- **`getListCmd()`** - Creates and configures the cobra command for listing clusters with flag initialization and validation
- **`runListClusters()`** - Executes the cluster listing logic by calling the service layer and formatting output
- **Command validation** - Pre-run validation of global and list-specific flags through the utils package
- **Service integration** - Uses the command service to retrieve cluster data and handle display formatting

## Usage Example

```go
// Create the list command
listCmd := getListCmd()

// The command supports various output formats
// Basic usage: openframe cluster list
// Verbose: openframe cluster list --verbose  
// Quiet: openframe cluster list --quiet

// Command execution flow:
// 1. Validates flags in PreRunE
// 2. Calls runListClusters via WrapCommandWithCommonSetup
// 3. Retrieves clusters using service.ListClusters()
// 4. Displays formatted output via service.DisplayClusterList()
```

The command integrates with the broader cluster management system through the utils package for common setup and the service layer for data operations, providing a clean separation of concerns between CLI interface and business logic.