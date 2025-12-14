<!-- source-hash: 676141c23c4319e41333f962b6e1d8e8 -->
This file provides a user interface layer for chart operations within the OpenFrame CLI, handling user interactions for installing charts on Kubernetes clusters.

## Key Components

- **OperationsUI**: Main struct that orchestrates chart operation user interactions
- **NewOperationsUI()**: Constructor function that initializes the operations UI service
- **SelectClusterForInstall()**: Handles cluster selection for chart installations
- **ConfirmInstallation()**: Prompts user for installation confirmation with timing warnings
- **ShowInstallation*()**: Series of methods for displaying installation status messages
- **PromptForGitHubCredentials()**: Handles GitHub authentication for repository access
- **ShowCloneProgress()/ShowCloneComplete()**: Display repository cloning status

## Usage Example

```go
// Initialize operations UI
opsUI := NewOperationsUI()

// Select a cluster for installation
clusterName, err := opsUI.SelectClusterForInstall(clusters, args)
if err != nil {
    return err
}

// Confirm installation with user
confirmed, err := opsUI.ConfirmInstallation(clusterName)
if err != nil || !confirmed {
    opsUI.ShowOperationCancelled("installation")
    return nil
}

// Show installation progress
opsUI.ShowInstallationStart(clusterName)

// On success
opsUI.ShowInstallationComplete()

// On error
opsUI.ShowInstallationError(err)
```