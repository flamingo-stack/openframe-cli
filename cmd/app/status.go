package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	statustui "github.com/flamingo-stack/openframe-cli/internal/app/status/tui"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
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
	cmd.Flags().StringP("context", "c", "", "Kube-context to use (defaults to the current context)")
	cmd.Flags().BoolP("watch", "w", false, "Keep the status on screen, refreshing every few seconds (Ctrl+C to exit)")
	cmd.Flags().BoolP("interactive", "i", false, "Open the interactive view: navigate apps, inspect details, trigger syncs")
	addOutputFlag(cmd)
	return cmd
}

func runStatusCommand(cmd *cobra.Command, _ []string) error {
	verbose := getVerboseFlag(cmd)
	contextName, _ := cmd.Flags().GetString("context")
	watch, _ := cmd.Flags().GetBool("watch")
	interactive, _ := cmd.Flags().GetBool("interactive")
	format, err := outputFormat(cmd)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	if (watch || interactive) && format != "text" {
		return fmt.Errorf("--watch/--interactive are live terminal views and cannot combine with --output %s", format)
	}
	if (watch || interactive) && (!ui.IsTerminal() || ui.IsPlain()) {
		return fmt.Errorf("--watch/--interactive need an interactive terminal (and cannot combine with --plain)")
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
	svc := appstatus.NewService(mgr, accessor, mgr)

	if interactive {
		return statustui.Run(cmd.Context(), svc, mgr.SyncApplications, verbose)
	}
	if watch {
		return watchStatus(cmd.Context(), svc, verbose)
	}

	rep, err := svc.Report(cmd.Context(), verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("could not read platform status: %w", err), verbose)
	}

	if format != "text" {
		return renderMachine(format, statusToJSON(rep))
	}
	renderStatus(rep)
	return nil
}

// watchStatus re-renders the status in place every few seconds — a
// lightweight `kubectl get -w` for the whole platform. A failing poll shows
// its error inside the view and keeps watching; Ctrl+C ends the watch
// cleanly via the signal-cancelled context.
func watchStatus(ctx context.Context, svc *appstatus.Service, verbose bool) error {
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return err
	}
	defer func() { _ = area.Stop() }()

	const interval = 3 * time.Second
	for {
		rep, rerr := svc.Report(ctx, verbose)
		if ctx.Err() != nil {
			return nil
		}
		area.Update(renderWatchFrame(rep, rerr))
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
		}
	}
}

// renderWatchFrame builds one --watch screen: header with the wall clock,
// reachability line, the application table, and the readiness summary.
func renderWatchFrame(rep appstatus.Report, rerr error) string {
	g := ui.Glyphs()
	var b strings.Builder
	fmt.Fprintf(&b, "%s OpenFrame status %s %s %s refresh 3s %s Ctrl+C to exit\n",
		pterm.FgCyan.Sprint(g.Running), g.Bullet, time.Now().Format("15:04:05"), g.Bullet, g.Bullet)
	if rerr != nil {
		b.WriteString(pterm.FgRed.Sprintf("%s could not read platform status: %v\n", g.Fail, rerr))
		return b.String()
	}
	switch {
	case rep.Health.Reachable:
		fmt.Fprintf(&b, "%s Cluster reachable (%d/%d nodes ready)\n",
			pterm.FgGreen.Sprint(g.OK), rep.Health.NodesReady, rep.Health.NodesTotal)
	case rep.Total > 0:
		fmt.Fprintf(&b, "%s Node health unavailable (apps still listed)\n", pterm.FgYellow.Sprint(g.Warn))
	default:
		fmt.Fprintf(&b, "%s Cluster is not reachable\n", pterm.FgRed.Sprint(g.Fail))
		return b.String()
	}
	if rep.Total == 0 {
		b.WriteString("No OpenFrame applications found — is it installed?\n")
		return b.String()
	}
	table := pterm.TableData{{"APPLICATION", "SYNC", "HEALTH"}}
	for _, a := range rep.Apps {
		table = append(table, []string{a.Name, a.Sync, a.Health})
	}
	if rendered, terr := pterm.DefaultTable.WithHasHeader().WithData(table).Srender(); terr == nil {
		b.WriteString(rendered + "\n")
	}
	line := rep.Summary()
	if rep.Ready() {
		b.WriteString(pterm.FgGreen.Sprintf("%s %s\n", g.OK, line))
	} else {
		b.WriteString(pterm.FgYellow.Sprintf("%s %s\n", g.Warn, line))
	}
	return b.String()
}

// statusAppJSON is the machine-readable shape of a single application.
type statusAppJSON struct {
	Name   string `json:"name"`
	Sync   string `json:"sync"`
	Health string `json:"health"`
}

// statusJSON is the machine-readable shape of `app status`.
type statusJSON struct {
	Reachable  bool `json:"reachable"`
	NodesReady int  `json:"nodesReady"`
	NodesTotal int  `json:"nodesTotal"`
	// HealthError says WHY reachable is false (e.g. an RBAC-denied node read
	// next to a fully populated applications array) — without it that
	// combination read as self-contradictory output.
	HealthError  string          `json:"healthError,omitempty"`
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
	healthErr := ""
	if rep.HealthErr != nil {
		healthErr = rep.HealthErr.Error()
	}
	return statusJSON{
		Reachable:    rep.Health.Reachable,
		HealthError:  healthErr,
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
	switch {
	case !rep.Health.Reachable && rep.Total == 0:
		pterm.Error.Println("Cluster is not reachable. Is it running and is your kube-context correct?")
		if rep.HealthErr != nil {
			pterm.Error.Printf("  cause: %v\n", rep.HealthErr)
		}
		return
	case !rep.Health.Reachable:
		// The application list DID come back, so the API server answered — only
		// the node read failed (commonly RBAC: a namespace-scoped credential
		// has no cluster-wide "nodes list"). Say that and keep the data instead
		// of claiming the cluster is down and hiding a healthy platform.
		msg := "Cluster node health could not be read"
		if rep.HealthErr != nil {
			msg += ": " + rep.HealthErr.Error()
		}
		pterm.Warning.Println(msg)
	default:
		pterm.Success.Printf("Cluster reachable (%d/%d nodes ready)\n", rep.Health.NodesReady, rep.Health.NodesTotal)
	}

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
