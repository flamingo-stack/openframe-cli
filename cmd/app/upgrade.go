package app

import (
	"context"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/app/target"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/services"
	"github.com/flamingo-stack/openframe-cli/internal/chart/ui/templates"
	chartconfig "github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	charttypes "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// getUpgradeCmd returns the upgrade subcommand.
func getUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade [cluster-name]",
		Short: "Upgrade the OpenFrame platform (change ref or force re-sync)",
		Long: `Upgrade an already-installed OpenFrame platform.

Two modes:

  1. Change ref (--ref): re-deploy the app-of-apps at a new git ref (branch or
     release tag), then let ArgoCD roll it out. Runs non-interactively against
     the existing helm-values.yaml — no config wizard and no certificate
     regeneration. The deployment mode is inferred from helm-values for OSS;
     SaaS installs must pass --deployment-mode (saas-tenant / saas-shared).

  2. Force re-sync (default, or --sync): keep the current ref but force ArgoCD
     to re-read git and sync the platform. Use it to pull a moved floating ref
     (e.g. the main branch advanced) when auto-sync is off or manifests are
     stale-cached. Pruning is OFF by default; pass --prune to also delete
     resources removed from git (destructive).

Examples:
  openframe app upgrade                          # Force re-sync current ref (default)
  openframe app upgrade --sync                   # Same, explicit
  openframe app upgrade --sync --prune           # Also delete resources removed from git
  openframe app upgrade --ref v1.3.0             # Upgrade to a release tag (OSS inferred)
  openframe app upgrade --ref v1.3.0 --deployment-mode saas-tenant
  openframe app upgrade --ref main --dry-run     # Preview a ref change
  openframe app upgrade my-cluster --context k3d-my-cluster`,
		RunE:          runUpgradeCommand,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	addInstallFlags(cmd)
	cmd.Flags().Bool("sync", false, "Force ArgoCD to refresh and re-sync the current ref (Mode 2)")
	cmd.Flags().Bool("prune", false, "During force-sync, delete resources no longer present in git (destructive)")

	return cmd
}

// runUpgradeCommand dispatches to change-ref (Mode 1) or force-sync (Mode 2).
func runUpgradeCommand(cmd *cobra.Command, args []string) error {
	flags, err := extractInstallFlags(cmd)
	if err != nil {
		return err
	}
	verbose := getVerboseFlag(cmd)
	sync, _ := cmd.Flags().GetBool("sync")

	if upgradeIsChangeRef(cmd.Flags().Changed("ref"), cmd.Flags().Changed("github-branch"), sync) {
		return runUpgradeChangeRef(cmd, args, flags, verbose)
	}
	return runUpgradeForceSync(cmd, args, flags, verbose)
}

// upgradeIsChangeRef decides the upgrade mode: a changed --ref/--github-branch
// means "deploy this ref" (Mode 1); otherwise, or when --sync is explicit,
// force-sync the current ref (Mode 2).
func upgradeIsChangeRef(refChanged, branchChanged, sync bool) bool {
	return (refChanged || branchChanged) && !sync
}

