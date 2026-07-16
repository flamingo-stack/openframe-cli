package cluster

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func getDeleteCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	deleteCmd := &cobra.Command{
		Use:   "delete [NAME]",
		Short: "Delete a Kubernetes cluster",
		Long: `Delete a Kubernetes cluster and clean up all associated resources.

Stops intercepts, deletes cluster, cleans up Docker resources,
and removes cluster configuration.

Examples:
  openframe cluster delete my-cluster
  openframe cluster delete my-cluster --force
  openframe cluster delete  # interactive selection`,
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			if err := utils.ValidateGlobalFlags(); err != nil {
				return err
			}
			globalFlags := utils.GetGlobalFlags()
			if globalFlags != nil && globalFlags.Delete != nil {
				return models.ValidateDeleteFlags(globalFlags.Delete)
			}
			return nil
		},
		RunE: utils.WrapCommandWithCommonSetup(runDeleteCluster),
	}

	// Add delete-specific flags
	globalFlags := utils.GetGlobalFlags()
	if globalFlags != nil && globalFlags.Delete != nil {
		models.AddDeleteFlags(deleteCmd, globalFlags.Delete)
	}

	return deleteCmd
}

// confirmCloudDeletion is the stronger destroy gate for cloud clusters:
// re-typing the cluster name. Local clusters and --force pass through
// (--force is the CI escape hatch); a non-interactive session without --force
// refuses rather than hanging on a prompt — defense in depth behind the
// generic confirmation, which may not always run before this.
func confirmCloudDeletion(clusterType models.ClusterType, clusterName string, force bool) (bool, error) {
	if clusterType != models.ClusterTypeEKS && clusterType != models.ClusterTypeGKE {
		return true, nil
	}
	if force {
		return true, nil
	}
	if sharedUI.IsNonInteractive() {
		return false, fmt.Errorf("refusing to destroy cloud cluster '%s' non-interactively; pass --force to confirm", clusterName)
	}
	return ui.ConfirmTypedClusterName(clusterName)
}

func runDeleteCluster(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	operationsUI := ui.NewOperationsUI()

	// Get all available clusters
	clusters, err := service.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	// Handle cluster selection with friendly UI (including confirmation)
	globalFlags := utils.GetGlobalFlags()
	clusterName, err := operationsUI.SelectClusterForDelete(clusters, args, globalFlags.Delete.Force)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, globalFlags.Global.Verbose)
	}

	// If no cluster selected (e.g., empty list or cancelled), exit gracefully
	if clusterName == "" {
		return nil
	}

	// Show friendly start message
	operationsUI.ShowOperationStart("delete", clusterName)

	// Detect cluster type
	clusterType, err := service.DetectClusterType(clusterName)
	if err != nil {
		operationsUI.ShowOperationError("delete", clusterName, err)
		return fmt.Errorf("failed to detect cluster type: %w", err)
	}

	// Destroying a cloud cluster deletes billed infrastructure irreversibly,
	// so it takes a stronger gate than the generic yes/no above.
	proceed, err := confirmCloudDeletion(clusterType, clusterName, globalFlags.Delete.Force)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, globalFlags.Global.Verbose)
	}
	if !proceed {
		pterm.Info.Println("Cluster name did not match — nothing was deleted")
		return nil
	}

	// Execute cluster deletion through service layer
	err = service.DeleteCluster(cmd.Context(), clusterName, clusterType, globalFlags.Delete.Force)
	if err != nil {
		operationsUI.ShowOperationError("delete", clusterName, err)
		return sharedErrors.HandleGlobalError(err, globalFlags.Global.Verbose)
	}

	// Show friendly success message
	operationsUI.ShowOperationSuccess("delete", clusterName)
	return nil
}
