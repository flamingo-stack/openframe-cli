package k8s

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
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

func TestResolveContextForCluster(t *testing.T) {
	kubeconfig := `apiVersion: v1
kind: Config
current-context: prod
clusters:
- {name: c1, cluster: {server: https://x}}
contexts:
- {name: prod, context: {cluster: c1, user: u}}
- {name: k3d-dev, context: {cluster: c1, user: u}}
users:
- {name: u, user: {}}
`
	path := writeKubeconfig(t, kubeconfig)

	// Empty cluster name → empty context.
	assert.Equal(t, "", ResolveContextForCluster(path, ""))

	// An exact context match wins (non-k3d / renamed context).
	assert.Equal(t, "prod", ResolveContextForCluster(path, "prod"))

	// No literal match → k3d-<name> convention (which happens to exist here).
	assert.Equal(t, "k3d-dev", ResolveContextForCluster(path, "dev"))

	// No match at all → k3d-<name> fallback (preserves prior behavior).
	assert.Equal(t, "k3d-missing", ResolveContextForCluster(path, "missing"))

	// Unreadable kubeconfig → k3d-<name> fallback, never empty for a named cluster.
	assert.Equal(t, "k3d-foo", ResolveContextForCluster(filepath.Join(t.TempDir(), "nope"), "foo"))
}

func TestDefaultKubeconfigPath_EnvWins(t *testing.T) {
	t.Setenv("KUBECONFIG", "/custom/kubeconfig")
	assert.Equal(t, "/custom/kubeconfig", DefaultKubeconfigPath())
}

func TestSwitchContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kubeconfig")
	kubeconfig := `apiVersion: v1
kind: Config
current-context: alpha
clusters:
- name: c1
  cluster:
    server: https://a.example
contexts:
- name: alpha
  context:
    cluster: c1
    user: u
- name: beta
  context:
    cluster: c1
    user: u
users:
- name: u
`
	if err := os.WriteFile(path, []byte(kubeconfig), 0o600); err != nil {
		t.Fatal(err)
	}

	if !HasContext(path, "beta") || HasContext(path, "ghost") {
		t.Fatal("HasContext misreports")
	}

	if err := SwitchContext(path, "beta"); err != nil {
		t.Fatalf("SwitchContext: %v", err)
	}
	_, current, err := LoadContexts(path)
	if err != nil || current != "beta" {
		t.Fatalf("current-context = %q (err %v), want beta", current, err)
	}

	// Switching only flips the pointer — both contexts must survive.
	contexts, _, _ := LoadContexts(path)
	if len(contexts) != 2 {
		t.Fatalf("contexts damaged: %v", contexts)
	}

	if err := SwitchContext(path, "ghost"); err == nil {
		t.Fatal("switching to a missing context must fail")
	}
}

// $KUBECONFIG may be a path-list of files (kubectl convention). These tests
// pin the merged handling: two files joined by the OS path-list separator.

const kubeconfigFileOne = `apiVersion: v1
kind: Config
current-context: one-ctx
clusters:
- {name: c1, cluster: {server: https://one.example}}
contexts:
- {name: one-ctx, context: {cluster: c1, user: u}}
users:
- {name: u, user: {}}
`

const kubeconfigFileTwo = `apiVersion: v1
kind: Config
clusters:
- {name: c2, cluster: {server: https://two.example}}
contexts:
- {name: two-ctx, context: {cluster: c2, user: u2}}
users:
- {name: u2, user: {}}
`

// writeKubeconfigList writes both files into one temp dir and returns their
// paths joined by the OS path-list separator, plus the individual paths.
func writeKubeconfigList(t *testing.T) (list, file1, file2 string) {
	t.Helper()
	dir := t.TempDir()
	file1 = filepath.Join(dir, "config1")
	file2 = filepath.Join(dir, "config2")
	require.NoError(t, os.WriteFile(file1, []byte(kubeconfigFileOne), 0o600))
	require.NoError(t, os.WriteFile(file2, []byte(kubeconfigFileTwo), 0o600))
	return file1 + string(os.PathListSeparator) + file2, file1, file2
}

func TestLoadContexts_KubeconfigList(t *testing.T) {
	list, _, _ := writeKubeconfigList(t)

	contexts, current, err := LoadContexts(list)
	require.NoError(t, err)
	assert.Equal(t, "one-ctx", current, "current-context comes from the file that sets it")
	require.Len(t, contexts, 2, "contexts from BOTH files must be merged")
	assert.Equal(t, "one-ctx", contexts[0].Name)
	assert.Equal(t, "two-ctx", contexts[1].Name)
	assert.Equal(t, "c2", contexts[1].Cluster)
}

func TestResolveContextForCluster_KubeconfigList(t *testing.T) {
	list, _, _ := writeKubeconfigList(t)

	// A context living only in the SECOND file must be found by exact match
	// instead of falling back to the k3d convention.
	assert.Equal(t, "two-ctx", ResolveContextForCluster(list, "two-ctx"))
}

func TestSwitchContext_KubeconfigList(t *testing.T) {
	list, file1, file2 := writeKubeconfigList(t)

	// A context defined only in file 2 must be switchable; kubectl writes the
	// pointer into the first file that already sets current-context (file 1).
	require.NoError(t, SwitchContext(list, "two-ctx"))

	cfg1, err := clientcmd.LoadFromFile(file1)
	require.NoError(t, err)
	assert.Equal(t, "two-ctx", cfg1.CurrentContext, "pointer belongs in the file that owned it")
	require.Len(t, cfg1.Contexts, 1, "file 1 must not absorb file 2's entries")

	cfg2, err := clientcmd.LoadFromFile(file2)
	require.NoError(t, err)
	assert.Empty(t, cfg2.CurrentContext, "file 2 must stay untouched")

	// The merged view reflects the switch.
	_, current, err := LoadContexts(list)
	require.NoError(t, err)
	assert.Equal(t, "two-ctx", current)

	// Missing contexts still fail against the merged view.
	require.Error(t, SwitchContext(list, "ghost"))
}

func TestSwitchContext_KubeconfigList_NoCurrentContextAnywhere(t *testing.T) {
	dir := t.TempDir()
	file2 := filepath.Join(dir, "config2")
	// file2 sets no current-context and the first list entry does not exist →
	// kubectl writes into the first EXISTING file.
	require.NoError(t, os.WriteFile(file2, []byte(kubeconfigFileTwo), 0o600))
	missing := filepath.Join(dir, "nope")
	list := missing + string(os.PathListSeparator) + file2

	require.NoError(t, SwitchContext(list, "two-ctx"))

	assert.NoFileExists(t, missing, "a missing list entry must not be created")
	cfg2, err := clientcmd.LoadFromFile(file2)
	require.NoError(t, err)
	assert.Equal(t, "two-ctx", cfg2.CurrentContext)
}
