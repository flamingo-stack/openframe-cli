package ui

import (
	"strings"
	"testing"
)

// TestRequireConfirmation_NonInteractiveFailsFast is the B3 contract guard:
// when no one can answer a prompt (CI, piped stdin), a required confirmation
// must fail fast with the skip-flag hint — never block, and never silently
// proceed with a destructive default.
func TestRequireConfirmation_NonInteractiveFailsFast(t *testing.T) {
	t.Setenv("CI", "1") // force IsNonInteractive() == true deterministically

	ok, err := RequireConfirmation("Remove everything?", "--yes", false)
	if err == nil {
		t.Fatal("RequireConfirmation must error in a non-interactive session")
	}
	if ok {
		t.Error("RequireConfirmation must never report confirmed on the fail-fast path")
	}
	for _, want := range []string{"--yes", "non-interactive"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

func TestIsNonInteractive_CIEnv(t *testing.T) {
	t.Setenv("CI", "1")
	if !IsNonInteractive() {
		t.Error("CI env must force non-interactive mode")
	}
}
