package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteValues_SecretFileIsOwnerOnly is the I2 regression guard: a helm
// values file may carry secrets (SaaS PAT, docker password), so it must be
// written 0600, never world-readable.
func TestWriteValues_SecretFileIsOwnerOnly(t *testing.T) {
	h := NewHelmValuesModifier()
	dir := t.TempDir()
	path := filepath.Join(dir, "helm-values.yaml")

	values := map[string]interface{}{
		"saas": map[string]interface{}{
			"repository": map[string]interface{}{
				"password": "ghp_secretValue", //nolint:gosec // test fixture
			},
		},
	}
	require.NoError(t, h.WriteValues(values, path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
		"helm values file with secrets must be 0600, got %v", info.Mode().Perm())
}
