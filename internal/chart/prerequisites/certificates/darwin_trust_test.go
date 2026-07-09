package certificates

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveLoginKeychain covers the macOS keychain selection: use the
// default-keychain when it exists (quotes stripped), else the login.keychain-db,
// else nothing.
func TestResolveLoginKeychain(t *testing.T) {
	home := "/Users/dev"
	loginDB := filepath.Join(home, "Library/Keychains/login.keychain-db")

	// default-keychain output (quoted, as `security` prints it) that exists.
	existsOnly := func(want string) func(string) bool {
		return func(p string) bool { return p == want }
	}
	assert.Equal(t, "/Users/dev/Library/Keychains/login.keychain-db",
		resolveLoginKeychain(`    "/Users/dev/Library/Keychains/login.keychain-db"`+"\n", home, existsOnly("/Users/dev/Library/Keychains/login.keychain-db")))

	// default-keychain reported but missing on disk → fall back to login.keychain-db.
	assert.Equal(t, loginDB,
		resolveLoginKeychain(`"/gone/custom.keychain-db"`, home, existsOnly(loginDB)))

	// Empty default + login.keychain-db present → login.keychain-db.
	assert.Equal(t, loginDB, resolveLoginKeychain("", home, existsOnly(loginDB)))

	// Nothing exists → "".
	assert.Equal(t, "", resolveLoginKeychain("", home, func(string) bool { return false }))
}

// TestParseMkcertCertSHAs proves the SHA-1 extraction from real `security
// find-certificate -Z` output (replacing the piped awk).
func TestParseMkcertCertSHAs(t *testing.T) {
	out := `keychain: "/Users/dev/Library/Keychains/login.keychain-db"
    version: 256
    SHA-1 hash: A1B2C3D4E5F600112233445566778899AABBCCDD
    SHA-256 hash: ignored
    "labl"<blob>="mkcert dev"
    SHA-1 hash:  0011223344556677889900AABBCCDDEEFF001122
`
	assert.Equal(t, []string{
		"A1B2C3D4E5F600112233445566778899AABBCCDD",
		"0011223344556677889900AABBCCDDEEFF001122",
	}, parseMkcertCertSHAs(out))

	assert.Empty(t, parseMkcertCertSHAs(""))
	assert.Empty(t, parseMkcertCertSHAs("no matching lines here"))
}

// TestClassifyAddTrustedCert locks the decision that drives silent vs interactive
// vs fatal handling of `security add-trusted-cert`.
func TestClassifyAddTrustedCert(t *testing.T) {
	assert.Equal(t, trustAdded, classifyAddTrustedCert("", nil))

	err := errors.New("exit status 1")
	assert.Equal(t, trustNeedsInteractive,
		classifyAddTrustedCert("SecTrustSettingsSetTrustSettings: User interaction is not allowed.", err))
	assert.Equal(t, trustCancelled,
		classifyAddTrustedCert("The authorization was canceled by the user.", err))
	assert.Equal(t, trustOtherError,
		classifyAddTrustedCert("some other failure", err))
}
