package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeValues(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestGetBranchFromHelmValuesPath locks branch resolution to the FLATTENED
// schema (top-level repository.branch) — the key SetRepositoryBranch writes and
// the chart consumes. The old test froze the nested legacy schema, which made
// the resolver read a key nothing else used (audit F1/T1-1: values-file branch
// silently ignored; stale legacy files overriding --ref).
func TestGetBranchFromHelmValuesPath(t *testing.T) {
	b := &Builder{}

	t.Run("flattened repository.branch is used", func(t *testing.T) {
		if got := b.getBranchFromHelmValuesPath(writeValues(t, "repository:\n  branch: oss-main\n")); got != "oss-main" {
			t.Fatalf("got %q, want oss-main", got)
		}
	})

	t.Run("legacy nested schema is ignored", func(t *testing.T) {
		// deployment.oss.repository.branch does nothing in the chart; honoring it
		// here is exactly the bug that let stale files override --ref.
		legacy := "deployment:\n  oss:\n    repository:\n      branch: stale-legacy\n"
		if got := b.getBranchFromHelmValuesPath(writeValues(t, legacy)); got != "" {
			t.Fatalf("legacy schema must be ignored, got %q", got)
		}
	})

	t.Run("missing file returns empty (use default)", func(t *testing.T) {
		if got := b.getBranchFromHelmValuesPath("/nonexistent/openframe-helm-values.yaml"); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("malformed YAML returns empty", func(t *testing.T) {
		if got := b.getBranchFromHelmValuesPath(writeValues(t, "not: [valid")); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("no branch set returns empty", func(t *testing.T) {
		if got := b.getBranchFromHelmValuesPath(writeValues(t, "repository:\n  url: something\n")); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}

// TestBuildInstallConfig_BranchFromValuesFile is the end-to-end F1 guard: a
// branch in the (flattened) values file must reach AppOfApps.GitHubBranch, so
// the app-of-apps clone and the children's targetRevision agree.
func TestBuildInstallConfig_BranchFromValuesFile(t *testing.T) {
	b := NewBuilder(nil)

	path := writeValues(t, "repository:\n  branch: develop\n")
	cfg, err := b.BuildInstallConfigWithCustomHelmPath(
		false, false, false, true,
		"test-cluster", "https://github.com/flamingo-stack/openframe-oss-tenant", "main", "", path,
	)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppOfApps == nil {
		t.Fatal("AppOfApps config expected")
	}
	if cfg.AppOfApps.GitHubBranch != "develop" {
		t.Fatalf("values-file branch must win over the flag default: got %q, want develop", cfg.AppOfApps.GitHubBranch)
	}
}
