package app

import "testing"

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

func TestAppCommand_RegistersStatusAndAccess(t *testing.T) {
	app := GetAppCmd()
	want := map[string]bool{"install": false, "status": false, "access": false}
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
