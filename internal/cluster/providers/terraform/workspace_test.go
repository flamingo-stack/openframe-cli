package terraform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRecord(name string) Record {
	return Record{
		Name:      name,
		Type:      models.ClusterTypeEKS,
		Status:    StatusCreating,
		Region:    "us-east-1",
		NodeCount: 3,
		CreatedAt: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
	}
}

func TestWorkspace_ScaffoldAndReadBack(t *testing.T) {
	base := t.TempDir()
	ws := OpenWorkspace(base, "demo")

	require.False(t, ws.Exists())
	require.NoError(t, ws.Scaffold(testRecord("demo"), []byte("# tf"), map[string]any{"region": "us-east-1"}))
	require.True(t, ws.Exists())

	rec, err := ws.ReadRecord()
	require.NoError(t, err)
	assert.Equal(t, "demo", rec.Name)
	assert.Equal(t, StatusCreating, rec.Status)

	mainTF, err := os.ReadFile(filepath.Join(ws.TerraformDir(), "main.tf"))
	require.NoError(t, err)
	assert.Equal(t, "# tf", string(mainTF))

	var vars map[string]any
	data, err := os.ReadFile(filepath.Join(ws.TerraformDir(), "terraform.tfvars.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &vars))
	assert.Equal(t, "us-east-1", vars["region"])
}

func TestWorkspace_SetStatus(t *testing.T) {
	ws := OpenWorkspace(t.TempDir(), "demo")
	require.NoError(t, ws.Scaffold(testRecord("demo"), nil, nil))

	require.NoError(t, ws.SetStatus(StatusReady))
	rec, err := ws.ReadRecord()
	require.NoError(t, err)
	assert.Equal(t, StatusReady, rec.Status)
}

func TestWorkspace_Remove(t *testing.T) {
	base := t.TempDir()
	ws := OpenWorkspace(base, "demo")
	require.NoError(t, ws.Scaffold(testRecord("demo"), nil, nil))

	require.NoError(t, ws.Remove())
	assert.False(t, ws.Exists())
}

func TestRegistry_ListSkipsForeignDirs(t *testing.T) {
	base := t.TempDir()
	reg := NewRegistry(base)

	require.NoError(t, reg.Workspace("one").Scaffold(testRecord("one"), nil, nil))
	require.NoError(t, reg.Workspace("two").Scaffold(testRecord("two"), nil, nil))
	// A directory without cluster.json must not break or pollute the listing.
	require.NoError(t, os.MkdirAll(filepath.Join(base, "not-a-cluster"), 0o750))

	records, err := reg.List()
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestRegistry_EmptyBaseIsEmptyRegistry(t *testing.T) {
	reg := NewRegistry(filepath.Join(t.TempDir(), "does-not-exist"))
	records, err := reg.List()
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestRegistry_GetMissingIsClusterNotFound(t *testing.T) {
	reg := NewRegistry(t.TempDir())
	_, err := reg.Get("ghost")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound)
}
