// Package tui is the interactive platform view (`app status --interactive`):
// a small k9s-style screen over the ArgoCD applications — arrows navigate,
// enter opens the app's detail (revision, conditions, operation state),
// s triggers a sync of the selected app, r refreshes, q quits. Read paths
// reuse the same status service as the plain command; the only write action
// is the explicit per-app sync.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// Syncer triggers a sync of the named applications.
type Syncer func(ctx context.Context, names []string) error

// Run starts the interactive status view and blocks until the user quits.
func Run(ctx context.Context, svc *appstatus.Service, sync Syncer, verbose bool) error {
	m := model{ctx: ctx, svc: svc, sync: sync, verbose: verbose}
	_, err := tea.NewProgram(&m, tea.WithContext(ctx)).Run()
	if err != nil && ctx.Err() != nil {
		return nil // Ctrl+C via the signal context is a clean exit
	}
	return err
}

const refreshEvery = 3 * time.Second

type refreshMsg struct {
	rep appstatus.Report
	err error
}

type syncDoneMsg struct {
	name string
	err  error
}

type tickMsg struct{}

type model struct {
	ctx     context.Context
	svc     *appstatus.Service
	sync    Syncer
	verbose bool

	rep     appstatus.Report
	loaded  bool
	loadErr error

	cursor int
	detail bool
	note   string
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.refresh(), tick())
}

func tick() tea.Cmd {
	return tea.Tick(refreshEvery, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *model) refresh() tea.Cmd {
	return func() tea.Msg {
		rep, err := m.svc.Report(m.ctx, m.verbose)
		return refreshMsg{rep: rep, err: err}
	}
}

func (m *model) selected() *argocd.Application {
	if !m.loaded || m.cursor < 0 || m.cursor >= len(m.rep.Apps) {
		return nil
	}
	return &m.rep.Apps[m.cursor]
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(m.refresh(), tick())
	case refreshMsg:
		m.loaded, m.loadErr = true, msg.err
		if msg.err == nil {
			sort.Slice(msg.rep.Apps, func(i, j int) bool { return msg.rep.Apps[i].Name < msg.rep.Apps[j].Name })
			m.rep = msg.rep
			if m.cursor >= len(m.rep.Apps) {
				m.cursor = len(m.rep.Apps) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
		return m, nil
	case syncDoneMsg:
		if msg.err != nil {
			m.note = pterm.FgRed.Sprintf("sync %s: %v", msg.name, msg.err)
		} else {
			m.note = pterm.FgGreen.Sprintf("sync triggered for %s", msg.name)
		}
		return m, m.refresh()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			if m.detail && msg.String() == "esc" {
				m.detail = false
				return m, nil
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rep.Apps)-1 {
				m.cursor++
			}
		case "enter":
			m.detail = !m.detail
		case "r":
			m.note = ""
			return m, m.refresh()
		case "s":
			if app := m.selected(); app != nil && m.sync != nil {
				name := app.Name
				m.note = pterm.FgYellow.Sprintf("syncing %s...", name)
				return m, func() tea.Msg { return syncDoneMsg{name: name, err: m.sync(m.ctx, []string{name})} }
			}
		}
	}
	return m, nil
}

func (m *model) View() string {
	g := ui.Glyphs()
	var b strings.Builder
	fmt.Fprintf(&b, "%s OpenFrame %s %s\n\n",
		pterm.FgCyan.Sprint(g.Running), g.Bullet, statusLine(m))
	if !m.loaded {
		b.WriteString("  loading...\n")
		return b.String()
	}
	if m.loadErr != nil {
		fmt.Fprintf(&b, "  %s %v\n", pterm.FgRed.Sprint(g.Fail), m.loadErr)
	}
	if m.detail {
		b.WriteString(detailView(m.selected()))
	} else {
		b.WriteString(listView(m))
	}
	if m.note != "" {
		fmt.Fprintf(&b, "\n%s\n", m.note)
	}
	b.WriteString(pterm.FgGray.Sprint("\n↑/↓ navigate · enter detail · s sync · r refresh · q quit\n"))
	return b.String()
}

func statusLine(m *model) string {
	if !m.loaded {
		return "connecting..."
	}
	return m.rep.Summary()
}

func listView(m *model) string {
	g := ui.Glyphs()
	var b strings.Builder
	nameW := 4
	for _, a := range m.rep.Apps {
		if len(a.Name) > nameW {
			nameW = len(a.Name)
		}
	}
	fmt.Fprintf(&b, "  %-*s  %-12s %s\n", nameW, "NAME", "SYNC", "HEALTH")
	for i, a := range m.rep.Apps {
		marker := "  "
		line := fmt.Sprintf("%-*s  %-12s %s", nameW, a.Name, a.Sync, a.Health)
		if i == m.cursor {
			marker = pterm.FgCyan.Sprint(g.Arrow) + " "
			line = pterm.Bold.Sprint(line)
		}
		b.WriteString(marker + colorByHealth(a.Health, line) + "\n")
	}
	if len(m.rep.Apps) == 0 {
		b.WriteString("  no OpenFrame applications found — is it installed?\n")
	}
	return b.String()
}

func detailView(app *argocd.Application) string {
	if app == nil {
		return "  nothing selected\n"
	}
	row := func(k, v string) string {
		if v == "" {
			return ""
		}
		return fmt.Sprintf("  %s %s\n", pterm.FgGray.Sprintf("%-12s", k), v)
	}
	var b strings.Builder
	b.WriteString(pterm.Bold.Sprint("  "+app.Name) + "\n\n")
	b.WriteString(row("health", colorByHealth(app.Health, app.Health)))
	b.WriteString(row("sync", app.Sync))
	b.WriteString(row("namespace", app.Namespace))
	b.WriteString(row("repo", app.RepoURL))
	b.WriteString(row("path", app.Path))
	b.WriteString(row("target", app.TargetRevision))
	b.WriteString(row("revision", app.SyncRevision))
	b.WriteString(row("reconciled", app.ReconciledAt))
	b.WriteString(row("operation", strings.TrimSpace(app.OperationPhase+" "+app.OperationMessage)))
	if app.Condition != "" {
		b.WriteString(row("condition", fmt.Sprintf("[%s] %s", app.ConditionType, app.Condition)))
	}
	b.WriteString(row("message", app.HealthMessage))
	return b.String()
}

func colorByHealth(health, text string) string {
	switch health {
	case argocd.ArgoCDHealthHealthy:
		return pterm.FgGreen.Sprint(text)
	case argocd.ArgoCDHealthDegraded:
		return pterm.FgRed.Sprint(text)
	case argocd.ArgoCDHealthProgressing:
		return pterm.FgYellow.Sprint(text)
	default:
		return text
	}
}
