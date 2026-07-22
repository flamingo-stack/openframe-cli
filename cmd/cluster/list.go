package cluster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/discovery"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func getListCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all Kubernetes clusters",
		Long: `List all Kubernetes clusters managed by OpenFrame CLI.

Displays cluster information including name, type, status, and node count
from all registered providers in a formatted table.

With --all, additionally discovers GKE clusters that exist in the GCP
projects of your gcloud configurations but were created outside openframe.
Discovered clusters are read-only: openframe never modifies or deletes them.

Examples:
  openframe cluster list
  openframe cluster list --all
  openframe cluster list --verbose
  openframe cluster list --quiet`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			if err := utils.ValidateGlobalFlags(); err != nil {
				return err
			}
			globalFlags := utils.GetGlobalFlags()
			if globalFlags != nil && globalFlags.List != nil {
				return models.ValidateListFlags(globalFlags.List)
			}
			return nil
		},
		RunE: utils.WrapCommandWithCommonSetup(runListClusters),
	}

	// Add list-specific flags
	globalFlags := utils.GetGlobalFlags()
	if globalFlags != nil && globalFlags.List != nil {
		models.AddListFlags(listCmd, globalFlags.List)
	}
	listCmd.Flags().StringP("output", "o", "text", "Output format: text, json, or yaml")

	return listCmd
}

func runListClusters(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	globalFlags := utils.GetGlobalFlags()

	// Get all clusters
	clusters, err := service.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	var notices []string
	if globalFlags.List.All {
		external, discoveryNotices := discoverExternalClusters(cmd.Context(), clusters)
		clusters = append(clusters, external...)
		notices = discoveryNotices
	}

	switch out, _ := cmd.Flags().GetString("output"); out {
	case "json":
		return printClustersJSON(clusters)
	case "yaml":
		return printClustersYAML(clusters)
	case "", "text":
		if err := service.DisplayClusterList(clusters, globalFlags.List.Quiet, globalFlags.Global.Verbose); err != nil {
			return err
		}
		for _, notice := range notices {
			pterm.Info.Println(notice)
		}
		return nil
	default:
		return fmt.Errorf("invalid --output %q (want \"text\", \"json\", or \"yaml\")", out)
	}
}

// discoverExternalClusters runs GKE discovery, dropping entries that are
// already managed (same name+project as a registry cluster). Auth problems
// degrade to notices, never errors: a logged-out gcloud must not break list.
// EKS discovery is not implemented yet — a notice says so.
func discoverExternalClusters(ctx context.Context, managed []models.ClusterInfo) ([]models.ClusterInfo, []string) {
	notices := []string{"AWS EKS discovery is coming soon — external EKS clusters are not shown yet"}

	exec := utils.CommandExecutor()
	d := discovery.NewGKEDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return nil, append(notices, "GKE: gcloud is not installed — install it to discover external clusters")
	case discovery.NotAuthenticated:
		// One unambiguous flow: offer the login right here (interactive only —
		// non-interactive sessions get the same message as before).
		if err := discovery.NewAuthFlow(exec).Ensure(ctx, false); err != nil {
			return nil, append(notices, "GKE: "+err.Error())
		}
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return nil, append(notices, fmt.Sprintf("GKE discovery failed: %v", err))
	}
	for _, w := range result.Warnings {
		notices = append(notices, "GKE discovery skipped "+w)
	}

	isManaged := func(c models.ClusterInfo) bool {
		for _, m := range managed {
			// Project-aware: a local k3d cluster (empty Project) must not
			// suppress an external GKE cluster that merely shares its name.
			if m.Name == c.Name && m.Project == c.Project {
				return true
			}
		}
		return false
	}
	var external []models.ClusterInfo
	for _, c := range result.Clusters {
		if !isManaged(c) {
			external = append(external, c)
		}
	}
	return external, notices
}

// clusterJSON is the machine-readable shape of a cluster.
type clusterJSON struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	NodeCount  int    `json:"nodeCount"`
	K8sVersion string `json:"k8sVersion,omitempty"`
}

func clustersToJSON(clusters []models.ClusterInfo) []clusterJSON {
	out := make([]clusterJSON, 0, len(clusters))
	for _, c := range clusters {
		out = append(out, clusterJSON{
			Name:       c.Name,
			Type:       string(c.Type),
			Status:     c.Status,
			NodeCount:  c.NodeCount,
			K8sVersion: c.K8sVersion,
		})
	}
	return out
}

func printClustersJSON(clusters []models.ClusterInfo) error {
	b, err := json.MarshalIndent(clustersToJSON(clusters), "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// printClustersYAML writes the cluster list as YAML. sigs.k8s.io/yaml reuses the
// same `json:` struct tags, so the field names match the JSON output.
func printClustersYAML(clusters []models.ClusterInfo) error {
	b, err := yaml.Marshal(clustersToJSON(clusters))
	if err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	fmt.Print(string(b)) // yaml.Marshal already terminates with a newline
	return nil
}
