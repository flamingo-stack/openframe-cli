package wsllauncher

import (
	"strings"
	"testing"
)

func TestIsReleaseVersion(t *testing.T) {
	release := []string{"1.2.3", "0.1.0", "v2.0.0", "10.4.1"}
	notRelease := []string{"", "dev", "none", "unknown", "snapshot", "0.0.0-dev.t123.abc", "1.2.3-rc1"}

	for _, v := range release {
		if !isReleaseVersion(v) {
			t.Errorf("%q should be a release version", v)
		}
	}
	for _, v := range notRelease {
		if isReleaseVersion(v) {
			t.Errorf("%q should NOT be a release version", v)
		}
	}
}

// TestStdinInstallScript locks the safety-critical shape of the WSL install:
// the binary arrives VERIFIED on stdin (cosign + SHA256, done on the Windows
// side by selfupdate.FetchVerifiedLinuxBinary) — the script itself must not
// download anything, and must install atomically (write tmp, chmod, rename).
func TestStdinInstallScript(t *testing.T) {
	s := stdinInstallScript()
	for _, want := range []string{
		"set -e",
		`mkdir -p "$BIN_DIR"`,
		`cat > "$BIN_DIR/openframe.tmp"`,
		`chmod 0755 "$BIN_DIR/openframe.tmp"`,
		`mv "$BIN_DIR/openframe.tmp" "$BIN_DIR/openframe"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("stdin install script missing %q:\n%s", want, s)
		}
	}
	// No network access from inside WSL: the old curl-based path checked only
	// checksums.txt from the same release (no authenticity).
	for _, banned := range []string{"curl", "wget", "https://"} {
		if strings.Contains(s, banned) {
			t.Errorf("stdin install script must not download anything, found %q:\n%s", banned, s)
		}
	}
	// Self-contained: no positional args to bash (avoids the wsl.exe arg bug).
	if strings.Contains(s, `"$@"`) || strings.Contains(s, `"$1"`) {
		t.Error("stdin install script must be self-contained (no positional args)")
	}
}

func TestShellSingleQuote(t *testing.T) {
	cases := map[string]string{
		`C:\bin\openframe`:  `'C:\bin\openframe'`,
		`C:\it's\openframe`: `'C:\it'\''s\openframe'`,
		`plain`:             `'plain'`,
	}
	for in, want := range cases {
		if got := shellSingleQuote(in); got != want {
			t.Errorf("shellSingleQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestLocalInstallScript locks the dev/CI local-binary install: convert the
// Windows path with wslpath, then install into ~/.openframe/bin.
func TestLocalInstallScript(t *testing.T) {
	s := localInstallScript(`C:\Users\ci\openframe-linux-amd64`)
	for _, want := range []string{
		"set -e",
		`wslpath -u 'C:\Users\ci\openframe-linux-amd64'`,
		`install -m 0755 "$SRC" "$BIN_DIR/openframe"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("local install script missing %q:\n%s", want, s)
		}
	}
	// Self-contained: no positional args to bash (avoids the wsl.exe arg bug).
	if strings.Contains(s, `"$@"`) || strings.Contains(s, `"$1"`) {
		t.Error("local install script must be self-contained (no positional args)")
	}
}
