package git

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// gitAuth holds credentials extracted from a clone URL together with the
// cleaned, credential-free URL.
//
// Keeping the token OUT of the URL is what keeps it off the git command line
// (audit I1): the token is re-supplied to git via a 0600 credentials file and
// the `store` credential helper, so it never appears in argv (visible to
// `ps`/`/proc/<pid>/cmdline`) nor in process/error output.
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
	token, _ := u.User.Password()
	u.User = nil
	return gitAuth{cleanURL: u.String(), username: username, token: token}
}

// hasToken reports whether a credential was present.
func (a gitAuth) hasToken() bool { return a.token != "" }

// credentialLine renders a single git-credentials `store` line for the URL's
// host, e.g. "https://x-access-token:TOKEN@github.com".
func (a gitAuth) credentialLine() (string, bool) {
	if a.token == "" {
		return "", false
	}
	u, err := url.Parse(a.cleanURL)
	if err != nil || u.Host == "" {
		return "", false
	}
	user := a.username
	if user == "" {
		user = "x-access-token"
	}
	return fmt.Sprintf("%s://%s:%s@%s", u.Scheme, user, a.token, u.Host), true
}

// writeGitCredentials writes line to a private (0600) temp file and returns its
// path plus a cleanup func. The token lives only in this file, never in argv.
func writeGitCredentials(line string) (string, func(), error) {
	noop := func() {}
	// os.CreateTemp creates the file with 0600 perms.
	f, err := os.CreateTemp("", "ofcred-*")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.Remove(f.Name()) }
	// Be explicit about the mode in case of a permissive umask/temp dir.
	if err := os.Chmod(f.Name(), 0o600); err != nil {
		_ = f.Close()
		cleanup()
		return "", noop, err
	}
	if _, err := f.WriteString(line + "\n"); err != nil {
		_ = f.Close()
		cleanup()
		return "", noop, err
	}
	if err := f.Close(); err != nil {
		cleanup()
		return "", noop, err
	}
	return f.Name(), cleanup, nil
}

// maskToken replaces every occurrence of token in s with a redaction marker.
// No-op when token is empty.
func maskToken(s, token string) string {
	if token == "" {
		return s
	}
	return strings.ReplaceAll(s, token, "***")
}
