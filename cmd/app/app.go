package app

import (
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/spf13/cobra"
)

// GetAppCmd returns the app command and its subcommands.
func GetAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Deploy the OpenFrame application onto a cluster",
		Long: `App Management - Install the OpenFrame application (ArgoCD + apps)

This command group deploys the OpenFrame application onto a Kubernetes cluster:
  • install - Install ArgoCD and the app-of-apps

Requires an existing, online cluster — one created with 'openframe cluster
create', made by you directly, or any other reachable cluster.

Examples:
  openframe app install
  openframe app install my-cluster`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// This command group defines its own PersistentPreRunE, which shadows
			// the root's, so honor --silent here too.
			if s, _ := cmd.Flags().GetBool("silent"); s {
				ui.SetSilent()
			}
			// Machine output (json/yaml): no logo, clean stdout for scripts.
			if isMachineOutput(cmd) {
				return nil
			}
			// Show logo for subcommands, but not for the root app command.
			if cmd.Use != "app" {
				ui.ShowLogoWithContext(cmd.Context())
			}
			// Prerequisites are checked ONCE inside the install/upgrade flow
			// (InstallChartsWithConfigContext), not here — so the check no longer
			// runs twice for install/upgrade, nor needlessly for uninstall (which
			// only needs helm + a reachable cluster, not the cert/k3d installer).
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show logo when no subcommand is provided
			ui.ShowLogoWithContext(cmd.Context())
			return cmd.Help()
		},
	}

	cmd.AddCommand(getInstallCmd())
	cmd.AddCommand(getUpgradeCmd())
	cmd.AddCommand(getStatusCmd())
	cmd.AddCommand(getAccessCmd())
	cmd.AddCommand(getUninstallCmd())
	return cmd
}
