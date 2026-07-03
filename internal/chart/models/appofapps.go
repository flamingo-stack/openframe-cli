package models

import (
	"fmt"
	"net/url"
	"strings"
)

// AppOfAppsConfig holds configuration for app-of-apps installation
type AppOfAppsConfig struct {
	// GitHub repository configuration
	GitHubRepo   string // Repository URL (e.g., "https://github.com/flamingo-stack/openframe-oss-tenant")
	GitHubBranch string // Branch to use (e.g., "main", "develop")
	ChartPath    string // Path to chart in repository (e.g., "manifests/app-of-apps")
	// Certificate configuration
	CertDir string // Directory containing certificates for TLS configuration
	// Values configuration
	ValuesFile string // Path to values file
	// Helm configuration
	Namespace string // Target namespace (e.g., "argocd")
	Timeout   string // Installation timeout (e.g., "60m")
}

// NewAppOfAppsConfig creates a new AppOfAppsConfig with defaults
func NewAppOfAppsConfig() *AppOfAppsConfig {
	return &AppOfAppsConfig{
		GitHubRepo:   RepoOSSTenant,
		GitHubBranch: DefaultGitBranch,
		ChartPath:    "manifests/app-of-apps",
		Namespace:    "argocd",
		Timeout:      "60m",
	}
}

// GetGitURL returns the formatted git URL for helm-git plugin
func (a *AppOfAppsConfig) GetGitURL() string {
	// helm-git plugin v1.4.0 format: git+https://github.com/org/repo@path?ref=branch
	//
	// Any embedded credential (e.g. an "x-access-token:PAT@" userinfo) is
	// stripped here so a token can never leak into the helm-git URL — and from
	// there into helm values, argv, or logs (audit I1). Authentication must be
	// supplied out of band (Git credentials / environment).
	baseURL := stripURLCredentials(strings.TrimSuffix(a.GitHubRepo, ".git"))
	return fmt.Sprintf("git+%s@%s?ref=%s", baseURL, a.ChartPath, a.GitHubBranch)
}

// stripURLCredentials removes any userinfo (username[:password]) from an
// absolute URL. It returns the input unchanged when it does not parse as a URL
// or carries no credential.
func stripURLCredentials(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = nil
	return u.String()
}
