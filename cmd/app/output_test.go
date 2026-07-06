package app

import (
	"encoding/json"
	"strings"
	"testing"

	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
)

func TestOutputFormat(t *testing.T) {
	cmd := getStatusCmd() // registers --output

	if f, err := outputFormat(cmd); err != nil || f != "text" {
		t.Fatalf("default = (%q, %v), want (text, nil)", f, err)
	}

	_ = cmd.Flags().Set("output", "json")
	if f, err := outputFormat(cmd); err != nil || f != "json" {
		t.Fatalf("json = (%q, %v), want (json, nil)", f, err)
	}

	_ = cmd.Flags().Set("output", "yaml")
	if f, err := outputFormat(cmd); err != nil || f != "yaml" {
		t.Fatalf("yaml = (%q, %v), want (yaml, nil)", f, err)
	}

	_ = cmd.Flags().Set("output", "toml")
	if _, err := outputFormat(cmd); err == nil {
		t.Fatal("expected an error for an unsupported --output value")
	}
}

// TestIsMachineOutput locks in that both json and yaml switch the command into
// machine mode (which suppresses the logo and the prerequisite gate), while
// text/default do not.
func TestIsMachineOutput(t *testing.T) {
	for _, tc := range []struct {
		format string
		want   bool
	}{
		{"", false},
		{"text", false},
		{"json", true},
		{"yaml", true},
	} {
		cmd := getStatusCmd()
		_ = cmd.Flags().Set("output", tc.format)
		if got := isMachineOutput(cmd); got != tc.want {
			t.Errorf("isMachineOutput(%q) = %v, want %v", tc.format, got, tc.want)
		}
	}
}

func TestStatusToJSON(t *testing.T) {
	rep := appstatus.Report{
		Health: k8s.Health{Reachable: true, NodesReady: 1, NodesTotal: 1},
		Apps: []argocd.Application{
			{Name: "a", Sync: "Synced", Health: "Healthy"},
			{Name: "b", Sync: "OutOfSync", Health: "Missing"},
		},
		Total: 2, Synced: 1, Healthy: 1,
		AdminPassword: "should-not-leak",
	}

	j := statusToJSON(rep)

	if !j.Reachable || j.NodesReady != 1 || j.NodesTotal != 1 {
		t.Fatalf("health mapping wrong: %+v", j)
	}
	if j.Ready {
		t.Fatal("not all apps synced+healthy → Ready must be false")
	}
	if j.Total != 2 || j.Synced != 1 || j.Healthy != 1 {
		t.Fatalf("counts wrong: %+v", j)
	}
	if len(j.Applications) != 2 || j.Applications[0].Name != "a" || j.Applications[1].Sync != "OutOfSync" {
		t.Fatalf("apps mapping wrong: %+v", j.Applications)
	}
	if j.Summary == "" {
		t.Fatal("summary should be populated")
	}
}

// TestStatusJSONHasNoPasswordField guards that the admin password is never
// emitted by `app status --output json` (that belongs to `app access`).
func TestStatusJSONHasNoPasswordField(t *testing.T) {
	b, err := json.Marshal(statusToJSON(appstatus.Report{AdminPassword: "secret"}))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "secret") || strings.Contains(strings.ToLower(s), "password") {
		t.Fatalf("status JSON must not contain a password:\n%s", s)
	}
}

func TestStatusAndAccessHaveOutputFlag(t *testing.T) {
	if getStatusCmd().Flags().Lookup("output") == nil {
		t.Error("status is missing --output")
	}
	if getAccessCmd().Flags().Lookup("output") == nil {
		t.Error("access is missing --output")
	}
}
