package selfupdate

import (
	"context"
	"fmt"
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
)

// FetchVerifiedLinuxBinary downloads the Linux release binary for the given
// version and GOARCH through the FULL self-update trust chain: release lookup
// via the GitHub API (tolerating a "v" prefix), cosign signature verification
// of checksums.txt against the pinned GitHub Actions identity, then SHA256
// verification of the archive before the openframe binary is extracted.
//
// This exists for the WSL auto-installer: its previous curl-based path checked
// only the checksums file, which comes from the SAME release as the archive —
// worthless against a compromised release upload. A binary that `openframe
// update` would reject must never be installed into WSL either (audit B5/T2).
func FetchVerifiedLinuxBinary(ctx context.Context, version, goarch string, log func(string)) ([]byte, error) {
	client := Client{Token: os.Getenv("GITHUB_TOKEN")}
	rel, err := client.ForTag(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("looking up release %s: %w", version, err)
	}
	return fetchVerifiedLinuxBinary(ctx, client, rel, goarch, log)
}

// fetchVerifiedLinuxBinary is the release-independent core (testable against a
// fixture server).
func fetchVerifiedLinuxBinary(ctx context.Context, client Client, rel Release, goarch string, log func(string)) ([]byte, error) {
	if log == nil {
		log = func(string) {}
	}

	name := archiveName("linux", goarch)
	assetURL, ok := rel.assetURL(name)
	if !ok {
		return nil, fmt.Errorf("release %s has no asset %s", rel.TagName, name)
	}

	// Same trust order as Apply: authenticate the checksums BEFORE using them
	// to verify the archive.
	checksums, err := client.fetchAsset(ctx, rel, checksumsFile)
	if err != nil {
		return nil, err
	}
	u := Updater{Client: client}
	if err := u.verifySignature(ctx, rel, checksums, log); err != nil {
		return nil, err
	}
	sum, err := parseChecksum(string(checksums), name)
	if err != nil {
		return nil, err
	}

	log(fmt.Sprintf("Downloading %s %s for WSL...", binaryName, rel.TagName))
	dl := download.Downloader{}
	return dl.FetchVerifiedTarGzMember(ctx, download.PinnedAsset{URL: assetURL, SHA256: sum}, binaryName)
}
