package cluster

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/spf13/cobra"
)

// GetClusterCmd returns the cluster command and its subcommands
func GetClusterCmd() *cobra.Command {
	// Initialize global flags
	utils.InitGlobalFlags()

	clusterCmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"k"},
		Short:   "Manage Kubernetes clusters",
		Long: `Cluster Management - Create, manage, and clean up Kubernetes clusters

This command group provides cluster lifecycle management functionality:
  • create - Create a new cluster with interactive configuration
  • delete - Remove a cluster and clean up resources
  • list - Show all managed clusters
  • status - Display detailed cluster information
  • use - Switch the kubectl context to a cluster
  • cleanup - Remove unused images and resources

Supports K3d clusters for local development and Google GKE / AWS EKS for cloud deployments.

Examples:
  openframe cluster create
  openframe cluster delete`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// This command group defines its own PersistentPreRunE, which shadows
			// the root's (cobra runs only the closest parent's hook), so apply
			// the global --silent AND --verbose contract here too — replicating
			// only the silent half left every cluster subcommand with zero
			// debug output under --verbose.
			ui.ApplyGlobalOutputFlags(cmd)
			// Machine output (json/yaml) is machine mode: no logo, no prerequisite
			// gate, so stdout stays clean for scripts.
			if out, _ := cmd.Flags().GetString("output"); out == "json" || out == "yaml" {
				return nil
			}
			// Show logo for subcommands, but not for the root cluster command
			if cmd.Use != "cluster" {
				ui.ShowLogoWithContext(cmd.Context())
				// One dim line naming the current kube-context — the cheapest
				// guard against pointing a destructive command at the wrong cluster.
				ui.ShowContextHeader()
			}
			// create runs its own type-aware gate after the cluster type is known
			// (a cloud cluster must not demand Docker/k3d); use only flips local
			// kubeconfig/gcloud state and needs no tools at all. status and list
			// are cross-provider read-only views: they must work against a cloud
			// cluster with Docker stopped, so they skip the k3d gate and degrade
			// gracefully instead (k3d enumeration is best-effort in the service).
			switch cmd.Name() {
			case "create", "use", "status", "list":
				return nil
			}
			return prerequisites.CheckPrerequisites()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show logo when no subcommand is provided
			ui.ShowLogoWithContext(cmd.Context())
			return cmd.Help()
		},
	}

	// Add subcommands - much simpler now
	clusterCmd.AddCommand(
		getCreateCmd(),
		getDeleteCmd(),
		getListCmd(),
		getStatusCmd(),
		getUseCmd(),
		getCleanupCmd(),
	)

	// Add global flags
	models.AddGlobalFlags(clusterCmd, utils.GetGlobalFlags().Global)

	return clusterCmd
}
