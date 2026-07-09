package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppOfAppsConfig(t *testing.T) {
	config := NewAppOfAppsConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "https://github.com/flamingo-stack/openframe-oss-tenant", config.GitHubRepo)
	assert.Equal(t, "main", config.GitHubBranch)
	assert.Equal(t, "manifests/app-of-apps", config.ChartPath)
	assert.Equal(t, "argocd", config.Namespace)
	assert.Equal(t, "60m", config.Timeout)
	assert.Empty(t, config.CertDir)
	assert.Empty(t, config.ValuesFile)
}

func TestAppOfAppsConfig_GetGitURL(t *testing.T) {
	config := &AppOfAppsConfig{
		GitHubRepo:   "https://github.com/test/repo",
		GitHubBranch: "develop",
		ChartPath:    "charts/app-of-apps",
	}

	gitURL := config.GetGitURL()
	expected := "git+https://github.com/test/repo@charts/app-of-apps?ref=develop"
	assert.Equal(t, expected, gitURL)
}

func TestAppOfAppsConfig_GetGitURL_WithGitSuffix(t *testing.T) {
	config := &AppOfAppsConfig{
		GitHubRepo:   "https://github.com/test/repo.git",
		GitHubBranch: "main",
		ChartPath:    "manifests",
	}

	gitURL := config.GetGitURL()
	expected := "git+https://github.com/test/repo@manifests?ref=main"
	assert.Equal(t, expected, gitURL)
}

func TestAppOfAppsConfig_GetGitURL_StripsEmbeddedToken(t *testing.T) {
	// A PAT embedded in the repo URL (x-access-token:TOKEN@) must never appear
	// in the helm-git URL (audit I1).
	config := &AppOfAppsConfig{
		GitHubRepo:   "https://x-access-token:ghp_supersecret@github.com/test/private-repo",
		GitHubBranch: "main",
		ChartPath:    "manifests/app-of-apps",
	}

	gitURL := config.GetGitURL()

	assert.Equal(t, "git+https://github.com/test/private-repo@manifests/app-of-apps?ref=main", gitURL)
	assert.NotContains(t, gitURL, "ghp_supersecret")
	assert.NotContains(t, gitURL, "x-access-token")
	assert.NotContains(t, gitURL, "@github.com", "userinfo separator must be gone")
}

func TestAppOfAppsConfig_GetGitURL_StripsTokenWithGitSuffix(t *testing.T) {
	config := &AppOfAppsConfig{
		GitHubRepo:   "https://x-access-token:ghp_tok@github.com/test/private-repo.git",
		GitHubBranch: "develop",
		ChartPath:    "charts",
	}

	gitURL := config.GetGitURL()

	assert.Equal(t, "git+https://github.com/test/private-repo@charts?ref=develop", gitURL)
	assert.NotContains(t, gitURL, "ghp_tok")
}

func TestStripURLCredentials(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no credential", "https://github.com/org/repo", "https://github.com/org/repo"},
		{"token userinfo", "https://x-access-token:ghp_x@github.com/org/repo", "https://github.com/org/repo"},
		{"username only", "https://user@github.com/org/repo", "https://github.com/org/repo"},
		{"empty string", "", ""},
		{"not a url", "::not a url::", "::not a url::"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stripURLCredentials(tc.in))
		})
	}
}

func TestAppOfAppsConfig_Fields(t *testing.T) {
	config := &AppOfAppsConfig{}

	// Test that all fields are accessible
	config.GitHubRepo = "https://github.com/test/repo"
	config.GitHubBranch = "develop"
	config.ChartPath = "charts"
	config.CertDir = "/certs"
	config.ValuesFile = "values.yaml"
	config.Namespace = "argocd"
	config.Timeout = "30m"

	assert.Equal(t, "https://github.com/test/repo", config.GitHubRepo)
	assert.Equal(t, "develop", config.GitHubBranch)
	assert.Equal(t, "charts", config.ChartPath)
	assert.Equal(t, "/certs", config.CertDir)
	assert.Equal(t, "values.yaml", config.ValuesFile)
	assert.Equal(t, "argocd", config.Namespace)
	assert.Equal(t, "30m", config.Timeout)
}

func TestAppOfAppsConfig_CompleteConfiguration(t *testing.T) {
	config := &AppOfAppsConfig{
		GitHubRepo:   "https://github.com/test/public-repo",
		GitHubBranch: "feature/new-charts",
		ChartPath:    "helm/charts/app-of-apps",
		CertDir:      "/etc/ssl/certs",
		ValuesFile:   "values.yaml",
		Namespace:    "openframe",
		Timeout:      "90m",
	}

	// Test all methods with complete configuration
	gitURL := config.GetGitURL()
	expected := "git+https://github.com/test/public-repo@helm/charts/app-of-apps?ref=feature/new-charts"
	assert.Equal(t, expected, gitURL)

	assert.Equal(t, "https://github.com/test/public-repo", config.GitHubRepo)
	assert.Equal(t, "feature/new-charts", config.GitHubBranch)
	assert.Equal(t, "helm/charts/app-of-apps", config.ChartPath)
	assert.Equal(t, "/etc/ssl/certs", config.CertDir)
	assert.Equal(t, "values.yaml", config.ValuesFile)
	assert.Equal(t, "openframe", config.Namespace)
	assert.Equal(t, "90m", config.Timeout)
}
