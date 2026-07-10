package wsllauncher

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// releaseRepo is the GitHub repository releases are published to.
const releaseRepo = "flamingo-stack/openframe-cli"

// localBinaryEnv, when set to the Windows path of a Linux openframe binary,
// installs that binary into WSL instead of downloading a release. Intended for
// dev/CI where there is no published release to fetch.
const localBinaryEnv = "OPENFRAME_WSL_BINARY"

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

// releaseTag maps a build version to its git tag. This repo tags releases with
// the BARE semver — release.yml runs `git tag -a "${VERSION}"`, so the tag for
// 0.4.7 is "0.4.7", not "v0.4.7" — and goreleaser's .Version strips a leading
// "v" anyway. A "v"-prefixed download URL 404s for every published release
// (T0-3), breaking WSL auto-install on first run of a released Windows binary.
func releaseTag(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
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
	// Probe WSL first. A missing WSL / missing distro returns actionable guidance
	// (fatal); "binary absent" (present=false, err=nil) means we may auto-install.
	present, err := wslBinaryStatus()
	if err != nil {
		return err
	}
	if present {
		return nil
	}

	// Dev/CI: install an explicit local Linux binary into WSL (no release needed).
	// When the caller opted in via OPENFRAME_WSL_BINARY, surface exactly why the
	// install/verify failed instead of the generic "not installed" guidance —
	// otherwise the real error (bad path, wrong arch, wslpath) is swallowed.
	if src := os.Getenv(localBinaryEnv); strings.TrimSpace(src) != "" {
		if ierr := installLocalBinaryInWSL(src); ierr != nil {
			return fmt.Errorf("%s=%s: %w", localBinaryEnv, src, ierr)
		}
		if verr := verifyOpenframeInWSL(); verr != nil {
			return fmt.Errorf("installed %s into WSL from %s but it is not runnable there — is it a linux/%s binary? (%w)", BinaryInWSL, src, goarch, verr)
		}
		return nil
	}

	if isReleaseVersion(version) {
		// Surface the real download/verify failure (404, checksum mismatch, no
		// curl in the distro) instead of swallowing it into generic guidance.
		if ierr := installOpenframeInWSL(version, goarch); ierr != nil {
			return fmt.Errorf("auto-installing openframe %s into WSL failed: %w", version, ierr)
		}
		if verr := verifyOpenframeInWSL(); verr != nil {
			return verr
		}
		return nil
	}

	return notInstalledError()
}

// shellSingleQuote safely single-quotes s for embedding in a bash script.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// localInstallScript returns a bash script (run inside WSL) that copies a Linux
// openframe binary — given by its Windows path — into ~/.openframe/bin. `wslpath`
// converts the Windows path to a WSL path. Pure and testable.
func localInstallScript(windowsPath string) string {
	return strings.Join([]string{
		"set -e",
		`BIN_DIR="$HOME/.openframe/bin"`,
		`mkdir -p "$BIN_DIR"`,
		`SRC="$(wslpath -u ` + shellSingleQuote(windowsPath) + `)"`,
		`install -m 0755 "$SRC" "$BIN_DIR/openframe"`,
	}, "\n")
}

// installLocalBinaryInWSL copies the Linux binary at the given Windows path into
// WSL. Thin exec wrapper around the tested localInstallScript.
func installLocalBinaryInWSL(windowsPath string) error {
	cmd := exec.Command("wsl", wslArgv("bash", "-lc", localInstallScript(windowsPath))...) // #nosec G204 -- path is single-quoted into a self-contained script
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("installing local openframe binary into WSL failed: %w\n%s", err, string(out))
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
	cmd := exec.Command("wsl", wslArgv("bash", "-lc", script)...) // #nosec G204 -- script built from constant-derived release URLs, no user input
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("installing openframe inside WSL failed: %w\n%s", err, string(out))
	}
	return nil
}
