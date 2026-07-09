package app

import (
	"fmt"

	appuninstall "github.com/flamingo-stack/openframe-cli/internal/app/uninstall"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/helm"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// getUninstallCmd returns the uninstall subcommand.
func getUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove OpenFrame (ArgoCD + apps) from a cluster, keeping the cluster",
		Long: `Remove the OpenFrame application from a cluster.

Deletes the ArgoCD applications and uninstalls the ArgoCD and app-of-apps Helm
releases. The cluster itself is NOT deleted — use 'openframe cluster delete' for
that. This is destructive and asks for confirmation unless --yes is given.

Examples:
  openframe app uninstall
  openframe app uninstall --context k3d-openframe-dev
  openframe app uninstall --yes --delete-namespace`,
		RunE: runUninstallCommand,
	}
	cmd.Flags().StringP("context", "c", "", "Kube-context to use (defaults to the current context)")
	cmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt (for automation)")
	cmd.Flags().Bool("delete-namespace", false, "Also delete the argocd namespace")
	return cmd
}

func runUninstallCommand(cmd *cobra.Command, _ []string) error {
	verbose := getVerboseFlag(cmd)
	contextName, _ := cmd.Flags().GetString("context")
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	deleteNS, _ := cmd.Flags().GetBool("delete-namespace")

	target := "the current kube-context"
	if contextName != "" {
		target = fmt.Sprintf("context %q", contextName)
	}

	if !skipConfirm {
		ok, err := ui.ConfirmActionInteractive(
			fmt.Sprintf("Remove ArgoCD and all OpenFrame apps from %s? The cluster itself is kept.", target), false)
		if err != nil {
			return sharedErrors.HandleGlobalError(err, verbose)
		}
		if !ok {
			pterm.Info.Println("Uninstall cancelled.")
			return nil
		}
	}

	cfg, err := resolveRestConfig(contextName)
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("could not connect to the cluster: %w", err), verbose)
	}
	mgr, err := newArgoCDManager(contextName, verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	helmMgr, err := helm.NewHelmManager(executor.NewRealCommandExecutor(false, verbose), cfg, verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	pterm.Info.Println("Uninstalling OpenFrame...")
	res, err := appuninstall.NewService(mgr, helmMgr, mgr, contextName).
		Uninstall(cmd.Context(), appuninstall.Options{DeleteNamespace: deleteNS})
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("uninstall failed: %w", err), verbose)
	}

	pterm.Success.Printf("Removed %d application(s) and %d Helm release(s).\n", res.AppsDeleted, len(res.ReleasesRemoved))
	if res.NamespaceDeleted {
		pterm.Success.Println("Deleted the argocd namespace.")
	}
	pterm.Info.Println("The cluster was left running. Re-install any time with: openframe app install")
	return nil
}
