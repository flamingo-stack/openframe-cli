package app

import (
	"fmt"

	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// getStatusCmd returns the status subcommand.
func getStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the OpenFrame platform status (cluster, apps, access)",
		Long: `Report whether OpenFrame is up and running on a cluster.

Checks the cluster is reachable, lists the ArgoCD applications with their
sync/health, summarizes overall readiness, and prints how to sign in.

Examples:
  openframe app status
  openframe app status --context k3d-openframe-dev`,
		RunE: runStatusCommand,
	}
	cmd.Flags().String("context", "", "Kube-context to use (defaults to the current context)")
	return cmd
}

func runStatusCommand(cmd *cobra.Command, _ []string) error {
	verbose := getVerboseFlag(cmd)
	contextName, _ := cmd.Flags().GetString("context")

	cfg, err := resolveRestConfig(contextName)
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("could not connect to the cluster: %w", err), verbose)
	}

	mgr, err := newArgoCDManager(contextName, verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	accessor, err := k8s.NewAccessorForConfig(cfg)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	rep, err := appstatus.NewService(mgr, accessor, mgr).Report(cmd.Context(), verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("could not read platform status: %w", err), verbose)
	}

	renderStatus(rep)
	return nil
}

func renderStatus(rep appstatus.Report) {
	if !rep.Health.Reachable {
		pterm.Error.Println("Cluster is not reachable. Is it running and is your kube-context correct?")
		return
	}
	pterm.Success.Printf("Cluster reachable (%d/%d nodes ready)\n", rep.Health.NodesReady, rep.Health.NodesTotal)

	if rep.Total == 0 {
		pterm.Warning.Println("No OpenFrame applications found — is it installed? Run: openframe app install")
		return
	}

	table := pterm.TableData{{"APPLICATION", "SYNC", "HEALTH"}}
	for _, a := range rep.Apps {
		table = append(table, []string{a.Name, a.Sync, a.Health})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(table).Render()

	line := rep.Summary()
	if rep.Ready() {
		pterm.Success.Println(line)
	} else {
		pterm.Warning.Println(line)
	}

	if rep.AdminPassword != "" {
		printAccess(rep.AdminPassword)
	}
}
