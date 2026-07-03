package services

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
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

func mode(m types.DeploymentMode) *types.DeploymentMode { return &m }

// req uses an explicit non-existent values path so branch resolution never
// picks up a stray helm-values.yaml from the working directory.
func baseReq() types.InstallationRequest {
	return types.InstallationRequest{GitHubRepo: "https://github.com/custom/repo", GitHubBranch: "main"}
}

// TestBuildConfiguration_OSSRepoStaysClean locks the security contract for the
// default (oss-tenant) path: a public repo URL with NO embedded credentials.
func TestBuildConfiguration_OSSRepoStaysClean(t *testing.T) {
	w := newTestWorkflow(t)
	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		DeploymentMode:     mode(types.DeploymentModeOSS),
		TempHelmValuesPath: "/nonexistent/values.yaml",
		// Even with a password present, OSS must never get a token.
		SaaSConfig: &types.SaaSConfig{RepositoryPassword: "ghp_should_not_leak"},
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := cfg.AppOfApps.GitHubRepo
	if repo != "https://github.com/flamingo-stack/openframe-oss-tenant" {
		t.Errorf("oss repo = %q, want the public oss-tenant URL", repo)
	}
	if strings.Contains(repo, "@") || strings.Contains(repo, "x-access-token") || strings.Contains(repo, "ghp_") {
		t.Errorf("oss repo must not carry credentials: %q", repo)
	}
}

// TestBuildConfiguration_SaaSInjectsTokenAndRegistersRedaction locks the SaaS
// contract: the PAT is injected for the private repo AND registered for
// redaction so verbose logs never print it.
func TestBuildConfiguration_SaaSInjectsTokenAndRegistersRedaction(t *testing.T) {
	defer redact.ClearSecrets()
	w := newTestWorkflow(t)

	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		DeploymentMode:     mode(types.DeploymentModeSaaS),
		TempHelmValuesPath: "/nonexistent/values.yaml",
		SaaSConfig:         &types.SaaSConfig{RepositoryPassword: "ghp_supersecret"},
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := cfg.AppOfApps.GitHubRepo
	want := "https://x-access-token:ghp_supersecret@github.com/flamingo-stack/openframe-saas-tenant"
	if repo != want {
		t.Errorf("saas repo = %q, want %q", repo, want)
	}
	// The token must have been registered with the redactor.
	if got := redact.Redact("token is ghp_supersecret here"); strings.Contains(got, "ghp_supersecret") {
		t.Errorf("token was not registered for redaction: %q", got)
	}
}

func TestBuildConfiguration_SaaSSharedUsesSharedRepo(t *testing.T) {
	defer redact.ClearSecrets()
	w := newTestWorkflow(t)

	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		DeploymentMode:     mode(types.DeploymentModeSaaSShared),
		TempHelmValuesPath: "/nonexistent/values.yaml",
		SaaSConfig:         &types.SaaSConfig{RepositoryPassword: "tok"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cfg.AppOfApps.GitHubRepo, "openframe-saas-shared") {
		t.Errorf("saas-shared must target the shared repo, got %q", cfg.AppOfApps.GitHubRepo)
	}
}

// Without a password the SaaS URL must stay clean rather than embed an empty
// credential.
func TestBuildConfiguration_SaaSWithoutPasswordStaysClean(t *testing.T) {
	w := newTestWorkflow(t)
	cfg, err := w.buildConfiguration(baseReq(), "test", &types.ChartConfiguration{
		DeploymentMode:     mode(types.DeploymentModeSaaS),
		TempHelmValuesPath: "/nonexistent/values.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(cfg.AppOfApps.GitHubRepo, "@") {
		t.Errorf("saas repo without password must stay clean: %q", cfg.AppOfApps.GitHubRepo)
	}
}

// With no deployment mode the caller's repo/branch pass through untouched.
func TestBuildConfiguration_NoModePassesRequestThrough(t *testing.T) {
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
