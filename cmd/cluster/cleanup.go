package cluster

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/spf13/cobra"
)

func getCleanupCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	cleanupCmd := &cobra.Command{
		Use:   "cleanup [NAME]",
		Short: "Prune unused container images from cluster nodes",
		Long: `Reclaim disk space by pruning unused container images inside each cluster node.

Only images no container references are removed. Installed applications, Helm
releases and namespaces are never touched — to remove the OpenFrame platform
use 'openframe app uninstall', to remove the whole cluster use
'openframe cluster delete'.

Examples:
  openframe cluster cleanup
  openframe cluster cleanup my-cluster
  openframe cluster cleanup my-cluster --force`,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"c"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			if err := utils.ValidateGlobalFlags(); err != nil {
				return err
			}
			return models.ValidateCleanupFlags(utils.GetGlobalFlags().Cleanup)
		},
		RunE: utils.WrapCommandWithCommonSetup(runCleanupCluster),
	}

	// Add cleanup-specific flags
	models.AddCleanupFlags(cleanupCmd, utils.GetGlobalFlags().Cleanup)

	return cleanupCmd
}

func runCleanupCluster(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	operationsUI := ui.NewOperationsUI()

	// Get all available clusters
	clusters, err := service.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	// Handle cluster selection with friendly UI (including confirmation)
	globalFlags := utils.GetGlobalFlags()
	clusterName, err := operationsUI.SelectClusterForCleanup(clusters, args, globalFlags.Cleanup.Force)
	if err != nil {
		return err
	}

	// If no cluster selected (e.g., empty list or cancelled), exit gracefully
	if clusterName == "" {
		return nil
	}

	// Show friendly start message
	operationsUI.ShowOperationStart("cleanup", clusterName)

	// Detect cluster type
	clusterType, err := service.DetectClusterType(clusterName)
	if err != nil {
		operationsUI.ShowOperationError("cleanup", clusterName, err)
		return fmt.Errorf("failed to detect cluster type: %w", err)
	}

	// Type-aware prerequisite gate, after the type is known (the group-level
	// generic gate skips cleanup, see cluster.go). Only the k3d path needs
	// tools: a cloud cluster is about to be rejected by the service with a
	// pointer to `cluster delete`, and that message must not be preempted by a
	// Docker demand.
	if clusterType == models.ClusterTypeK3d {
		if err := prerequisites.CheckForClusterType(clusterType); err != nil {
			return err
		}
	}

	// Execute cluster cleanup through service layer. A nil error with failed
	// phases is a partial cleanup: the summary names what was left behind.
	result, err := service.CleanupCluster(cmd.Context(), clusterName, clusterType, utils.GetGlobalFlags().Global.Verbose)
	if err != nil {
		operationsUI.ShowOperationError("cleanup", clusterName, err)
		return err
	}

	operationsUI.ShowCleanupSummary(clusterName, result)
	return nil
}
