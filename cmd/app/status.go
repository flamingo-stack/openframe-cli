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
		RunE:        runStatusCommand,
		Annotations: map[string]string{"readonly": "true"},
	}
	cmd.Flags().String("context", "", "Kube-context to use (defaults to the current context)")
	addOutputFlag(cmd)
	return cmd
}

func runStatusCommand(cmd *cobra.Command, _ []string) error {
	verbose := getVerboseFlag(cmd)
	contextName, _ := cmd.Flags().GetString("context")
	format, err := outputFormat(cmd)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

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

	if format == "json" {
		return printJSON(statusToJSON(rep))
	}
	renderStatus(rep)
	return nil
}

// statusAppJSON is the machine-readable shape of a single application.
type statusAppJSON struct {
	Name   string `json:"name"`
	Sync   string `json:"sync"`
	Health string `json:"health"`
}

// statusJSON is the machine-readable shape of `app status`.
type statusJSON struct {
	Reachable    bool            `json:"reachable"`
	NodesReady   int             `json:"nodesReady"`
	NodesTotal   int             `json:"nodesTotal"`
	Ready        bool            `json:"ready"`
	Summary      string          `json:"summary"`
	Total        int             `json:"total"`
	Synced       int             `json:"synced"`
	Healthy      int             `json:"healthy"`
	Applications []statusAppJSON `json:"applications"`
}

func statusToJSON(rep appstatus.Report) statusJSON {
	apps := make([]statusAppJSON, 0, len(rep.Apps))
	for _, a := range rep.Apps {
		apps = append(apps, statusAppJSON{Name: a.Name, Sync: a.Sync, Health: a.Health})
	}
	return statusJSON{
		Reachable:    rep.Health.Reachable,
		NodesReady:   rep.Health.NodesReady,
		NodesTotal:   rep.Health.NodesTotal,
		Ready:        rep.Ready(),
		Summary:      rep.Summary(),
		Total:        rep.Total,
		Synced:       rep.Synced,
		Healthy:      rep.Healthy,
		Applications: apps,
	}
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
