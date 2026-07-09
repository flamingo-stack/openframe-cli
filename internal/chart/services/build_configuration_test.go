package services

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
)

// newTestWorkflow builds an InstallationWorkflow good enough for
// buildConfiguration (it only needs operationsUI + configService).
func newTestWorkflow(t *testing.T) *InstallationWorkflow {
	t.Helper()
	svc, err := NewChartServiceDeferred(NewMockClusterLister(), false, false)
	if err != nil {
		t.Fatal(err)
	}
	return &InstallationWorkflow{chartService: svc, clusterService: NewMockClusterLister()}
}

// req uses an explicit non-existent values path so branch resolution never
// picks up a stray openframe-helm-values.yaml from the working directory.
func baseReq() types.InstallationRequest {
	return types.InstallationRequest{GitHubRepo: "https://github.com/custom/repo", GitHubBranch: "main"}
}

// TestBuildConfiguration_RepoStaysClean locks the security contract for the
// only supported (oss-tenant) path: the request's public repo URL with NO
// embedded credentials.
func TestBuildConfiguration_RepoStaysClean(t *testing.T) {
	w := newTestWorkflow(t)
	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		TempHelmValuesPath: "/nonexistent/values.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := cfg.AppOfApps.GitHubRepo
	if repo != "https://github.com/custom/repo" {
		t.Errorf("repo = %q, want the request's repo", repo)
	}
	if strings.Contains(repo, "@") || strings.Contains(repo, "x-access-token") || strings.Contains(repo, "ghp_") {
		t.Errorf("repo must not carry credentials: %q", repo)
	}
}

// The caller's repo/branch pass through untouched.
func TestBuildConfiguration_PassesRequestThrough(t *testing.T) {
	w := newTestWorkflow(t)
	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		TempHelmValuesPath: "/nonexistent/values.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppOfApps.GitHubRepo != "https://github.com/custom/repo" {
		t.Errorf("repo = %q, want the request's repo", cfg.AppOfApps.GitHubRepo)
	}
	if cfg.AppOfApps.GitHubBranch != "main" {
		t.Errorf("branch = %q, want main", cfg.AppOfApps.GitHubBranch)
	}
}
