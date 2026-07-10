package ui

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	clusterUI "github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// OperationsUI provides user-friendly interfaces for chart operations
type OperationsUI struct {
	clusterSelector *clusterUI.Selector
}

// NewOperationsUI creates a new chart operations UI service
func NewOperationsUI() *OperationsUI {
	return &OperationsUI{
		clusterSelector: clusterUI.NewSelector("chart installation"),
	}
}

// SelectClusterForInstall handles cluster selection for chart installation
func (ui *OperationsUI) SelectClusterForInstall(clusters []models.ClusterInfo, args []string) (string, error) {
	return ui.clusterSelector.SelectCluster(clusters, args)
}

// ShowNoClusterMessage displays a friendly message when no clusters are available
func (ui *OperationsUI) ShowNoClusterMessage() {
	pterm.Error.Println("No clusters found. Create a cluster first with: openframe cluster create")
}

// ConfirmInstallationOnCluster asks for user confirmation before starting chart installation
func (ui *OperationsUI) ConfirmInstallationOnCluster(clusterName string) (bool, error) {
	fmt.Println() // Add blank line for better spacing
	message := fmt.Sprintf("Are you sure you want to install OpenFrame chart on '%s'? It could take up to 30 minutes", clusterName)
	return sharedUI.ConfirmActionInteractive(message, false)
}
