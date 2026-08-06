package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// KubeconfigHasContext must report what the kubeconfig actually holds — the
// list's CONTEXT column is derived from it, and used to fabricate a context
// for clusters whose create failed at plan time.
func TestKubeconfigHasContext(t *testing.T) {
	kubeconfig := filepath.Join(t.TempDir(), "config")
	require.NoError(t, os.WriteFile(kubeconfig, []byte(`
apiVersion: v1
kind: Config
contexts:
- name: my-eks
  context:
    cluster: my-eks
    user: my-eks
`), 0o600))
	t.Setenv("KUBECONFIG", kubeconfig)

	assert.True(t, KubeconfigHasContext("my-eks"))
	assert.False(t, KubeconfigHasContext("never-created"),
		"a context absent from the kubeconfig must not be reported")

	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "missing"))
	assert.False(t, KubeconfigHasContext("my-eks"),
		"an unreadable kubeconfig means no context, not an invented one")
}
