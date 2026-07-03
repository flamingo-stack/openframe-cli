package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateTemporaryValuesFile_UniqueAndPrivate verifies the temp values file
// (which can hold registry/repository secrets) lands in the OS temp dir — not
// the user's working directory — with a unique name and 0600 perms, so
// concurrent runs don't clobber each other and a pre-created file can't
// redirect the write.
func TestCreateTemporaryValuesFile_UniqueAndPrivate(t *testing.T) {
	h := &HelmValuesModifier{}
	values := map[string]interface{}{"deployment": map[string]interface{}{"saas": map[string]interface{}{"repository": map[string]interface{}{"password": "s3cret"}}}}

	p1, err := h.CreateTemporaryValuesFile(values)
	if err != nil {
		t.Fatalf("CreateTemporaryValuesFile: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(p1) })
	p2, err := h.CreateTemporaryValuesFile(values)
	if err != nil {
		t.Fatalf("CreateTemporaryValuesFile (2nd): %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(p2) })

	// Lands in the OS temp dir, NOT the current working directory.
	if filepath.Dir(p1) != strings.TrimRight(os.TempDir(), string(os.PathSeparator)) {
		t.Errorf("temp file must be in os.TempDir() (%s), got dir %s", os.TempDir(), filepath.Dir(p1))
	}
	cwd, _ := os.Getwd()
	if filepath.Dir(p1) == cwd {
		t.Errorf("temp file must not be created in the working directory: %q", p1)
	}

	// Unique names, not the old fixed filename.
	if p1 == p2 {
		t.Fatalf("expected unique temp names, both = %q", p1)
	}
	if filepath.Base(p1) == "helm-values-tmp.yaml" {
		t.Fatalf("temp file should not use the fixed name: %q", p1)
	}
	if !strings.HasPrefix(filepath.Base(p1), "helm-values-tmp-") || !strings.HasSuffix(p1, ".yaml") {
		t.Fatalf("unexpected temp name pattern: %q", p1)
	}

	// 0600 perms (secret-bearing).
	info, err := os.Stat(p1)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perms = %v, want 0600", info.Mode().Perm())
	}

	// Content actually written.
	b, err := os.ReadFile(p1) // #nosec G304 -- reads a path this test just created
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "s3cret") {
		t.Fatalf("values not written to temp file:\n%s", b)
	}
}
