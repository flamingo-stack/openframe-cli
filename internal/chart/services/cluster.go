package services

import (
	"fmt"
	"strings"

	chartUI "github.com/flamingo-stack/openframe-cli/internal/chart/ui"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/pterm/pterm"
)

// ClusterSelector handles cluster selection logic for chart operations
type ClusterSelector struct {
	clusterService types.ClusterLister
	operationsUI   *chartUI.OperationsUI
}

// NewClusterSelector creates a new cluster selector
func NewClusterSelector(clusterService types.ClusterLister, operationsUI *chartUI.OperationsUI) *ClusterSelector {
	return &ClusterSelector{
		clusterService: clusterService,
		operationsUI:   operationsUI,
	}
}

// SelectCluster manages the cluster selection process. When nonInteractive is
// set it must never fall through to an interactive picker (that would hang CI):
// without a cluster-name argument it fails fast with an actionable error.
func (c *ClusterSelector) SelectCluster(args []string, nonInteractive, verbose bool) (string, error) {
	clusters, err := c.clusterService.ListClusters()
	if err != nil {
		if verbose {
			pterm.Error.Printf("Failed to list clusters: %v\n", err)
		}
		c.operationsUI.ShowNoClusterMessage()
		return "", nil
	}

	if len(clusters) == 0 {
		if verbose {
			pterm.Info.Printf("Found 0 clusters\n")
		}
		c.operationsUI.ShowNoClusterMessage()
		return "", nil
	}

	if verbose {
		pterm.Info.Printf("Found %d clusters\n", len(clusters))
		for _, cluster := range clusters {
			pterm.Info.Printf("  - %s (%s)\n", cluster.Name, cluster.Status)
		}
	}

	// Non-interactive with no cluster name would otherwise reach the interactive
	// picker and block forever in CI. Fail fast, listing the available clusters.
	if nonInteractive && (len(args) == 0 || strings.TrimSpace(args[0]) == "") {
		names := make([]string, len(clusters))
		for i, cl := range clusters {
			names[i] = cl.Name
		}
		return "", fmt.Errorf("--non-interactive requires a cluster name argument; available clusters: %s", strings.Join(names, ", "))
	}

	return c.operationsUI.SelectClusterForInstall(clusters, args)
}
