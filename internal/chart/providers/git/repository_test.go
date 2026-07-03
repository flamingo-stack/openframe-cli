package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// makeLocalRepo creates a real local git repository containing chartPath and
// returns its filesystem URL plus the branch its single commit lives on. This
// lets the clone path run end-to-end (go-git, no network, no git binary).
func makeLocalRepo(t *testing.T, chartPath string) (repoURL, branch string) {
	t.Helper()
	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, chartPath), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, chartPath, "Chart.yaml"), []byte("name: test\n"), 0o600))

	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, wt.AddGlob("."))
	_, err = wt.Commit("init", &gogit.CommitOptions{
		Author: &object.Signature{Name: "t", Email: "t@example.com", When: time.Unix(1700000000, 0)},
	})
	require.NoError(t, err)

	head, err := repo.Head()
	require.NoError(t, err)
	return dir, head.Name().Short()
}

func TestCloneChartRepository_Success(t *testing.T) {
	url, branch := makeLocalRepo(t, "manifests/app-of-apps")
	repo := NewRepository()

	res, err := repo.CloneChartRepository(context.Background(), &models.AppOfAppsConfig{
		GitHubRepo:   url,
		GitHubBranch: branch,
		ChartPath:    "manifests/app-of-apps",
	})
	require.NoError(t, err)
	t.Cleanup(func() { repo.Cleanup(res.TempDir) })

	if _, err := os.Stat(filepath.Join(res.ChartPath, "Chart.yaml")); err != nil {
		t.Fatalf("cloned chart path must contain Chart.yaml: %v", err)
	}
	// The cloned .git must exist — proving a real checkout happened.
	if _, err := os.Stat(filepath.Join(res.TempDir, ".git")); err != nil {
		t.Errorf("expected a .git in the clone: %v", err)
	}
}

func TestCloneChartRepository_BranchNotFound(t *testing.T) {
	url, _ := makeLocalRepo(t, "manifests/app-of-apps")
	repo := NewRepository()

	_, err := repo.CloneChartRepository(context.Background(), &models.AppOfAppsConfig{
		GitHubRepo:   url,
		GitHubBranch: "does-not-exist",
		ChartPath:    "manifests/app-of-apps",
	})
	require.Error(t, err)
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected a friendly branch-not-found error, got: %v", err)
	}
}

func TestCloneChartRepository_ChartPathMissing(t *testing.T) {
	url, branch := makeLocalRepo(t, "manifests/app-of-apps")
	repo := NewRepository()

	_, err := repo.CloneChartRepository(context.Background(), &models.AppOfAppsConfig{
		GitHubRepo:   url,
		GitHubBranch: branch,
		ChartPath:    "no/such/path",
	})
	require.Error(t, err)
	if !strings.Contains(err.Error(), "does not exist in repository") {
		t.Errorf("expected a chart-path error, got: %v", err)
	}
}

// TestCloneChartRepository_NoCredentialFileLeftBehind is the I1 regression
// guard: cloning must not create an on-disk credentials file (the go-git
// migration removed the ofcred temp file entirely).
func TestCloneChartRepository_NoCredentialFileLeftBehind(t *testing.T) {
	before := ofcredCount(t)

	url, branch := makeLocalRepo(t, "manifests/app-of-apps")
	repo := NewRepository()
	res, err := repo.CloneChartRepository(context.Background(), &models.AppOfAppsConfig{
		GitHubRepo:   url,
		GitHubBranch: branch,
		ChartPath:    "manifests/app-of-apps",
	})
	require.NoError(t, err)
	t.Cleanup(func() { repo.Cleanup(res.TempDir) })

	if after := ofcredCount(t); after != before {
		t.Errorf("clone must not create a git credentials file (ofcred-*): before=%d after=%d", before, after)
	}
}

func ofcredCount(t *testing.T) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "ofcred-*"))
	require.NoError(t, err)
	return len(matches)
}
