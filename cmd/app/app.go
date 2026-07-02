package app

import (
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/spf13/cobra"
)

// GetAppCmd returns the app command and its subcommands.
// "chart" is kept as a hidden alias for backward compatibility.
func GetAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app",
		Aliases: []string{"chart", "c"},
		Short:   "Deploy the OpenFrame application onto a cluster",
		Long: `App Management - Install the OpenFrame application (ArgoCD + apps)

This command group deploys the OpenFrame application onto a Kubernetes cluster:
  • install - Install ArgoCD and the app-of-apps

Requires an existing, online cluster — one created with 'openframe cluster
create', made by you directly, or any other reachable cluster.

Examples:
  openframe app install
  openframe app install my-cluster`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// JSON output is machine mode: no logo, no interactive prerequisite
			// gate, so stdout stays clean and scripts never hit a prompt.
			if isJSONOutput(cmd) {
				return nil
			}
			// Show logo for subcommands, but not for the root app command
			if cmd.Use != "app" {
				ui.ShowLogoWithContext(cmd.Context())
			}
			// Read-only commands (status, access) talk to an existing cluster via
			// client-go and never install local tooling, so they skip the
			// interactive prerequisite gate — which could otherwise prompt to
			// install helm/k3d and hang a script.
			if cmd.Annotations["readonly"] == "true" {
				return nil
			}
			return prerequisites.NewInstaller().CheckAndInstall()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show logo when no subcommand is provided
			ui.ShowLogoWithContext(cmd.Context())
			return cmd.Help()
		},
	}

	cmd.AddCommand(getInstallCmd())
	cmd.AddCommand(getStatusCmd())
	cmd.AddCommand(getAccessCmd())
	cmd.AddCommand(getUninstallCmd())
	return cmd
}
