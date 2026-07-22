package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configurationsJSON = `[
  {"name": "default", "properties": {"core": {"project": "none"}}},
  {"name": "dev-tenant-runners", "properties": {"core": {"project": "tenant-runners-db9z"}}},
  {"name": "dev-shared", "properties": {"core": {"project": "shared-j62b"}}},
  {"name": "dev-shared-dup", "properties": {"core": {"project": "shared-j62b"}}},
  {"name": "prod-shared", "properties": {"core": {"project": "shared-4t5d"}}}
]`

const runnersClustersJSON = `[
  {"name": "tenant-cluster-1", "location": "us-central1", "status": "RUNNING",
   "currentNodeCount": 3, "currentMasterVersion": "1.33.2-gke.100"}
]`

func kubeconfigWith(t *testing.T, contexts map[string]string) {
	t.Helper()
	content := "apiVersion: v1\nkind: Config\nclusters:\ncontexts:\n"
	for name := range contexts {
		content += "- name: " + name + "\n  context:\n    cluster: c\n    user: u\n"
	}
	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	t.Setenv("KUBECONFIG", path)
}

func TestAuthStatus(t *testing.T) {
	t.Run("active account", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "dev@flamingo.example\n"})
		assert.Equal(t, Authenticated, NewGKEDiscoverer(mock).AuthStatus(context.Background()))
	})

	t.Run("no active account", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: ""})
		assert.Equal(t, NotAuthenticated, NewGKEDiscoverer(mock).AuthStatus(context.Background()))
	})

	t.Run("gcloud errors map to not-authenticated", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "not logged in")
		assert.Equal(t, NotAuthenticated, NewGKEDiscoverer(mock).AuthStatus(context.Background()))
	})
}

func TestProjects_DedupesAndSkipsEmpty(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: configurationsJSON})

	projects, err := NewGKEDiscoverer(mock).Projects(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"shared-4t5d", "shared-j62b", "tenant-runners-db9z"}, projects)
}

func TestDiscover_ListsClustersAcrossProjects(t *testing.T) {
	kubeconfigWith(t, map[string]string{
		"connectgateway_tenant-runners-db9z_us-central1_tenant-cluster-1": "",
	})

	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: configurationsJSON})
	mock.SetResponse("clusters list --project tenant-runners-db9z", &executor.CommandResult{ExitCode: 0, Stdout: runnersClustersJSON})
	mock.SetResponse("clusters list --project shared-j62b", &executor.CommandResult{ExitCode: 0, Stdout: "[]"})
	// prod project: PERMISSION_DENIED must not break discovery of the rest.
	mock.SetResponse("clusters list --project shared-4t5d", &executor.CommandResult{ExitCode: 1, Stderr: "PERMISSION_DENIED"})

	result, err := NewGKEDiscoverer(mock).Discover(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Clusters, 1)
	c := result.Clusters[0]
	assert.Equal(t, "tenant-cluster-1", c.Name)
	assert.Equal(t, models.ClusterTypeGKE, c.Type)
	assert.Equal(t, models.SourceExternal, c.Source)
	assert.Equal(t, "Running", c.Status)
	assert.Equal(t, 3, c.NodeCount)
	assert.Equal(t, "tenant-runners-db9z", c.Project)
	assert.Equal(t, "us-central1", c.Region)
	assert.Equal(t, "connectgateway_tenant-runners-db9z_us-central1_tenant-cluster-1", c.Context)

	require.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "shared-4t5d")
}

func TestMatchContext(t *testing.T) {
	contexts := []string{
		"k3d-openframe-dev",
		"gke_proj_us-central1_alpha",
		"connectgateway_proj_us-central1_beta",
		"gamma",
	}
	assert.Equal(t, "gke_proj_us-central1_alpha", matchContext(contexts, "proj", "us-central1", "alpha"))
	assert.Equal(t, "connectgateway_proj_us-central1_beta", matchContext(contexts, "proj", "us-central1", "beta"))
	assert.Equal(t, "gamma", matchContext(contexts, "proj", "us-central1", "gamma"))
	assert.Equal(t, "", matchContext(contexts, "proj", "us-central1", "delta"))
}

func TestConfigurationForProject(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: configurationsJSON})
	d := NewGKEDiscoverer(mock)

	name, err := d.ConfigurationForProject(context.Background(), "tenant-runners-db9z")
	require.NoError(t, err)
	assert.Equal(t, "dev-tenant-runners", name)

	name, err = d.ConfigurationForProject(context.Background(), "unknown-project")
	require.NoError(t, err)
	assert.Equal(t, "", name)
}
