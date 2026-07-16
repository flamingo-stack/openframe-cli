package git

import (
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// GitTokenUser is the conventional username for GitHub token (PAT) auth: the
// token is the password and the username is ignored, so "x-access-token" is
// used by convention. Single source of truth for both the in-memory clone auth
// and any token-in-URL the CLI builds.
const GitTokenUser = "x-access-token"

// gitAuth holds credentials extracted from a clone URL together with the
// cleaned, credential-free URL.
//
// Keeping the token OUT of the URL is what keeps it off the process command
// line and out of any on-disk file (audit I1): it is handed to go-git only via
// an in-memory HTTP basic-auth method (see buildAuth), so it never appears in
// argv (visible to `ps`/`/proc/<pid>/cmdline`), in a credentials file, nor in
// process/error output.
type gitAuth struct {
	cleanURL string // URL with any userinfo stripped
	username string // username from the URL, if any
	token    string // password / PAT; empty for public repos
}

// extractGitAuth parses rawURL and separates any embedded credential from it.
func extractGitAuth(rawURL string) gitAuth {
	u, err := url.Parse(rawURL)
	if err != nil || u.User == nil {
		return gitAuth{cleanURL: rawURL}
	}
	username := u.User.Username()
	token, hasPassword := u.User.Password()
	// A single-field userinfo (e.g. https://<token>@host, a common GitHub PAT
	// shorthand) carries the token as the username with no password. Treat it as
	// the token so it is used for auth (and masked in output) rather than
	// silently stripped from the URL and dropped.
	if !hasPassword {
		token = username
		username = ""
	}
	u.User = nil
	return gitAuth{cleanURL: u.String(), username: username, token: token}
}

// buildAuth returns the in-memory HTTP auth method for a private repository, or
// nil for a public one. The token lives only in memory — never in the URL,
// argv, or a credentials file. GitHub PAT auth expects the token as the
// password with any username (conventionally "x-access-token").
func (a gitAuth) buildAuth() transport.AuthMethod {
	if a.token == "" {
		return nil
	}
	user := a.username
	if user == "" {
		user = GitTokenUser
	}
	return &githttp.BasicAuth{Username: user, Password: a.token}
}

// maskToken replaces every occurrence of token in s with a redaction marker.
// No-op when token is empty.
func maskToken(s, token string) string {
	if token == "" {
		return s
	}
	return strings.ReplaceAll(s, token, "***")
}
