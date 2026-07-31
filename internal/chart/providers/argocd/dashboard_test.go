package argocd

import (
	"strings"
	"testing"
	"time"
)

// A nil dashboard is the "not active" mode — every method must be a no-op,
// because the wait loop calls them unconditionally.
func TestAppDashboard_NilSafe(t *testing.T) {
	var d *appDashboard
	d.Update(1, 2, []Application{{Name: "x"}})
	d.Note("note")
	d.Fail("fail")
	d.FinishSuccess("ok")
	d.Stop()
}

func TestNotReadyLines(t *testing.T) {
	apps := []Application{
		{Name: "ready", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced},
		{Name: "tenant", Health: ArgoCDHealthProgressing, Sync: ArgoCDSyncSynced},
		{Name: "gateway", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncSynced},
	}
	out := notReadyLines(apps)
	if strings.Contains(out, "ready") {
		t.Fatalf("ready apps must not be listed:\n%s", out)
	}
	if !strings.Contains(out, "tenant") || !strings.Contains(out, "gateway") {
		t.Fatalf("not-ready apps must be listed:\n%s", out)
	}
	// Alphabetical: gateway before tenant.
	if strings.Index(out, "gateway") > strings.Index(out, "tenant") {
		t.Fatalf("expected alphabetical order:\n%s", out)
	}
}

func TestNotReadyLines_CapsWithMore(t *testing.T) {
	var apps []Application
	for i := 0; i < dashboardMaxApps+3; i++ {
		apps = append(apps, Application{Name: strings.Repeat("a", i+1), Health: ArgoCDHealthProgressing})
	}
	out := notReadyLines(apps)
	if !strings.Contains(out, "+3 more") {
		t.Fatalf("overflow must be summarized:\n%s", out)
	}
}

func TestAppDashboard_SlowestLine(t *testing.T) {
	d := &appDashboard{readyAt: map[string]time.Duration{
		"fast":   30 * time.Second,
		"slow":   9 * time.Minute,
		"medium": 4 * time.Minute,
	}}
	line := d.slowestLine(2)
	if !strings.Contains(line, "slow 9m00s") || !strings.Contains(line, "medium 4m00s") {
		t.Fatalf("line = %q", line)
	}
	if strings.Contains(line, "fast") {
		t.Fatalf("only the N slowest belong in the line: %q", line)
	}

	empty := &appDashboard{readyAt: map[string]time.Duration{}}
	if empty.slowestLine(3) != "" {
		t.Fatal("no ready apps → empty line")
	}
}
