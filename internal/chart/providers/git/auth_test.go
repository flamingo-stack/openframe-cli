package git

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "ghp_SECRETtoken1234567890" //nolint:gosec // test fixture, not a real credential

// TestClone_TokenNeverInArgv is the I1 regression guard: a PAT embedded in the
// repo URL must never appear in the git command line; it must travel via the
// credentials file + store helper instead.
func TestClone_TokenNeverInArgv(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	repo := NewRepository(mock)

	cfg := &models.AppOfAppsConfig{
		GitHubRepo:   "https://x-access-token:" + fakeToken + "@github.com/flamingo-stack/openframe-saas-tenant",
		GitHubBranch: "main",
		ChartPath:    "manifests/app-of-apps",
	}

	// The clone "succeeds" (mock default), then chart-path stat fails — we only
	// care about the command that was constructed.
	_, _ = repo.CloneChartRepository(context.Background(), cfg)

	cmds := mock.Commands()
	require.NotEmpty(t, cmds, "expected a git command to be recorded")

	// Token must not be in any argv element or command name.
	testutil.AssertNoArgContains(t, cmds, fakeToken)

	// The clone URL argument must be the clean URL (no userinfo).
	testutil.AssertArgIsDiscrete(t, cmds, "https://github.com/flamingo-stack/openframe-saas-tenant")

	// Auth must be supplied via the file-based store helper.
	var sawStoreHelper bool
	var env map[string]string
	for _, c := range cmds {
		if c.Name != "git" {
			continue
		}
		env = c.Env
		for _, a := range c.Args {
			if strings.HasPrefix(a, "credential.helper=store --file=") {
				sawStoreHelper = true
			}
		}
	}
	assert.True(t, sawStoreHelper, "expected a file-based store credential helper, got %v", cmds)
	assert.Equal(t, "0", env["GIT_TERMINAL_PROMPT"], "non-interactive prompts must be disabled")
}

// TestClone_PublicRepoNoCredentialFile: a public URL produces no credential
// helper and no token handling.
func TestClone_PublicRepoNoCredentialFile(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	repo := NewRepository(mock)

	cfg := &models.AppOfAppsConfig{
		GitHubRepo:   "https://github.com/flamingo-stack/openframe-oss-tenant",
		GitHubBranch: "main",
		ChartPath:    "manifests/app-of-apps",
	}
	_, _ = repo.CloneChartRepository(context.Background(), cfg)

	for _, c := range mock.Commands() {
		for _, a := range c.Args {
			assert.NotContains(t, a, "credential.helper=store", "public repo must not use a credential helper")
		}
	}
}

// TestClone_ErrorOutputRedactsToken: even if git echoes the token in stderr,
// the surfaced error must not contain it.
func TestClone_ErrorOutputRedactsToken(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "fatal: Authentication failed for 'https://x-access-token:"+fakeToken+"@github.com/org/repo'")
	repo := NewRepository(mock)

	cfg := &models.AppOfAppsConfig{
		GitHubRepo:   "https://x-access-token:" + fakeToken + "@github.com/org/repo",
		GitHubBranch: "main",
		ChartPath:    "manifests/app-of-apps",
	}
	_, err := repo.CloneChartRepository(context.Background(), cfg)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), fakeToken, "token leaked into clone error: %s", err.Error())
}

func TestExtractGitAuth(t *testing.T) {
	a := extractGitAuth("https://x-access-token:" + fakeToken + "@github.com/org/repo")
	assert.Equal(t, "https://github.com/org/repo", a.cleanURL)
	assert.Equal(t, "x-access-token", a.username)
	assert.Equal(t, fakeToken, a.token)
	assert.True(t, a.hasToken())

	pub := extractGitAuth("https://github.com/org/repo")
	assert.Equal(t, "https://github.com/org/repo", pub.cleanURL)
	assert.False(t, pub.hasToken())
}

func TestCredentialLine(t *testing.T) {
	a := extractGitAuth("https://x-access-token:" + fakeToken + "@github.com/org/repo")
	line, ok := a.credentialLine()
	require.True(t, ok)
	assert.Equal(t, "https://x-access-token:"+fakeToken+"@github.com", line)

	_, ok = extractGitAuth("https://github.com/org/repo").credentialLine()
	assert.False(t, ok)
}

func TestWriteGitCredentials_Mode0600(t *testing.T) {
	path, cleanup, err := writeGitCredentials("https://x-access-token:" + fakeToken + "@github.com")
	require.NoError(t, err)
	defer cleanup()

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "git-credentials file must be 0600")

	cleanup()
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "credentials file must be removed by cleanup")
}

func TestMaskToken(t *testing.T) {
	assert.Equal(t, "auth failed for ***@github.com",
		maskToken("auth failed for "+fakeToken+"@github.com", fakeToken))
	assert.Equal(t, "no token here", maskToken("no token here", ""))
}
