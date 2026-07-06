package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/sigstore-go/pkg/testing/ca"
	"github.com/stretchr/testify/require"
)

// TestVerifyChecksumsWithVirtualSigstore exercises the real signature-verification
// path (sigVerifier + the pinned identity) against an in-memory virtual Sigstore
// — no network, no TUF. It proves the happy path passes and that both a tampered
// artifact and a wrong-signer identity are rejected.
func TestVerifyChecksumsWithVirtualSigstore(t *testing.T) {
	vs, err := ca.NewVirtualSigstore()
	require.NoError(t, err)

	checksums := []byte("abc123  openframe-cli_linux_amd64.tar.gz\n")

	id, err := pinnedIdentity()
	require.NoError(t, err)
	v := sigVerifier{trust: vs, identity: id}

	// A signature from this repo's release workflow with the pinned issuer.
	good := "https://github.com/flamingo-stack/openframe-cli/.github/workflows/release.yml@refs/heads/main"
	entity, err := vs.Sign(good, signerIssuer, checksums)
	require.NoError(t, err)
	require.NoError(t, v.verifyArtifact(checksums, entity), "a valid, correctly-identified signature must verify")

	// Same signature, tampered artifact → reject.
	require.Error(t, v.verifyArtifact([]byte("tampered checksums"), entity))

	// Correct issuer but a different repo's workflow → fails the SAN pin.
	evil, err := vs.Sign("https://github.com/evil/openframe-cli/.github/workflows/release.yml@refs/heads/main", signerIssuer, checksums)
	require.NoError(t, err)
	require.Error(t, v.verifyArtifact(checksums, evil), "a signature from another repo must be rejected")

	// Right SAN, wrong issuer → fails the issuer pin.
	wrongIssuer, err := vs.Sign(good, "https://accounts.google.com", checksums)
	require.NoError(t, err)
	require.Error(t, v.verifyArtifact(checksums, wrongIssuer), "a signature from another issuer must be rejected")
}

// TestPinnedIdentityCompiles guards that the pinned issuer/SAN regex are valid
// matcher inputs (a malformed regex would only fail at runtime otherwise).
func TestPinnedIdentity(t *testing.T) {
	if _, err := pinnedIdentity(); err != nil {
		t.Fatalf("pinnedIdentity: %v", err)
	}
}

// TestApplyRefusesUnsignedRelease verifies that, without the insecure override,
// a release lacking a signature bundle is refused before anything is installed.
func TestApplyRefusesUnsignedRelease(t *testing.T) {
	if os.Getenv(insecureSkipEnv) != "" {
		t.Skip("insecure override set in the environment")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("deadbeef  openframe-cli_linux_amd64.tar.gz\n"))
	}))
	defer srv.Close()

	rel := Release{TagName: "v9.9.9", Assets: []Asset{
		{Name: checksumsFile, URL: srv.URL + "/checksums.txt"}, // no bundle asset
	}}
	dir := t.TempDir()
	exe := filepath.Join(dir, "openframe")
	_ = os.WriteFile(exe, []byte("OLD"), 0o755)

	u := Updater{Current: "v1.0.0", GOOS: "linux", GOARCH: "amd64", Client: Client{APIBase: srv.URL}, exePath: exe}
	err := u.Apply(context.Background(), rel, nil)
	require.Error(t, err, "an unsigned release must be refused")
	if got, _ := os.ReadFile(exe); string(got) != "OLD" {
		t.Fatalf("binary was modified for an unsigned release: %q", got)
	}
}
