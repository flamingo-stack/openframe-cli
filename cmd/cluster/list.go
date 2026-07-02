package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/spf13/cobra"
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

Examples:
  openframe cluster list
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
	listCmd.Flags().StringP("output", "o", "text", "Output format: text or json")

	return listCmd
}

func runListClusters(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()

	// Get all clusters
	clusters, err := service.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	switch out, _ := cmd.Flags().GetString("output"); out {
	case "json":
		return printClustersJSON(clusters)
	case "", "text":
		globalFlags := utils.GetGlobalFlags()
		return service.DisplayClusterList(clusters, globalFlags.List.Quiet, globalFlags.Global.Verbose)
	default:
		return fmt.Errorf("invalid --output %q (want \"text\" or \"json\")", out)
	}
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
