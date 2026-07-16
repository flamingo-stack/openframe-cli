package services

import (
	"os"
	"path/filepath"
	"testing"

	chartUI "github.com/flamingo-stack/openframe-cli/internal/chart/ui"
	"github.com/flamingo-stack/openframe-cli/internal/chart/ui/templates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	clusterDomain "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// twoClusters is the shared fixture: more than one cluster and no obvious
// default, which is exactly when 0.4.6 dropped into an interactive picker.
func twoClusters() *MockClusterLister {
	m := NewMockClusterLister()
	m.SetClusters([]clusterDomain.ClusterInfo{
		{Name: "cluster-a", Status: "running"},
		{Name: "cluster-b", Status: "running"},
	})
	return m
}

// Finding 8 (0.4.6 regression guard): with --non-interactive and NO cluster-name
// argument, cluster selection must fail fast with an actionable error instead of
// dropping into an interactive picker that hangs CI forever.
func TestClusterSelector_NonInteractive_NoName_FailsFast(t *testing.T) {
	sel := NewClusterSelector(twoClusters(), chartUI.NewOperationsUI())

	name, err := sel.SelectCluster(nil, true /*nonInteractive*/, false)

	require.Error(t, err, "non-interactive with no name must fail fast, not prompt")
	assert.Empty(t, name)
	assert.Contains(t, err.Error(), "non-interactive")
	// The error must list the candidates so the operator knows what to pass.
	assert.Contains(t, err.Error(), "cluster-a")
	assert.Contains(t, err.Error(), "cluster-b")
}

// A blank/whitespace arg is treated the same as no name (still must fail fast).
func TestClusterSelector_NonInteractive_BlankName_FailsFast(t *testing.T) {
	sel := NewClusterSelector(twoClusters(), chartUI.NewOperationsUI())

	_, err := sel.SelectCluster([]string{"   "}, true, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive")
}

// A valid cluster name provided non-interactively is honored directly — no
// prompt, no error — so the CI happy path (bootstrap passes the cluster name)
// keeps working.
func TestClusterSelector_NonInteractive_WithName_OK(t *testing.T) {
	sel := NewClusterSelector(twoClusters(), chartUI.NewOperationsUI())

	name, err := sel.SelectCluster([]string{"cluster-b"}, true, false)

	require.NoError(t, err)
	assert.Equal(t, "cluster-b", name)
}

// Finding 1 (0.4.6 regression guard): an explicitly pinned ref (--ref) must
// be written into the flattened top-level
// repository.branch so BOTH the app-of-apps clone and the child Applications'
// targetRevision track it — rather than silently staying on the values-file
// branch (0.4.6 left every Application on "main" while reporting success).
func TestBuildConfiguration_ExplicitRefPinsRepositoryBranch(t *testing.T) {
	w := newTestWorkflow(t)

	// A real temp values file whose branch says "main"; the explicit ref must win.
	valuesPath := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	require.NoError(t, os.WriteFile(valuesPath, []byte("repository:\n  branch: main\n"), 0o600))

	req := types.InstallationRequest{
		GitHubRepo:        "https://github.com/flamingo-stack/openframe-oss-tenant",
		GitHubBranch:      "feature-x",
		GitHubRefExplicit: true,
	}
	_, err := w.buildConfiguration(req, "test", &types.ChartConfiguration{TempHelmValuesPath: valuesPath})
	require.NoError(t, err)

	modifier := templates.NewHelmValuesModifier()
	values, err := modifier.LoadExistingValues(valuesPath)
	require.NoError(t, err)
	assert.Equal(t, "feature-x", modifier.GetCurrentOSSBranch(values),
		"explicit ref must be pinned to top-level repository.branch, not left on the values-file branch")
}

// Without an explicit ref the values-file branch is left untouched (no
// accidental rewrite of an operator-provided branch).
func TestBuildConfiguration_NoExplicitRef_LeavesBranch(t *testing.T) {
	w := newTestWorkflow(t)

	valuesPath := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	require.NoError(t, os.WriteFile(valuesPath, []byte("repository:\n  branch: develop\n"), 0o600))

	req := types.InstallationRequest{
		GitHubRepo:   "https://github.com/flamingo-stack/openframe-oss-tenant",
		GitHubBranch: "main", // default, not explicitly set
		// GitHubRefExplicit: false
	}
	_, err := w.buildConfiguration(req, "test", &types.ChartConfiguration{TempHelmValuesPath: valuesPath})
	require.NoError(t, err)

	got, err := os.ReadFile(valuesPath) // #nosec G304 -- test-created path
	require.NoError(t, err)
	assert.Contains(t, string(got), "develop", "values-file branch must be preserved when no ref is pinned")
}
