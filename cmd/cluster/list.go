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

With --all, additionally discovers cloud clusters created outside openframe:
GKE clusters in the GCP projects of your gcloud configurations, and EKS
clusters in each AWS profile's default region. Discovered clusters are
read-only: openframe never modifies or deletes them.

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

// discoverExternalClusters runs GKE and EKS discovery, dropping entries that
// are already managed (same name+type+project as a registry cluster). Auth
// problems degrade to notices, never errors: a logged-out gcloud or an expired
// AWS session must not break list.
func discoverExternalClusters(ctx context.Context, managed []models.ClusterInfo) ([]models.ClusterInfo, []string) {
	var notices []string
	var discovered []models.ClusterInfo

	gkeClusters, gkeNotices := discoverGKE(ctx)
	discovered = append(discovered, gkeClusters...)
	notices = append(notices, gkeNotices...)

	eksClusters, eksNotices := discoverEKS(ctx)
	discovered = append(discovered, eksClusters...)
	notices = append(notices, eksNotices...)

	isManaged := func(c models.ClusterInfo) bool {
		for _, m := range managed {
			// Type- and project-aware: a local k3d cluster must not suppress an
			// external cloud cluster that merely shares its name, and a GKE
			// cluster's name must not shadow a same-named EKS one.
			if m.Name == c.Name && m.Type == c.Type && m.Project == c.Project {
				return true
			}
		}
		return false
	}
	var external []models.ClusterInfo
	for _, c := range discovered {
		if !isManaged(c) {
			external = append(external, c)
		}
	}
	return external, notices
}

// discoverGKE is the GCP half of --all.
func discoverGKE(ctx context.Context) ([]models.ClusterInfo, []string) {
	exec := utils.CommandExecutor()
	d := discovery.NewGKEDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return nil, []string{"GKE: gcloud is not installed — install it to discover external clusters"}
	case discovery.NotAuthenticated:
		// One unambiguous flow: offer the login right here (interactive only —
		// non-interactive sessions get the same message as before).
		if err := discovery.NewAuthFlow(exec).Ensure(ctx, false); err != nil {
			return nil, []string{"GKE: " + err.Error()}
		}
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return nil, []string{fmt.Sprintf("GKE discovery failed: %v", err)}
	}
	var notices []string
	for _, w := range result.Warnings {
		notices = append(notices, "GKE discovery skipped "+w)
	}
	return result.Clusters, notices
}

// discoverEKS is the AWS half of --all. There is no interactive AWS login to
// offer (credentials come from profiles/SSO/env, not a browser flow the CLI
// can drive), so an unusable identity degrades straight to a notice.
func discoverEKS(ctx context.Context) ([]models.ClusterInfo, []string) {
	exec := utils.CommandExecutor()
	d := discovery.NewEKSDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return nil, []string{"EKS: the AWS CLI is not installed — install it to discover external clusters"}
	case discovery.NotAuthenticated:
		return nil, []string{"EKS: AWS credentials are not usable — run 'aws configure' or 'aws sso login' to discover external clusters"}
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return nil, []string{fmt.Sprintf("EKS discovery failed: %v", err)}
	}
	var notices []string
	for _, w := range result.Warnings {
		notices = append(notices, "EKS discovery skipped "+w)
	}
	return result.Clusters, notices
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
