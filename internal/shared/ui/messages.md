<!-- source-hash: f55b65771cdbf42714cf56d9cbe2f584 -->
Provides user-friendly message display functions for CLI operations, including resource availability warnings, operation status updates, and error handling with troubleshooting suggestions.

## Key Components

- **ShowNoResourcesMessage()** - Displays a formatted warning when no resources are found, with helpful next steps
- **ShowOperationStart()** - Shows an info message when beginning an operation, supports custom messages
- **ShowOperationSuccess()** - Displays success confirmation after operation completion
- **ShowOperationError()** - Shows detailed error information with formatted troubleshooting tips
- **TroubleshootingTip** - Struct containing description and command for help suggestions

## Usage Example

```go
import "path/to/ui"

// Show no resources found
ui.ShowNoResourcesMessage(
    "deployments", 
    "start", 
    "kubectl create deployment", 
    "kubectl get deployments",
)

// Show operation progress
ui.ShowOperationStart("deploy", "my-app", nil)

// Show success with custom message
customMessages := map[string]string{
    "deploy": "Successfully deployed application!",
}
ui.ShowOperationSuccess("deploy", "my-app", customMessages)

// Show error with troubleshooting
tips := []ui.TroubleshootingTip{
    {Description: "Check status", Command: "kubectl get pods"},
    {Description: "View logs", Command: "kubectl logs my-app"},
}
ui.ShowOperationError("deploy", "my-app", err, tips)
```