package app

import (
	"testing"

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

func TestReadOnlyCommandsSkipPrereqGate(t *testing.T) {
	// status and access must be annotated read-only so PersistentPreRunE skips
	// the interactive prerequisite install gate (which could hang a script).
	for name, mk := range map[string]func() *cobra.Command{"status": getStatusCmd, "access": getAccessCmd} {
		if mk().Annotations["readonly"] != "true" {
			t.Errorf("%s command must be annotated readonly=true", name)
		}
	}
	// install is NOT read-only (it needs helm/kubectl).
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
