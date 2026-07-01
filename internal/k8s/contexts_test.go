package k8s

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleKubeconfig = `apiVersion: v1
kind: Config
current-context: ctx-b
contexts:
- name: ctx-a
  context:
    cluster: cluster-a
    user: user-a
- name: ctx-b
  context:
    cluster: cluster-b
    user: user-b
clusters:
- name: cluster-a
  cluster:
    server: https://a.example
- name: cluster-b
  cluster:
    server: https://b.example
users:
- name: user-a
- name: user-b
`

func writeKubeconfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadContexts(t *testing.T) {
	path := writeKubeconfig(t, sampleKubeconfig)

	contexts, current, err := LoadContexts(path)
	require.NoError(t, err)
	assert.Equal(t, "ctx-b", current)
	require.Len(t, contexts, 2)

	// sorted by name
	assert.Equal(t, "ctx-a", contexts[0].Name)
	assert.Equal(t, "cluster-a", contexts[0].Cluster)
	assert.False(t, contexts[0].Current)

	assert.Equal(t, "ctx-b", contexts[1].Name)
	assert.True(t, contexts[1].Current, "current context must be flagged")
}

func TestLoadContexts_MissingFile(t *testing.T) {
	_, _, err := LoadContexts(filepath.Join(t.TempDir(), "nope"))
	require.Error(t, err)
}

func TestDefaultKubeconfigPath_EnvWins(t *testing.T) {
	t.Setenv("KUBECONFIG", "/custom/kubeconfig")
	assert.Equal(t, "/custom/kubeconfig", DefaultKubeconfigPath())
}
