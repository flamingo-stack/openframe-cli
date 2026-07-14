package helm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

func installConfigWithValuesFile(path string) config.ChartInstallConfig {
	return config.ChartInstallConfig{
		ClusterName: "test",
		AppOfApps:   &models.AppOfAppsConfig{ValuesFile: path},
	}
}

// TestInstallArgoCDHelm_BrokenValuesFileFails locks the honesty fix: an
// existing-but-unparseable values file must fail the install instead of
// silently dropping the user's `argocd:` override and deploying the baseline.
func TestInstallArgoCDHelm_BrokenValuesFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	if err := os.WriteFile(path, []byte("argocd:\n\tbad: {yaml\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mock := executor.NewMockCommandExecutor()
	m, _ := NewHelmManager(mock, nil, false)

	result, err := m.installArgoCDHelm(context.Background(), installConfigWithValuesFile(path))
	if err == nil {
		t.Fatal("broken values file must fail the install")
	}
	if result != nil {
		t.Error("helm must not run when the values file is unparseable (nil result gates the diagnostics dump)")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error must name the values file, got: %v", err)
	}
	for _, c := range mock.Commands() {
		if c.Name == "helm" && len(c.Args) > 0 && c.Args[0] == "upgrade" {
			t.Fatal("helm upgrade must not be attempted with a broken values file")
		}
	}
}

// TestInstallArgoCDHelm_NonMappingArgoCDFails: the in-flow defense-in-depth
// check behind the pre-flight — a non-mapping `argocd:` fails before helm runs.
func TestInstallArgoCDHelm_NonMappingArgoCDFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	if err := os.WriteFile(path, []byte("argocd: [dex]\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mock := executor.NewMockCommandExecutor()
	m, _ := NewHelmManager(mock, nil, false)

	result, err := m.installArgoCDHelm(context.Background(), installConfigWithValuesFile(path))
	if err == nil {
		t.Fatal("non-mapping argocd override must fail the install")
	}
	if result != nil {
		t.Error("helm must not run on a merge error")
	}
}

// TestInstallArgoCDHelm_MissingValuesFileUsesBaseline: a missing user file
// stays normal — the install proceeds with the embedded baseline.
func TestInstallArgoCDHelm_MissingValuesFileUsesBaseline(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	m, _ := NewHelmManager(mock, nil, false)

	cfg := installConfigWithValuesFile(filepath.Join(t.TempDir(), "absent.yaml"))
	if _, err := m.installArgoCDHelm(context.Background(), cfg); err != nil {
		t.Fatalf("missing values file must not fail the install, got: %v", err)
	}
	up := findHelmUpgrade(t, mock.Commands())
	if len(up.Stdin) == 0 {
		t.Fatal("baseline values must still be piped to helm")
	}
}
