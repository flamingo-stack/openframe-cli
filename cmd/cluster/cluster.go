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

Supports K3d clusters for local development and Google GKE for cloud deployments (AWS EKS coming soon).

Examples:
  openframe cluster create
  openframe cluster delete`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// This command group defines its own PersistentPreRunE, which shadows
			// the root's, so honor --silent here too.
			if s, _ := cmd.Flags().GetBool("silent"); s {
				ui.SetSilent()
			}
			// Machine output (json/yaml) is machine mode: no logo, no prerequisite
			// gate, so stdout stays clean for scripts.
			if out, _ := cmd.Flags().GetString("output"); out == "json" || out == "yaml" {
				return nil
			}
			// Show logo for subcommands, but not for the root cluster command
			if cmd.Use != "cluster" {
				ui.ShowLogoWithContext(cmd.Context())
			}
			// create runs its own type-aware gate after the cluster type is known
			// (a cloud cluster must not demand Docker/k3d); use only flips local
			// kubeconfig/gcloud state and needs no tools at all. The other
			// subcommands are k3d-scoped, so the k3d gate stays here.
			if cmd.Name() == "create" || cmd.Name() == "use" {
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
