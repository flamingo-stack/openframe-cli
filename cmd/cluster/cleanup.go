package cluster

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func getCleanupCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	cleanupCmd := &cobra.Command{
		Use:   "cleanup [NAME]",
		Short: "Clean up unused cluster resources",
		Long: `Remove unused images and resources from cluster nodes.

Cleans up Docker images and resources, freeing disk space.
Useful for development clusters with many builds.

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

	// Inject the ArgoCD-backed application cleaner (composition root: only the
	// command layer may import both the cluster and the chart subsystems).
	// Without it, cleanup skips the Application delete/finalizer-strip phases and
	// the argocd namespace can stay stuck in Terminating. Best-effort: a cluster
	// that is unreachable or has no ArgoCD simply cleans up without it.
	if cfg, cerr := service.GetRestConfig(clusterName); cerr == nil {
		if mgr, merr := argocd.NewManagerWithConfig(executor.NewRealCommandExecutor(false, globalFlags.Global.Verbose), cfg); merr == nil {
			service = service.WithApplicationCleaner(mgr)
		} else if globalFlags.Global.Verbose {
			pterm.Warning.Printf("ArgoCD cleanup unavailable: %v\n", merr)
		}
	} else if globalFlags.Global.Verbose {
		pterm.Warning.Printf("Cluster not reachable for ArgoCD cleanup: %v\n", cerr)
	}

	// Execute cluster cleanup through service layer. A nil error with failed
	// phases is a partial cleanup: the summary names what was left behind.
	result, err := service.CleanupCluster(cmd.Context(), clusterName, clusterType, utils.GetGlobalFlags().Global.Verbose, utils.GetGlobalFlags().Cleanup.Force)
	if err != nil {
		operationsUI.ShowOperationError("cleanup", clusterName, err)
		return err
	}

	operationsUI.ShowCleanupSummary(clusterName, result)
	return nil
}
