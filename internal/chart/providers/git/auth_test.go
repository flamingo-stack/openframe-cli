package git

import (
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "ghp_SECRETtoken1234567890" //nolint:gosec // test fixture, not a real credential

func TestExtractGitAuth(t *testing.T) {
	a := extractGitAuth("https://x-access-token:" + fakeToken + "@github.com/org/repo")
	assert.Equal(t, "https://github.com/org/repo", a.cleanURL)
	assert.Equal(t, "x-access-token", a.username)
	assert.Equal(t, fakeToken, a.token)

	// Single-field userinfo (https://<token>@host, a common PAT shorthand): the
	// token is in the username field with no password — it must be recognized as
	// the token, not silently dropped.
	tok := extractGitAuth("https://" + fakeToken + "@github.com/org/repo")
	assert.Equal(t, "https://github.com/org/repo", tok.cleanURL)
	assert.Empty(t, tok.username)
	assert.Equal(t, fakeToken, tok.token)

	pub := extractGitAuth("https://github.com/org/repo")
	assert.Equal(t, "https://github.com/org/repo", pub.cleanURL)
	assert.Empty(t, pub.token)
}

// TestBuildAuth is the I1 guard: a private-repo token is handed to go-git only
// as an in-memory HTTP basic-auth method (never a URL, argv, or on-disk file),
// and a public repo gets no auth at all.
func TestBuildAuth(t *testing.T) {
	t.Run("private repo → in-memory basic auth", func(t *testing.T) {
		a := extractGitAuth("https://x-access-token:" + fakeToken + "@github.com/org/repo")
		m := a.buildAuth()
		basic, ok := m.(*githttp.BasicAuth)
		require.True(t, ok, "expected *http.BasicAuth, got %T", m)
		assert.Equal(t, "x-access-token", basic.Username)
		assert.Equal(t, fakeToken, basic.Password)
	})

	t.Run("custom username is preserved", func(t *testing.T) {
		a := extractGitAuth("https://alice:" + fakeToken + "@github.com/org/repo")
		basic := a.buildAuth().(*githttp.BasicAuth)
		assert.Equal(t, "alice", basic.Username)
		assert.Equal(t, fakeToken, basic.Password)
	})

	t.Run("token-only URL → default username + token as password", func(t *testing.T) {
		a := extractGitAuth("https://" + fakeToken + "@github.com/org/repo")
		basic := a.buildAuth().(*githttp.BasicAuth)
		assert.Equal(t, GitTokenUser, basic.Username)
		assert.Equal(t, fakeToken, basic.Password)
	})

	t.Run("public repo → no auth", func(t *testing.T) {
		a := extractGitAuth("https://github.com/org/repo")
		assert.Nil(t, a.buildAuth(), "public repo must not carry auth")
	})
}

func TestMaskToken(t *testing.T) {
	assert.Equal(t, "auth failed for ***@github.com",
		maskToken("auth failed for "+fakeToken+"@github.com", fakeToken))
	assert.Equal(t, "no token here", maskToken("no token here", ""))
}
