package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

// Signing-identity pins for the release's keyless cosign signature. A valid
// signature must be issued to GitHub Actions' OIDC provider AND carry a subject
// (SAN) proving it was produced by *this* repo's release workflow — so a
// signature from any other repo, workflow, or issuer is rejected.
const (
	signerIssuer = "https://token.actions.githubusercontent.com"
	// The SAN GitHub Actions puts in the Fulcio cert is
	//   https://github.com/<owner>/<repo>/.github/workflows/<file>@<ref>
	// The ref varies with the trigger (workflow_dispatch from main →
	// refs/heads/main; a tag push → refs/tags/vX.Y.Z), so we pin the exact repo
	// and workflow file but allow either ref form.
	signerSANRegex = `^https://github\.com/flamingo-stack/openframe-cli/\.github/workflows/release\.yml@refs/(heads/main|tags/.+)$`
	// bundleAsset is the sigstore bundle (new format) published alongside
	// checksums.txt; it carries the signature, Fulcio cert, and Rekor entry.
	bundleAsset = checksumsFile + ".bundle"
	// insecureSkipEnv disables signature verification (integrity-only). An
	// escape hatch for emergencies; never set it in normal use.
	insecureSkipEnv = "OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY"
)

// pinnedIdentity builds the certificate-identity policy that a valid release
// signature must satisfy.
func pinnedIdentity() (verify.CertificateIdentity, error) {
	return verify.NewShortCertificateIdentity(signerIssuer, "", "", signerSANRegex)
}

// sigVerifier verifies that an artifact is covered by a signature from a trusted
// signer. It is split from the trust-root bootstrap so it can be unit-tested
// against a virtual Sigstore instance offline.
type sigVerifier struct {
	trust    root.TrustedMaterial
	identity verify.CertificateIdentity
}

// verifyArtifact confirms entity is a valid Sigstore signature over artifact,
// anchored in the transparency log and matching the pinned identity.
func (v sigVerifier) verifyArtifact(artifact []byte, entity verify.SignedEntity) error {
	sev, err := verify.NewVerifier(v.trust,
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	if err != nil {
		return fmt.Errorf("building verifier: %w", err)
	}
	if _, err := sev.Verify(entity, verify.NewPolicy(
		verify.WithArtifact(bytes.NewReader(artifact)),
		verify.WithCertificateIdentity(v.identity),
	)); err != nil {
		return err
	}
	return nil
}

// loadBundle parses a new-format Sigstore bundle from its JSON bytes.
func loadBundle(data []byte) (*bundle.Bundle, error) {
	var b bundle.Bundle
	if err := b.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("parsing signature bundle: %w", err)
	}
	return &b, nil
}

// verifyChecksumsProd is the production checksumVerifier: it bootstraps the
// Sigstore trust root via TUF (cached under the CLI state dir) and verifies the
// checksums bundle against the pinned GitHub Actions identity.
func verifyChecksumsProd(_ context.Context, artifact, bundleJSON []byte) error {
	id, err := pinnedIdentity()
	if err != nil {
		return err
	}
	trust, err := fetchTrustRoot()
	if err != nil {
		return fmt.Errorf("loading Sigstore trust root: %w", err)
	}
	b, err := loadBundle(bundleJSON)
	if err != nil {
		return err
	}
	return sigVerifier{trust: trust, identity: id}.verifyArtifact(artifact, b)
}

// fetchTrustRoot returns the current Sigstore public-good trust root, fetched
// and cached via TUF under the CLI state dir (rather than the default
// ~/.sigstore) so the CLI keeps all of its state in one place.
func fetchTrustRoot() (root.TrustedMaterial, error) {
	opts := tuf.DefaultOptions()
	if home, err := os.UserHomeDir(); err == nil {
		opts.CachePath = filepath.Join(home, ".openframe", "state", "tuf")
	}
	return root.FetchTrustedRootWithOptions(opts)
}
