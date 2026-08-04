package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

// CopyPlanInputs feeds a preview plan of an existing workspace: state and
// backend travel to the throwaway dir, the (stale) module does not, and a
// state-less workspace is fine.
func TestCopyPlanInputs(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("terraform.tfstate", `{"version":4}`)
	write("backend.tf", `terraform { backend "gcs" {} }`)
	write("main.tf", "# stale module — must NOT be copied")

	if err := CopyPlanInputs(src, dst); err != nil {
		t.Fatalf("CopyPlanInputs: %v", err)
	}

	state, err := os.ReadFile(filepath.Join(dst, "terraform.tfstate"))
	if err != nil || string(state) != `{"version":4}` {
		t.Fatalf("state not copied: %s, %v", state, err)
	}
	backend, err := os.ReadFile(filepath.Join(dst, "backend.tf"))
	if err != nil || string(backend) == "" {
		t.Fatalf("backend not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "main.tf")); err == nil {
		t.Fatal("the stale module must not be copied — the preview regenerates it")
	}
}

func TestCopyPlanInputs_MissingFilesAreFine(t *testing.T) {
	if err := CopyPlanInputs(t.TempDir(), t.TempDir()); err != nil {
		t.Fatalf("a workspace without state/backend must copy nothing and succeed, got %v", err)
	}
}
