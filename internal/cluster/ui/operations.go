package ui

import (
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// OperationsUI provides user-friendly interfaces for cluster operations
type OperationsUI struct{}

// NewOperationsUI creates a new operations UI service
func NewOperationsUI() *OperationsUI {
	return &OperationsUI{}
}

// SelectClusterForOperation provides a friendly interface for selecting a cluster for a specific operation
func (ui *OperationsUI) SelectClusterForOperation(clusters []models.ClusterInfo, args []string, operation string) (string, error) {
	// If cluster name provided as argument, use it directly
	if len(args) > 0 {
		clusterName := strings.TrimSpace(args[0])
		if clusterName == "" {
			return "", fmt.Errorf("cluster name cannot be empty")
		}

		// Validate that the cluster exists in the available clusters
		found := false
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("cluster '%s' not found", clusterName)
		}

		return clusterName, nil
	}

	// Check if clusters are available
	if len(clusters) == 0 {
		ui.ShowNoResourcesMessage("clusters", operation)
		return "", nil
	}

	// Use interactive selection
	clusterName, err := SelectClusterByName(clusters, fmt.Sprintf("Select cluster to %s", operation))
	if err != nil {
		return "", fmt.Errorf("cluster selection failed: %w", err)
	}

	return clusterName, nil
}

// SelectClusterForDelete provides a friendly interface for selecting a cluster to delete with confirmation
func (ui *OperationsUI) SelectClusterForDelete(clusters []models.ClusterInfo, args []string, force bool) (string, error) {
	// If cluster name provided as argument, use it directly
	if len(args) > 0 {
		clusterName := strings.TrimSpace(args[0])
		if clusterName == "" {
			return "", fmt.Errorf("cluster name cannot be empty")
		}

		// Validate that the cluster exists in the available clusters
		// Skip validation when force is set (allows fallback cleanup when k3d list fails)
		if !force {
			found := false
			for _, cluster := range clusters {
				if cluster.Name == clusterName {
					found = true
					break
				}
			}
			if !found {
				return "", fmt.Errorf("cluster '%s' not found", clusterName)
			}
		}

		// Ask for confirmation unless forced
		if !force {
			confirmed, err := ui.confirmDeletion(clusterName)
			if err := errors.WrapConfirmationError(err, "failed to get deletion confirmation"); err != nil {
				return "", err
			}
			if !confirmed {
				pterm.Info.Println("Deletion cancelled.")
				return "", nil
			}
		}

		return clusterName, nil
	}

	// Check if clusters are available
	if len(clusters) == 0 {
		ui.ShowNoResourcesMessage("clusters", "delete")
		return "", nil
	}

	// Use interactive selection
	clusterName, err := SelectClusterByName(clusters, "Select cluster to delete")
	if err != nil {
		return "", fmt.Errorf("cluster selection failed: %w", err)
	}

	if clusterName == "" {
		return "", nil
	}

	// Ask for confirmation unless forced
	if !force {
		confirmed, err := ui.confirmDeletion(clusterName)
		if err := errors.WrapConfirmationError(err, "failed to get deletion confirmation"); err != nil {
			return "", err
		}
		if !confirmed {
			pterm.Info.Println("Deletion cancelled.")
			return "", nil
		}
	}

	return clusterName, nil
}

