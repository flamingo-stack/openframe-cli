package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestConfigForContext_ExplicitContext(t *testing.T) {
	path := writeKubeconfig(t, sampleKubeconfig)

	a, err := RestConfigForContext(path, "ctx-a")
	require.NoError(t, err)
	assert.Equal(t, "https://a.example", a.Host)

	b, err := RestConfigForContext(path, "ctx-b")
	require.NoError(t, err)
	assert.Equal(t, "https://b.example", b.Host)
}

func TestRestConfigForContext_EmptyUsesCurrent(t *testing.T) {
	path := writeKubeconfig(t, sampleKubeconfig)
	// current-context in the sample is ctx-b
	cfg, err := RestConfigForContext(path, "")
	require.NoError(t, err)
	assert.Equal(t, "https://b.example", cfg.Host)
}

func TestRestConfigForContext_UnknownContext(t *testing.T) {
	path := writeKubeconfig(t, sampleKubeconfig)
	_, err := RestConfigForContext(path, "does-not-exist")
	require.Error(t, err)
}

func TestRestConfigForContext_KubeconfigList(t *testing.T) {
	list, _, _ := writeKubeconfigList(t)

	// A context living only in the second file of a $KUBECONFIG list must
	// resolve — the raw list is not a single file path.
	cfg, err := RestConfigForContext(list, "two-ctx")
	require.NoError(t, err)
	assert.Equal(t, "https://two.example", cfg.Host)
}
