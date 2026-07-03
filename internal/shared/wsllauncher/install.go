package wsllauncher

import (
	"fmt"
	"os/exec"
	"strings"
)

// releaseRepo is the GitHub repository releases are published to.
const releaseRepo = "flamingo-stack/openframe-cli"

// isReleaseVersion reports whether version looks like a real published release
// (not a dev/snapshot build), so we know a Linux artifact exists to download.
func isReleaseVersion(version string) bool {
	switch strings.TrimSpace(version) {
	case "", "dev", "none", "unknown", "snapshot":
		return false
	}
	// A snapshot/pseudo build often carries a commit suffix; releases are clean
	// semver like "1.2.3".
	return !strings.Contains(version, "-")
}

// releaseTag reconstructs the git tag for a goreleaser .Version (which has the
// leading "v" stripped).
func releaseTag(version string) string {
	v := strings.TrimSpace(version)
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

// linuxArchiveName is the goreleaser archive filename for the Linux build of the
// given GOARCH (name_template "openframe-cli_{{.Os}}_{{.Arch}}", tar.gz).
func linuxArchiveName(goarch string) string {
	return fmt.Sprintf("openframe-cli_linux_%s.tar.gz", goarch)
}

// releaseAssetURL builds the download URL for a release asset.
func releaseAssetURL(version, asset string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", releaseRepo, releaseTag(version), asset)
}

// installScript returns a bash script (run inside WSL) that downloads the Linux
// archive, verifies it against the release checksums, and installs the openframe
// binary into ~/.openframe/bin. It is a pure function so the logic is testable.
func installScript(archiveURL, checksumsURL, archiveName string) string {
	// Single-quoted heredoc-free script; all inputs are our own constant-derived
	// URLs (no user data).
	return strings.Join([]string{
		"set -e",
		`BIN_DIR="$HOME/.openframe/bin"`,
		`mkdir -p "$BIN_DIR"`,
		`TMP="$(mktemp -d)"`,
		`trap 'rm -rf "$TMP"' EXIT`,
		`cd "$TMP"`,
		fmt.Sprintf(`curl -fsSL -o archive.tar.gz %q`, archiveURL),
		fmt.Sprintf(`curl -fsSL -o checksums.txt %q`, checksumsURL),
		fmt.Sprintf(`EXPECTED="$(grep " %s$" checksums.txt | awk '{print $1}')"`, archiveName),
		`ACTUAL="$(sha256sum archive.tar.gz | awk '{print $1}')"`,
		`[ -n "$EXPECTED" ] && [ "$EXPECTED" = "$ACTUAL" ] || { echo "checksum verification failed" >&2; exit 1; }`,
		`tar -xzf archive.tar.gz openframe`,
		`install -m 0755 openframe "$BIN_DIR/openframe"`,
	}, "\n")
}

// ensureOpenframeInWSL makes sure the openframe binary is available inside WSL,
// auto-installing the matching Linux release when possible. It returns setup
// guidance if the binary is missing and cannot be installed automatically (dev
// build, or the download/verify failed) — in that case the caller falls back to
// showing instructions.
func ensureOpenframeInWSL(version, goarch string) error {
	if verifyOpenframeInWSL() == nil {
		return nil
	}
	if !isReleaseVersion(version) {
		return notInstalledError()
	}
	if err := installOpenframeInWSL(version, goarch); err != nil {
		return notInstalledError()
	}
	// Re-verify: the install put it under ~/.openframe/bin, which a login shell
	// (bash -lc) resolves via PATH.
	if verifyOpenframeInWSL() != nil {
		return notInstalledError()
	}
	return nil
}

// installOpenframeInWSL runs the install script inside WSL. Thin exec wrapper
// around the tested installScript / URL builders.
func installOpenframeInWSL(version, goarch string) error {
	archive := linuxArchiveName(goarch)
	script := installScript(
		releaseAssetURL(version, archive),
		releaseAssetURL(version, "checksums.txt"),
		archive,
	)
	cmd := exec.Command("wsl", "-d", Distro, "--", "bash", "-lc", script) // #nosec G204 -- script built from constant-derived release URLs, no user input
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("installing openframe inside WSL failed: %w\n%s", err, string(out))
	}
	return nil
}
