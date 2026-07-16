package wsllauncher

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/selfupdate"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
)

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

// stdinInstallScript returns a bash script (run inside WSL) that installs the
// openframe binary streamed on STDIN into ~/.openframe/bin. The binary itself
// is downloaded and verified (cosign identity + SHA256) on the Windows side by
// selfupdate.FetchVerifiedLinuxBinary — nothing unverified ever reaches WSL,
// and the distro needs no curl. Pure and testable.
func stdinInstallScript() string {
	return strings.Join([]string{
		"set -e",
		`BIN_DIR="$HOME/.openframe/bin"`,
		`mkdir -p "$BIN_DIR"`,
		`cat > "$BIN_DIR/openframe.tmp"`,
		`chmod 0755 "$BIN_DIR/openframe.tmp"`,
		`mv "$BIN_DIR/openframe.tmp" "$BIN_DIR/openframe"`,
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

// installOpenframeInWSL fetches the matching Linux release through the full
// self-update trust chain (cosign-verified checksums, SHA256-verified archive)
// on the Windows side, then streams the binary into WSL via stdin. The old
// curl-inside-WSL path verified only checksums.txt from the same release —
// no authenticity — so a compromised release upload meant arbitrary code
// execution in WSL while `openframe update` would have rejected the same
// binary (audit B5/T2).
func installOpenframeInWSL(version, goarch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// This is the first thing a Windows user ever sees, and it downloads several
	// megabytes and runs a cosign verification before their command starts.
	// FetchVerifiedLinuxBinary already narrates each step; the caller passed nil
	// and dropped every line, so the first run looked frozen.
	sp := spinner.Start("Preparing the OpenFrame Linux binary for WSL...")
	binary, err := selfupdate.FetchVerifiedLinuxBinary(ctx, version, goarch, sp.UpdateText)
	if err != nil {
		sp.Fail("Could not fetch the OpenFrame Linux binary")
		return err
	}

	sp.UpdateText("Installing openframe inside WSL...")
	cmd := exec.Command("wsl", wslArgv("bash", "-lc", stdinInstallScript())...) // #nosec G204 -- constant script, binary delivered via stdin
	cmd.Stdin = bytes.NewReader(binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		sp.Fail("Installing openframe inside WSL failed")
		return fmt.Errorf("installing openframe inside WSL failed: %w\n%s", err, string(out))
	}
	sp.Success("OpenFrame is installed inside WSL")
	return nil
}
