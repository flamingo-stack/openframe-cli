package services

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/pterm/pterm"
)

// TestLoadExistingConfiguration_MissingFileIsHardErrorForUpgrade is the
// F3/T1-2 guard: an upgrade must REFUSE to run without
// openframe-helm-values.yaml. The old fallback to an empty values map made
// `helm upgrade` replace the release values with chart defaults — silently
// wiping registry credentials and ingress settings when the command ran from
// the wrong directory.
func TestLoadExistingConfiguration_MissingFileIsHardErrorForUpgrade(t *testing.T) {
	t.Chdir(t.TempDir()) // empty cwd: no openframe-helm-values.yaml

	w := &InstallationWorkflow{}
	_, err := w.loadExistingConfiguration(true)
	if err == nil {
		t.Fatal("missing values file must be a hard error when existing values are required (upgrade)")
	}
	if !strings.Contains(err.Error(), config.DefaultHelmValuesFile) {
		t.Errorf("error %q should name the missing file", err)
	}
}

// TestLoadExistingConfiguration_MissingFileAllowedForFreshInstall: fresh
// non-interactive install/bootstrap on a clean machine has no values file yet —
// chart defaults are a valid starting point (the contract command
// `bootstrap oss-tenant --non-interactive` must keep working), just announced
// with a warning instead of silently.
func TestLoadExistingConfiguration_MissingFileAllowedForFreshInstall(t *testing.T) {
	t.Chdir(t.TempDir())

	var infoBuf, warnBuf bytes.Buffer
	oldInfo, oldWarn := pterm.Info, pterm.Warning
	pterm.Info = *pterm.Info.WithWriter(&infoBuf)
	pterm.Warning = *pterm.Warning.WithWriter(&warnBuf)
	t.Cleanup(func() { pterm.Info, pterm.Warning = oldInfo, oldWarn })

	w := &InstallationWorkflow{}
	cfg, err := w.loadExistingConfiguration(false)
	if err != nil {
		t.Fatalf("fresh install without a values file must proceed with defaults: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cfg.TempHelmValuesPath) })
	if len(cfg.ExistingValues) != 0 {
		t.Errorf("expected empty values (chart defaults), got %#v", cfg.ExistingValues)
	}

	// N1 messaging guard: the missing file is announced as chart defaults, and
	// nothing may claim an existing file was used.
	if !strings.Contains(warnBuf.String(), "deploying chart defaults") {
		t.Errorf("missing values file must be announced loudly, got: %q", warnBuf.String())
	}
	if strings.Contains(infoBuf.String(), "Using existing") {
		t.Errorf("must not claim an existing values file was used, got: %q", infoBuf.String())
	}
}

// TestLoadExistingConfiguration_ExistingFileAnnounced: when the file IS there,
// say so (the counterpart of the N1 guard above).
func TestLoadExistingConfiguration_ExistingFileAnnounced(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(config.DefaultHelmValuesFile, []byte("repository:\n  branch: main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var infoBuf bytes.Buffer
	oldInfo := pterm.Info
	pterm.Info = *pterm.Info.WithWriter(&infoBuf)
	t.Cleanup(func() { pterm.Info = oldInfo })

	w := &InstallationWorkflow{}
	cfg, err := w.loadExistingConfiguration(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(cfg.TempHelmValuesPath) })

	if !strings.Contains(infoBuf.String(), "Using existing "+config.DefaultHelmValuesFile) {
		t.Errorf("existing values file must be announced, got: %q", infoBuf.String())
	}
}

// TestLoadExistingConfiguration_ExistingFileLoads: the happy path keeps working
// and the loaded values reach the configuration.
func TestLoadExistingConfiguration_ExistingFileLoads(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(config.DefaultHelmValuesFile, []byte("repository:\n  branch: develop\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	w := &InstallationWorkflow{}
	cfg, err := w.loadExistingConfiguration(true)
	if err != nil {
		t.Fatalf("loadExistingConfiguration: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cfg.TempHelmValuesPath) })

	repo, _ := cfg.ExistingValues["repository"].(map[string]interface{})
	if repo == nil || repo["branch"] != "develop" {
		t.Errorf("loaded values must carry the file content, got %#v", cfg.ExistingValues)
	}
}
