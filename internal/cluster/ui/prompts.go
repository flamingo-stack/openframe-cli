package ui

import (
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
)

// Use domain types for consistency - no duplicate definitions needed
type ClusterType = models.ClusterType
type ClusterInfo = models.ClusterInfo

// Re-export domain constants for UI convenience
const (
	ClusterTypeK3d = models.ClusterTypeK3d
)

// UI should not depend on business logic interfaces
// Business logic functions will be injected as simple parameters

// SelectClusterByName allows user to interactively select from available clusters by name
// Takes pre-fetched cluster list instead of manager to separate UI from business logic
func SelectClusterByName(clusters []ClusterInfo, prompt string) (string, error) {
	if len(clusters) == 0 {
		pterm.Warning.Println("No clusters found")
		return "", nil
	}

	clusterNames := make([]string, 0, len(clusters))
	for _, cl := range clusters {
		clusterNames = append(clusterNames, cl.Name)
	}

	if len(clusterNames) == 0 {
		pterm.Warning.Println("No clusters available")
		return "", nil
	}

	selectedIndex, _, err := selectFromList(prompt, clusterNames)
	if err != nil {
		return "", err
	}

	return clusterNames[selectedIndex], nil
}

// selectFromList shows a selection prompt for a list of items
func selectFromList(prompt string, items []string) (int, string, error) {
	// Use common UI function
	return sharedUI.SelectFromList(prompt, items)
}

// CostHint is the running-cost warning shown for cloud cluster types. The
// figures are deliberately rough baselines, not quotes.
func CostHint(clusterType models.ClusterType) string {
	switch clusterType {
	case models.ClusterTypeEKS:
		return "This creates billed AWS resources: EKS control plane (~$73/mo), EC2 nodes, and a NAT gateway (~$33/mo + traffic)"
	case models.ClusterTypeGKE:
		return "This creates billed GCP resources: GKE cluster management fee (~$73/mo), VM nodes, and networking"
	default:
		return "Cloud clusters create resources that incur costs"
	}
}

// ConfirmTypedClusterName requires the user to re-type the cluster name
// before a cloud destroy — a stronger gate than yes/no, because the action
// deletes billed infrastructure irreversibly.
func ConfirmTypedClusterName(name string) (bool, error) {
	pterm.Warning.Printf("Deleting a cloud cluster destroys all its cloud resources.\n")
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Type the cluster name (%s) to confirm", name),
	}
	entered, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(entered) == name, nil
}
