package argocd

import (
	"reflect"
	"testing"
)

func TestAssessApplications(t *testing.T) {
	apps := []Application{
		{Name: "a", Health: "Healthy", Sync: "Synced"},
		{Name: "b", Health: "Healthy", Sync: "OutOfSync"},
		{Name: "c", Health: "Progressing", Sync: "Synced"},
		{Name: "d", Health: "Degraded", Sync: "OutOfSync"},
	}
	everReady := map[string]bool{}

	got := assessApplications(apps, everReady)

	if got.ready != 1 {
		t.Errorf("ready = %d, want 1", got.ready)
	}
	if !reflect.DeepEqual(got.healthyNames, []string{"a", "b"}) {
		t.Errorf("healthyNames = %v, want [a b]", got.healthyNames)
	}
	wantNotReady := []string{
		"b (Sync: OutOfSync)",
		"c (Health: Progressing)",
		"d (Degraded/OutOfSync)",
	}
	if !reflect.DeepEqual(got.notReady, wantNotReady) {
		t.Errorf("notReady = %v, want %v", got.notReady, wantNotReady)
	}
	if !everReady["a"] || len(everReady) != 1 {
		t.Errorf("everReady = %v, want only {a}", everReady)
	}
}

func TestAssessApplications_EverReadyIsSticky(t *testing.T) {
	everReady := map[string]bool{}

	// Tick 1: app is ready → marked.
	assessApplications([]Application{{Name: "a", Health: "Healthy", Sync: "Synced"}}, everReady)
	// Tick 2: same app went out of sync → must STAY marked.
	got := assessApplications([]Application{{Name: "a", Health: "Healthy", Sync: "OutOfSync"}}, everReady)

	if !everReady["a"] {
		t.Error("everReady must be sticky across ticks")
	}
	if got.ready != 0 {
		t.Errorf("ready = %d, want 0 (currently not ready)", got.ready)
	}
}

// TestIsDeploymentComplete locks the completion decision — the single most
// correctness-critical predicate of the install wait.
func TestIsDeploymentComplete(t *testing.T) {
	cases := []struct {
		name                       string
		total, ready, maxSeenTotal int
		want                       bool
	}{
		{"all ready at high-water mark", 27, 27, 27, true},
		{"all visible ready but count dropped below HWM", 20, 20, 27, false},
		{"not all ready", 27, 26, 27, false},
		{"zero apps never completes", 0, 0, 0, false},
		{"more than ever seen, all ready", 28, 28, 27, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isDeploymentComplete(c.total, c.ready, c.maxSeenTotal); got != c.want {
				t.Fatalf("isDeploymentComplete(%d,%d,%d) = %v, want %v", c.total, c.ready, c.maxSeenTotal, got, c.want)
			}
		})
	}
}

func TestClassifyAppIssues(t *testing.T) {
	counts := map[string]int{"gone": 3} // stale entry from a previous tick
	apps := []Application{
		{Name: "u", Health: "Unknown", Sync: "Synced"},
		{Name: "r", Health: "Progressing", Sync: "OutOfSync", Condition: "rpc error: EOF"},
		{Name: "m", Health: "Degraded", Sync: "OutOfSync", Condition: "failed to generate manifest for app"},
		{Name: "ok", Health: "Healthy", Sync: "Synced"},
		{Name: "recovered", Health: "Healthy", Sync: "Synced"}, // previously had issues
	}
	counts["recovered"] = 2

	unknown, condErrs := classifyAppIssues(apps, counts)

	if len(unknown) != 1 || unknown[0].Name != "u" {
		t.Errorf("unknown = %v, want [u]", names(unknown))
	}
	if len(condErrs) != 2 || condErrs[0].Name != "r" || condErrs[1].Name != "m" {
		t.Errorf("conditionErrors = %v, want [r m]", names(condErrs))
	}
	if counts["r"] != 1 || counts["m"] != 1 {
		t.Errorf("issue counts not incremented: %v", counts)
	}
	if _, ok := counts["recovered"]; ok {
		t.Error("recovered app's counter must be cleared")
	}
}

func TestClassifyAppIssues_CountsAccumulate(t *testing.T) {
	counts := map[string]int{}
	app := []Application{{Name: "r", Health: "Progressing", Sync: "OutOfSync", Condition: "server Unavailable"}}

	classifyAppIssues(app, counts)
	classifyAppIssues(app, counts)

	if counts["r"] != 2 {
		t.Fatalf("counts[r] = %d, want 2 (consecutive ticks accumulate)", counts["r"])
	}
}

func names(apps []Application) []string {
	out := make([]string, 0, len(apps))
	for _, a := range apps {
		out = append(out, a.Name)
	}
	return out
}
