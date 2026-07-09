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

const ossBranchValues = `deployment:
  oss:
    repository:
      branch: oss-main
`

// TestGetBranchFromHelmValuesPath covers OSS branch resolution. The CLI supports
// only the OSS (oss-tenant) deployment, so the branch is always read from
// deployment.oss.repository.branch.
func TestGetBranchFromHelmValuesPath(t *testing.T) {
	b := &Builder{}

	t.Run("oss branch is used", func(t *testing.T) {
		if got := b.getBranchFromHelmValuesPath(writeValues(t, ossBranchValues)); got != "oss-main" {
			t.Fatalf("got %q, want oss-main", got)
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
		if got := b.getBranchFromHelmValuesPath(writeValues(t, "deployment:\n  oss:\n    enabled: true\n")); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}
