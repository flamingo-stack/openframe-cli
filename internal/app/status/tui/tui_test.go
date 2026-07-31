package tui

import (
	"context"
	"strings"
	"testing"

	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	tea "github.com/charmbracelet/bubbletea"
)

func loadedModel(apps ...argocd.Application) *model {
	m := &model{ctx: context.Background(), loaded: true}
	m.rep = appstatus.Report{Apps: apps, Total: len(apps)}
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestModel_NavigationClamps(t *testing.T) {
	m := loadedModel(
		argocd.Application{Name: "a"},
		argocd.Application{Name: "b"},
	)
	m.Update(key("up"))
	if m.cursor != 0 {
		t.Fatal("up at the top must clamp")
	}
	m.Update(key("down"))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	m.Update(key("down"))
	if m.cursor != 1 {
		t.Fatal("down at the bottom must clamp")
	}
}

func TestModel_DetailToggleAndEsc(t *testing.T) {
	m := loadedModel(argocd.Application{Name: "tenant", Health: argocd.ArgoCDHealthDegraded, Namespace: "tenant"})
	m.Update(key("enter"))
	if !m.detail {
		t.Fatal("enter must open the detail view")
	}
	view := m.View()
	if !strings.Contains(view, "tenant") || !strings.Contains(view, "namespace") {
		t.Fatalf("detail view lacks fields:\n%s", view)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.detail {
		t.Fatal("esc must close the detail view, not quit")
	}
}

func TestModel_SyncUsesSelectedApp(t *testing.T) {
	var synced []string
	m := loadedModel(argocd.Application{Name: "a"}, argocd.Application{Name: "b"})
	m.sync = func(_ context.Context, names []string) error {
		synced = append(synced, names...)
		return nil
	}
	m.cursor = 1
	_, cmd := m.Update(key("s"))
	if cmd == nil {
		t.Fatal("s must produce a sync command")
	}
	if msg, ok := cmd().(syncDoneMsg); !ok || msg.name != "b" {
		t.Fatalf("sync must target the selected app, got %+v", msg)
	}
	if len(synced) != 1 || synced[0] != "b" {
		t.Fatalf("synced = %v", synced)
	}
}

func TestModel_RefreshClampsCursorWhenAppsShrink(t *testing.T) {
	m := loadedModel(argocd.Application{Name: "a"}, argocd.Application{Name: "b"}, argocd.Application{Name: "c"})
	m.cursor = 2
	m.Update(refreshMsg{rep: appstatus.Report{Apps: []argocd.Application{{Name: "only"}}, Total: 1}})
	if m.cursor != 0 {
		t.Fatalf("cursor must clamp into the new list, got %d", m.cursor)
	}
}

func TestView_ListShowsAppsAndHelp(t *testing.T) {
	m := loadedModel(argocd.Application{Name: "gateway", Sync: "Synced", Health: argocd.ArgoCDHealthHealthy})
	view := m.View()
	for _, want := range []string{"gateway", "SYNC", "HEALTH", "q quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view lacks %q:\n%s", want, view)
		}
	}
}
