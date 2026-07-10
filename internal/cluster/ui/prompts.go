package ui

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// Use domain types for consistency - no duplicate definitions needed
type ClusterType = models.ClusterType
type ClusterInfo = models.ClusterInfo

// Re-export domain constants for UI convenience
const (
	ClusterTypeK3d = models.ClusterTypeK3d
	ClusterTypeGKE = models.ClusterTypeGKE
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
