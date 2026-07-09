package services

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDryRunConfiguration_WritesRealValuesFile locks the fix for the dry-run
// anomaly: TempHelmValuesPath must point at a file that actually exists on disk.
// Previously dry-run set a fixed "helm-values-tmp.yaml" that nothing ever wrote,
// so `helm --dry-run -f helm-values-tmp.yaml` ran against a missing file.
func TestDryRunConfiguration_WritesRealValuesFile(t *testing.T) {
	w := &InstallationWorkflow{}

	cfg, err := w.dryRunConfiguration()
	if err != nil {
		t.Fatalf("dryRunConfiguration: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cfg.TempHelmValuesPath) })

	// The core of the bug: the values file must exist.
	if _, err := os.Stat(cfg.TempHelmValuesPath); err != nil {
		t.Fatalf("dry-run values file must exist on disk, got stat error: %v", err)
	}

	// And it must not be the old phantom fixed name in the working directory.
	if filepath.Base(cfg.TempHelmValuesPath) == "helm-values-tmp.yaml" {
		t.Errorf("must not use the phantom fixed name: %q", cfg.TempHelmValuesPath)
	}
	if dir := filepath.Dir(cfg.TempHelmValuesPath); dir == "." || dir == "" {
		t.Errorf("dry-run values file must be an absolute temp path, got %q", cfg.TempHelmValuesPath)
	}
}
