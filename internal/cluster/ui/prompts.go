package ui

import (
	"fmt"
	"strconv"
	"strings"

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

	selectedIndex, _, err := selectFromList(prompt, clusterPickerRows(clusters))
	if err != nil {
		return "", err
	}

	return clusters[selectedIndex].Name, nil
}

// clusterPickerRows renders one aligned multi-column row per cluster —
// name, type, where it lives, status — so picking between a local k3d and
// two same-named cloud clusters is not a guess. Row order mirrors clusters.
func clusterPickerRows(clusters []ClusterInfo) []string {
	nameW, typeW := 0, 0
	for _, cl := range clusters {
		nameW = max(nameW, len(cl.Name))
		typeW = max(typeW, len(string(cl.Type)))
	}
	g := sharedUI.Glyphs()
	rows := make([]string, 0, len(clusters))
	for _, cl := range clusters {
		where := cl.Region
		if where == "" {
			where = "local"
		}
		if cl.NodeCount > 0 {
			where = fmt.Sprintf("%s %s %d node(s)", where, g.Bullet, cl.NodeCount)
		}
		rows = append(rows, fmt.Sprintf("%-*s  %-*s %s %s %s %s",
			nameW, cl.Name, typeW, strings.ToUpper(string(cl.Type)), g.Bullet, where, g.Bullet, cl.Status))
	}
	return rows
}

// selectFromList shows a selection prompt for a list of items
func selectFromList(prompt string, items []string) (int, string, error) {
	// Use common UI function
	return sharedUI.SelectFromList(prompt, items)
}

// CostHint is the running-cost warning shown for cloud cluster types. It
// deliberately carries NO figures — real numbers come only from the optional
// infracost estimate in the dry-run preview; otherwise the user gets the
// provider's pricing page, never a stale hardcoded price.
func CostHint(clusterType models.ClusterType) string {
	switch clusterType {
	case models.ClusterTypeEKS:
		return "This creates billed AWS resources (managed control plane, EC2 nodes, networking) — pricing: https://aws.amazon.com/eks/pricing/"
	case models.ClusterTypeGKE:
		return "This creates billed GCP resources (managed control plane, VM nodes, networking) — pricing: https://cloud.google.com/kubernetes-engine/pricing"
	default:
		return "Cloud clusters create resources that incur costs"
	}
}

// NodesLine renders a config's node count honestly: a regional (--ha) cluster
// provisions its count PER ZONE, so "3" would silently mean 9 nodes and ~3×
// the expected bill — a verification pass hit exactly that (the summary said
// 3, GCP came up with 9). Zonal clusters keep the plain number.
func NodesLine(config models.ClusterConfig) string {
	if config.Cloud == nil || !config.Cloud.HA {
		return strconv.Itoa(config.NodeCount)
	}
	return fmt.Sprintf("%d per zone × %d zones = %d total (regional)",
		config.NodeCount, models.GKERegionalZones, config.NodeCount*models.GKERegionalZones)
}

// SpotHint nudges test-cluster users toward spot capacity, next to the cost
// warning — the flag already exists but nothing advertised it. Empty when spot
// is already on (nothing to suggest) or for non-cloud types. Like CostHint it
// carries no exact price: the discount range is the provider's own
// (preemptible/spot pricing), not a figure this CLI computes.
func SpotHint(config models.ClusterConfig) string {
	if config.Cloud == nil || config.Cloud.Spot {
		return ""
	}
	switch config.Type {
	case models.ClusterTypeEKS, models.ClusterTypeGKE:
		return "Tip: for test clusters, --spot runs nodes on spot capacity (typically 60-90% off the node cost)"
	}
	return ""
}

// ConfirmTypedClusterName requires the user to re-type the cluster name
// before a cloud destroy — a stronger gate than yes/no, because the action
// deletes billed infrastructure irreversibly.
func ConfirmTypedClusterName(name string) (bool, error) {
	pterm.Warning.Printf("Deleting a cloud cluster destroys all its cloud resources.\n")
	entered, err := sharedUI.PromptInput(fmt.Sprintf("Type the cluster name (%s) to confirm", name), "", nil)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(entered) == name, nil
}