// SelectClusterForCleanup provides a friendly interface for selecting a cluster for cleanup with confirmation
func (ui *OperationsUI) SelectClusterForCleanup(clusters []models.ClusterInfo, args []string, force bool) (string, error) {
	// If cluster name provided as argument, use it directly
	if len(args) > 0 {
		clusterName := strings.TrimSpace(args[0])
		if clusterName == "" {
			return "", fmt.Errorf("cluster name cannot be empty")
		}

		// Validate that the cluster exists in the available clusters
		found := false
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("cluster '%s' not found", clusterName)
		}

		// Ask for confirmation unless forced (--force is documented as "skip
		// confirmation prompts"; the old behavior of prompting anyway hung CI).
		if !force {
			confirmed, err := ui.confirmCleanup(clusterName)
			if err != nil {
				return "", err
			}
			if !confirmed {
				pterm.Info.Println("Cleanup cancelled.")
				return "", nil
			}
		}

		return clusterName, nil
	}

	// Check if clusters are available
	if len(clusters) == 0 {
		ui.ShowNoResourcesMessage("clusters", "cleanup")
		return "", nil
	}

	// Use interactive selection
	clusterName, err := SelectClusterByName(clusters, "Select cluster to cleanup")
	if err != nil {
		return "", fmt.Errorf("cluster selection failed: %w", err)
	}

	if clusterName == "" {
		return "", nil
	}

	// Ask for confirmation unless forced (same semantics as the argument path).
	if !force {
		confirmed, err := ui.confirmCleanup(clusterName)
		if err != nil {
			return "", err
		}
		if !confirmed {
			pterm.Info.Println("Cleanup cancelled.")
			return "", nil
		}
	}

	return clusterName, nil
}

// confirmCleanup asks for user confirmation before cleaning up a cluster.
// Non-interactive sessions fail fast with a --force hint instead of blocking.
func (ui *OperationsUI) confirmCleanup(clusterName string) (bool, error) {
	return sharedUI.RequireConfirmation(
		fmt.Sprintf("Are you sure you want to cleanup cluster '%s'?", pterm.Cyan(clusterName)),
		"--force", false)
}

// confirmDeletion asks for user confirmation before deleting a cluster.
// Non-interactive sessions fail fast with a --force hint instead of blocking.
func (ui *OperationsUI) confirmDeletion(clusterName string) (bool, error) {
	if sharedUI.IsNonInteractive() {
		return false, fmt.Errorf("confirmation required but the session is non-interactive; re-run with --force")
	}
	return sharedUI.ConfirmDeletion("cluster", clusterName)
}

// ShowOperationStart displays a friendly message when starting an operation
func (ui *OperationsUI) ShowOperationStart(operation, clusterName string) {
	switch strings.ToLower(operation) {
	case "cleanup":
		pterm.Info.Printf("Cleaning up cluster '%s'...\n", pterm.Cyan(clusterName))
	case "delete":
		pterm.Info.Printf("Deleting cluster '%s'...\n", pterm.Cyan(clusterName))
	default:
		pterm.Info.Printf("Processing '%s' for cluster '%s'...\n", operation, pterm.Cyan(clusterName))
	}
}

// ShowCleanupSummary reports what cleanup actually removed.
//
// It deliberately does not print a fixed list of accomplishments: cleanup is
// best-effort, every phase can fail independently, and the previous summary
// ("Removed unused Docker images / Freed up disk space / Optimized cluster
// performance") was printed verbatim even when nothing was removed and every
// phase had failed.
func (ui *OperationsUI) ShowCleanupSummary(clusterName string, result models.CleanupResult) {
	if result.Partial() {
		pterm.Warning.Printf("Cluster '%s' cleanup finished with problems\n", pterm.Cyan(clusterName))
	} else {
		pterm.Success.Printf("Cluster '%s' cleanup completed\n", pterm.Cyan(clusterName))
	}

	// DefaultBasicText, not bare pterm.Printf/fmt.Println: those write straight
	// to stdout and survive --silent, whose contract is "errors only".
	pterm.DefaultBasicText.Println()
	if result.Removed() == 0 {
		pterm.Info.Println("Nothing to remove: the cluster had no OpenFrame resources left.")
	} else {
		pterm.Info.Printf("Removed:\n")
		for _, line := range []struct {
			n     int
			label string
		}{
			{result.ApplicationsDeleted, "ArgoCD application(s)"},
			{result.FinalizersCleared, "stuck application finalizer(s) cleared"},
			{result.ReleasesRemoved, "Helm release(s)"},
			{result.NamespacesDeleted, "namespace(s)"},
			{result.NodesPruned, "node(s) pruned (images, containers, volumes, networks)"},
		} {
			if line.n > 0 {
				pterm.DefaultBasicText.Printf("  %d %s\n", line.n, line.label)
			}
		}
	}

	if result.Partial() {
		pterm.DefaultBasicText.Println()
		pterm.Warning.Printf("These phases did not complete; some resources may remain:\n")
		for _, f := range result.Failures {
			pterm.DefaultBasicText.Printf("  • %s\n", f)
		}
		pterm.Info.Printf("Re-run with --force, or delete the cluster: openframe cluster delete %s\n", clusterName)
	}
}

