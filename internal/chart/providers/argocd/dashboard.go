package argocd

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/steps"
	"github.com/pterm/pterm"
)

// appDashboard is the live wait view: instead of one spinner line for up to
// an hour, an in-place block shows the progress bar, the elapsed time, and
// WHICH applications are still not ready (with their health), so a working
// install and a wedged one look different at a glance:
//
//	⠼ Installing OpenFrame applications · 12m34s
//	  ████████████░░░░░░░░  14/17 ready
//	    tenant     Progressing · Synced
//	    gateway    Degraded · Synced
//
// Active only on an interactive terminal, outside --silent and --verbose
// (verbose wants scrolling logs, which an in-place area would fight). All
// methods are nil-safe: a nil dashboard is the "not active" mode.
type appDashboard struct {
	mu       sync.Mutex
	area     *pterm.AreaPrinter
	start    time.Time
	stopped  bool
	frame   int
	readyAt map[string]time.Duration
	notes   []string
}

// dashboardMaxApps caps the not-ready list so a 40-app install still fits on
// a screen; the remainder is summarized as "+N more".
const dashboardMaxApps = 8

func newAppDashboard() *appDashboard {
	if !ui.IsTerminal() || ui.IsSilent() {
		return nil
	}
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return nil
	}
	return &appDashboard{
		area:    area,
		start:   time.Now(),
		readyAt: make(map[string]time.Duration),
	}
}

// Update redraws the dashboard from the current poll. total == 0 means the
// app-of-apps has not materialized its Applications yet.
func (d *appDashboard) Update(ready, total int, apps []Application) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	d.frame++
	elapsed := time.Since(d.start)
	for _, a := range apps {
		if a.Health == ArgoCDHealthHealthy && a.Sync == ArgoCDSyncSynced {
			if _, seen := d.readyAt[a.Name]; !seen {
				d.readyAt[a.Name] = elapsed
			}
		}
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	g := ui.Glyphs()
	var b strings.Builder
	fmt.Fprintf(&b, "%s Installing OpenFrame applications %s %s\n",
		pterm.FgCyan.Sprint(frames[d.frame%len(frames)]), g.Bullet, steps.FormatDuration(elapsed))
	if total == 0 {
		b.WriteString(pterm.FgGray.Sprint("  waiting for applications to be created by app-of-apps...") + "\n")
	} else {
		fmt.Fprintf(&b, "  %s  %d/%d ready\n", ui.ProgressBar(float64(ready)/float64(total), 24), ready, total)
		b.WriteString(notReadyLines(apps))
	}
	for _, n := range d.notes {
		b.WriteString(n + "\n")
	}
	d.area.Update(b.String())
}

// notReadyLines renders the not-ready applications (alphabetical, capped),
// colored by how bad their state is.
func notReadyLines(apps []Application) string {
	var pending []Application
	for _, a := range apps {
		if a.Health == ArgoCDHealthHealthy && a.Sync == ArgoCDSyncSynced {
			continue
		}
		pending = append(pending, a)
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Name < pending[j].Name })
	nameW := 0
	for _, a := range pending {
		if len(a.Name) > nameW {
			nameW = len(a.Name)
		}
	}
	g := ui.Glyphs()
	var b strings.Builder
	for i, a := range pending {
		if i >= dashboardMaxApps {
			fmt.Fprintf(&b, "    %s\n", pterm.FgGray.Sprintf("+%d more", len(pending)-dashboardMaxApps))
			break
		}
		state := fmt.Sprintf("%s %s %s", a.Health, g.Bullet, a.Sync)
		switch a.Health {
		case ArgoCDHealthDegraded:
			state = pterm.FgRed.Sprint(state)
		case ArgoCDHealthProgressing:
			state = pterm.FgYellow.Sprint(state)
		default:
			state = pterm.FgGray.Sprint(state)
		}
		fmt.Fprintf(&b, "    %-*s  %s\n", nameW, a.Name, state)
	}
	return b.String()
}

// Note pins a one-off event line (stall hint, recovery notice) under the
// dashboard so the area redraw cannot visually swallow it. Capped to the
// last five.
func (d *appDashboard) Note(styled string) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	d.notes = append(d.notes, styled)
	if len(d.notes) > 5 {
		d.notes = d.notes[len(d.notes)-5:]
	}
}

// Fail tears the dashboard down and prints the failure line in its place.
func (d *appDashboard) Fail(text string) { d.finish(text, false) }

// Stop tears the dashboard down without a final line.
func (d *appDashboard) Stop() { d.finish("", false) }

// FinishSuccess tears the dashboard down and prints the success line,
// appending which applications took longest — the number to bring to the
// team when the install feels slow.
func (d *appDashboard) FinishSuccess(text string) {
	if d == nil {
		return
	}
	if line := d.slowestLine(3); line != "" {
		text += " " + pterm.FgGray.Sprint(line)
	}
	d.finish(text, true)
}

func (d *appDashboard) finish(text string, ok bool) {
	if d == nil {
		return
	}
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.stopped = true
	d.mu.Unlock()
	_ = d.area.Stop()
	if text == "" {
		return
	}
	if ok {
		pterm.Success.Println(text)
		return
	}
	pterm.Error.Println(text)
}

// slowestLine renders "(slowest: tenant 9m12s · datasources 4m01s)".
func (d *appDashboard) slowestLine(n int) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	type rt struct {
		name string
		at   time.Duration
	}
	var all []rt
	for name, at := range d.readyAt {
		all = append(all, rt{name, at})
	}
	if len(all) == 0 {
		return ""
	}
	sort.Slice(all, func(i, j int) bool { return all[i].at > all[j].at })
	if n > len(all) {
		n = len(all)
	}
	g := ui.Glyphs()
	parts := make([]string, 0, n)
	for _, r := range all[:n] {
		parts = append(parts, fmt.Sprintf("%s %s", r.name, steps.FormatDuration(r.at)))
	}
	return "(slowest: " + strings.Join(parts, " "+g.Bullet+" ") + ")"
}
