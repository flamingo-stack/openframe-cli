package app

import (
	"fmt"
	"strings"
	"testing"

	appstatus "github.com/flamingo-stack/openframe-cli/internal/app/status"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/spf13/cobra"
)

func TestStatusCommand_Wiring(t *testing.T) {
	cmd := getStatusCmd()
	if cmd.Use != "status" {
		t.Fatalf("Use = %q, want status", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatal("status command has no RunE")
	}
	if cmd.Flags().Lookup("context") == nil {
		t.Fatal("status command is missing the --context flag")
	}
}

func TestAccessCommand_Wiring(t *testing.T) {
	cmd := getAccessCmd()
	if cmd.Use != "access" {
		t.Fatalf("Use = %q, want access", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatal("access command has no RunE")
	}
	if cmd.Flags().Lookup("context") == nil {
		t.Fatal("access command is missing the --context flag")
	}
}

func TestUninstallCommand_Wiring(t *testing.T) {
	cmd := getUninstallCmd()
	if cmd.Use != "uninstall" {
		t.Fatalf("Use = %q, want uninstall", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Fatal("uninstall command has no RunE")
	}
	for _, f := range []string{"context", "yes", "delete-namespace"} {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("uninstall command is missing the --%s flag", f)
		}
	}
	if yes := cmd.Flags().Lookup("yes"); yes == nil || yes.Shorthand != "y" {
		t.Error("--yes should have the -y shorthand")
	}
}

func TestReadOnlyCommandsAreAnnotated(t *testing.T) {
	// status and access only read an existing cluster, so they carry the
	// read-only annotation; install/upgrade/uninstall mutate it and do not.
	for name, mk := range map[string]func() *cobra.Command{"status": getStatusCmd, "access": getAccessCmd} {
		if mk().Annotations["readonly"] != "true" {
			t.Errorf("%s command must be annotated readonly=true", name)
		}
	}
	// install mutates the cluster (installs ArgoCD + apps).
	if getInstallCmd().Annotations["readonly"] == "true" {
		t.Error("install must not be marked read-only")
	}
}

func TestInstallCommandHasContextFlag(t *testing.T) {
	if getInstallCmd().Flags().Lookup("context") == nil {
		t.Fatal("install is missing the --context flag")
	}
}

func TestAppCommand_RegistersStatusAndAccess(t *testing.T) {
	app := GetAppCmd()
	want := map[string]bool{"install": false, "status": false, "access": false, "uninstall": false}
	for _, sub := range app.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("app command does not register %q subcommand", name)
		}
	}
}

// reachable=false next to a populated applications array was self-contradictory
// machine output; healthError now says why (e.g. an RBAC-denied node read).
func TestStatusToJSON_CarriesHealthError(t *testing.T) {
	rep := appstatus.Report{HealthErr: fmt.Errorf("nodes is forbidden")}
	j := statusToJSON(rep)
	if j.HealthError != "nodes is forbidden" {
		t.Fatalf("HealthError = %q", j.HealthError)
	}
	if statusToJSON(appstatus.Report{}).HealthError != "" {
		t.Fatal("no health error → field stays empty (omitted from JSON)")
	}
}

// --watch is a live terminal view: machine output must be rejected up front,
// and each frame must carry the data a human polls for.
func TestRenderWatchFrame(t *testing.T) {
	rep := appstatus.Report{
		Health: k8s.Health{Reachable: true, NodesReady: 3, NodesTotal: 3},
		Apps: []argocd.Application{
			{Name: "tenant", Sync: "Synced", Health: "Progressing"},
		},
		Total: 1,
	}
	frame := renderWatchFrame(rep, nil)
	for _, want := range []string{"OpenFrame status", "3/3 nodes ready", "tenant", "Progressing", "Ctrl+C"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("frame lacks %q:\n%s", want, frame)
		}
	}

	errFrame := renderWatchFrame(appstatus.Report{}, fmt.Errorf("boom"))
	if !strings.Contains(errFrame, "boom") {
		t.Fatalf("a failing poll must show its error and keep watching:\n%s", errFrame)
	}

	empty := renderWatchFrame(appstatus.Report{Health: k8s.Health{Reachable: true}}, nil)
	if !strings.Contains(empty, "is it installed") {
		t.Fatalf("no apps → install hint:\n%s", empty)
	}
}

func TestStatusCommand_WatchRejectsMachineOutput(t *testing.T) {
	cmd := getStatusCmd()
	cmd.SetArgs([]string{"--watch", "--output", "json"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--watch") {
		t.Fatalf("expected the watch/output conflict error, got %v", err)
	}
}
