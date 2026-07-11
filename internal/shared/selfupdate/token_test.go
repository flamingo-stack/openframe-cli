package selfupdate

import "testing"

// TestGitHubToken_AcceptsBothConventions locks the fix for the CI 403: the
// updater reads GITHUB_TOKEN (Actions) but must also honour GH_TOKEN (the gh
// CLI). A step that exported only GH_TOKEN left the updater unauthenticated,
// which rate-limited to HTTP 403 on shared macOS runner IPs.
func TestGitHubToken_AcceptsBothConventions(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	if got := GitHubToken(); got != "" {
		t.Errorf("no token set must yield empty, got %q", got)
	}

	t.Setenv("GH_TOKEN", "gh-cli-token")
	if got := GitHubToken(); got != "gh-cli-token" {
		t.Errorf("GH_TOKEN must be honoured when GITHUB_TOKEN is unset, got %q", got)
	}

	// GITHUB_TOKEN wins when both are present (the Actions-native name).
	t.Setenv("GITHUB_TOKEN", "actions-token")
	if got := GitHubToken(); got != "actions-token" {
		t.Errorf("GITHUB_TOKEN must take precedence, got %q", got)
	}
}