// ShowOperationSuccess displays a friendly success message
func (ui *OperationsUI) ShowOperationSuccess(operation, clusterName string) {
	switch strings.ToLower(operation) {
	case "delete":
		pterm.Success.Printf("Cluster '%s' deleted successfully\n", pterm.Cyan(clusterName))

		// Show detailed deletion box
		pterm.DefaultBasicText.Println()
		boxContent := fmt.Sprintf(
			"NAME:         %s\n"+
				"TYPE:         %s\n"+
				"STATUS:       %s\n"+
				"NETWORK:      %s\n"+
				"RESOURCES:    %s",
			pterm.Bold.Sprint(clusterName),
			"k3d",
			pterm.Red("Deleted"),
			pterm.Gray("Removed"),
			pterm.Gray("Cleaned up"),
		)

		pterm.DefaultBox.
			WithTitle(" Cluster Deleted ").
			WithTitleTopCenter().
			Println(boxContent)

		// Show deletion summary
		pterm.DefaultBasicText.Println()
		pterm.Info.Printf("Deletion Summary:\n")
		pterm.DefaultBasicText.Printf("  Cluster and nodes removed\n")
		pterm.DefaultBasicText.Printf("  Docker containers cleaned up\n")
		pterm.DefaultBasicText.Printf("  Network configuration removed\n")
		pterm.DefaultBasicText.Printf("  Kubeconfig entries cleaned\n")

	default:
		pterm.Success.Printf("Operation '%s' completed for cluster '%s'\n", operation, pterm.Cyan(clusterName))
	}
	pterm.DefaultBasicText.Println()
}

// ShowOperationError displays a friendly error message
func (ui *OperationsUI) ShowOperationError(operation, clusterName string, err error) {
	troubleshootingTips := []sharedUI.TroubleshootingTip{
		{Description: "Check cluster exists:", Command: "openframe cluster list"},
		{Description: "Check cluster status:", Command: "openframe cluster status " + clusterName},
		{Description: "Try with verbose output:", Command: "openframe cluster " + operation + " " + clusterName + " --verbose"},
	}

	sharedUI.ShowOperationError(operation, clusterName, err, troubleshootingTips)
}

// ShowConfigurationSummary displays the cluster configuration summary
func (ui *OperationsUI) ShowConfigurationSummary(config models.ClusterConfig, dryRun bool, skipWizard bool) {
	pterm.Info.Printf("Configuration Summary\n")

	// Clean, simple format without heavy table styling. DefaultBasicText (not
	// raw fmt): --silent redirects it, while raw fmt leaked these lines into
	// "silent" output (verification report saw the leak and graded it silent).
	pterm.DefaultBasicText.Printf("   Name: %s\n", pterm.Cyan(config.Name))
	pterm.DefaultBasicText.Printf("   Type: %s\n", string(config.Type))
	pterm.DefaultBasicText.Printf("  Nodes: %d\n", config.NodeCount)

	if config.K8sVersion != "" {
		pterm.DefaultBasicText.Printf("Version: %s\n", config.K8sVersion)
	}

	pterm.DefaultBasicText.Println()

	if dryRun {
		pterm.Warning.Println("DRY RUN MODE - No cluster will be created")
	} else if skipWizard {
		pterm.Info.Println("Proceeding with cluster creation...")
	}
}

// ShowNoResourcesMessage displays a friendly message when no clusters are available
func (ui *OperationsUI) ShowNoResourcesMessage(resourceType, operation string) {
	sharedUI.ShowNoResourcesMessage(
		resourceType,
		operation,
		"openframe cluster create",
		"openframe cluster list",
	)
}