// runUpgradeChangeRef re-deploys the platform at a new ref using the EXISTING
// configuration (Mode 1). It runs non-interactively — no config wizard and no
// certificate regeneration — so a scripted ref bump does not re-prompt or rotate
// TLS certs. The deployment mode comes from --deployment-mode or is inferred
// from the existing helm-values (OSS only; SaaS requires the explicit flag).
func runUpgradeChangeRef(cmd *cobra.Command, args []string, flags *InstallFlags, verbose bool) error {
	flags.Force = true
	flags.NonInteractive = true
	if flags.DeploymentMode == "" {
		mode, err := deploymentModeForUpgrade()
		if err != nil {
			return sharedErrors.HandleGlobalError(err, verbose)
		}
		flags.DeploymentMode = mode
	}

	req, err := buildInstallRequest(cmd, args, flags, verbose, "Upgrading")
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	pterm.Info.Printf("Upgrading OpenFrame to ref %q\n", flags.resolvedRef())
	if err := services.InstallChartsWithConfigContext(cmd.Context(), req); err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

// deploymentModeForUpgrade infers the CLI deployment-mode string from the
// existing helm-values.yaml when --deployment-mode was not given. OSS is
// unambiguous (and the fallback when no values exist); SaaS values cannot be
// told apart as saas-tenant vs saas-shared, so those require the explicit flag.
func deploymentModeForUpgrade() (string, error) {
	modifier := templates.NewHelmValuesModifier()
	values, err := modifier.LoadOrCreateBaseValues()
	if err != nil {
		return "", fmt.Errorf("reading helm-values.yaml to infer deployment mode: %w", err)
	}
	if modifier.GetCurrentDeploymentMode(values) == charttypes.DeploymentModeOSS {
		return "oss-tenant", nil
	}
	return "", fmt.Errorf("cannot infer SaaS deployment mode from helm-values (saas-tenant vs saas-shared) — pass --deployment-mode")
}

// runUpgradeForceSync refreshes and re-syncs the current ref via ArgoCD (Mode 2).
func runUpgradeForceSync(cmd *cobra.Command, args []string, flags *InstallFlags, verbose bool) error {
	cfg, clusterName, err := resolveUpgradeTarget(cmd, args, flags, verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	manager, err := argocd.NewManagerWithConfig(executor.NewRealCommandExecutor(flags.DryRun, verbose), cfg)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	prune, _ := cmd.Flags().GetBool("prune")

	if flags.DryRun {
		// HandleGlobalError is a no-op on nil, so this returns nil on success.
		return sharedErrors.HandleGlobalError(previewOutOfSync(cmd.Context(), manager, verbose, prune), verbose)
	}

	if prune {
		pterm.Warning.Println("Refreshing and syncing with --prune: resources removed from git will be DELETED.")
	}
	pterm.Info.Println("Refreshing and syncing the OpenFrame platform...")
	if err := manager.RefreshAndSync(cmd.Context(), prune); err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	waitCfg := chartconfig.ChartInstallConfig{
		ClusterName:    clusterName,
		Verbose:        verbose,
		NonInteractive: flags.NonInteractive,
	}
	if err := manager.WaitForApplications(cmd.Context(), waitCfg); err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	pterm.Success.Println("OpenFrame platform re-synced.")
	return nil
}

// previewOutOfSync lists applications and reports which are not Synced, without
// triggering a sync (Mode 2 --dry-run). prune reflects whether the real run
// would delete resources removed from git, so the preview can warn about it.
func previewOutOfSync(ctx context.Context, manager *argocd.Manager, verbose, prune bool) error {
	apps, err := manager.ListApplications(ctx, verbose)
	if err != nil {
		return err
	}
	if len(apps) == 0 {
		pterm.Warning.Println("No ArgoCD applications found — is OpenFrame installed?")
		return nil
	}
	outOfSync := 0
	for _, a := range apps {
		if a.Sync != argocd.ArgoCDSyncSynced {
			outOfSync++
			pterm.Info.Printf("  OutOfSync: %s (health=%s, sync=%s)\n", a.Name, a.Health, a.Sync)
		}
	}
	if outOfSync == 0 {
		pterm.Success.Printf("All %d applications are Synced; a force-sync would be a no-op.\n", len(apps))
	} else {
		pterm.Info.Printf("%d/%d applications would be re-synced.\n", outOfSync, len(apps))
	}
	if prune {
		pterm.Warning.Println("--prune is set: the real run would DELETE any resources no longer present in git.")
	}
	return nil
}

// resolveUpgradeTarget resolves the rest.Config (and cluster name, if any) for
// the force-sync path: --context, then a named cluster (k3d-<name>), then the
// current context for --non-interactive, else an interactive context prompt.
func resolveUpgradeTarget(cmd *cobra.Command, args []string, flags *InstallFlags, verbose bool) (*rest.Config, string, error) {
	path := k8s.DefaultKubeconfigPath()

	if contextName, _ := cmd.Flags().GetString("context"); contextName != "" {
		cfg, err := k8s.RestConfigForContext(path, contextName)
		if err != nil {
			return nil, "", fmt.Errorf("could not use context %q: %w", contextName, err)
		}
		return cfg, clusterNameArg(args), nil
	}

	if name := clusterNameArg(args); name != "" {
		cfg, err := k8s.RestConfigForContext(path, k8s.ResolveContextForCluster(path, name))
		if err != nil {
			return nil, "", fmt.Errorf("could not use cluster %q: %w", name, err)
		}
		return cfg, name, nil
	}

	if flags.NonInteractive {
		cfg, err := k8s.RestConfigForContext(path, "") // current context
		if err != nil {
			return nil, "", err
		}
		return cfg, "", nil
	}

	sel := target.NewSelector(target.UIPrompter{}, recommendedRequirements())
	res, err := sel.Select(cmd.Context())
	if err != nil {
		return nil, "", err
	}
	return res.Config, "", nil
}

// clusterNameArg returns the first positional arg (cluster name) or "".
func clusterNameArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
