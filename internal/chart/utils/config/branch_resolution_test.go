package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeValues(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "helm-values.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

const bothBranches = `deployment:
  oss:
    repository:
      branch: oss-main
  saas:
    repository:
      branch: saas-main
`

func TestGetBranchForDeploymentMode(t *testing.T) {
	b := &Builder{}

	t.Run("saas-shared uses the saas branch", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode(writeValues(t, bothBranches), "saas-shared"); got != "saas-main" {
			t.Fatalf("got %q, want saas-main", got)
		}
	})

	t.Run("oss-tenant uses the oss branch", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode(writeValues(t, bothBranches), "oss-tenant"); got != "oss-main" {
			t.Fatalf("got %q, want oss-main", got)
		}
	})

	t.Run("saas-tenant uses the oss branch (app-of-apps lives in oss repo)", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode(writeValues(t, bothBranches), "saas-tenant"); got != "oss-main" {
			t.Fatalf("got %q, want oss-main", got)
		}
	})

	t.Run("missing file returns empty (use default)", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode("/nonexistent/helm-values.yaml", "oss-tenant"); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("malformed YAML returns empty", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode(writeValues(t, "not: [valid"), "oss-tenant"); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("no branch set returns empty", func(t *testing.T) {
		if got := b.getBranchForDeploymentMode(writeValues(t, "deployment:\n  oss:\n    enabled: true\n"), "oss-tenant"); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("saas-shared with only oss branch set returns empty (no fallback)", func(t *testing.T) {
		onlyOSS := "deployment:\n  oss:\n    repository:\n      branch: oss-main\n"
		if got := b.getBranchForDeploymentMode(writeValues(t, onlyOSS), "saas-shared"); got != "" {
			t.Fatalf("got %q, want empty (saas-shared must not fall back to oss)", got)
		}
	})
}
