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

// TestReleaseTag locks the tag convention of this repo: release.yml tags with
// the BARE semver (`git tag -a "${VERSION}"` → "0.4.7"). A "v"-prefixed tag
// produced a download URL that 404s for every published release (T0-3).
func TestReleaseTag(t *testing.T) {
	if got := releaseTag("1.2.3"); got != "1.2.3" {
		t.Errorf("tag = %q, want 1.2.3 (bare, matching release.yml tagging)", got)
	}
	if got := releaseTag("v1.2.3"); got != "1.2.3" {
		t.Errorf("v-prefixed version tag = %q, want 1.2.3", got)
	}
	if got := releaseTag(" 0.4.7 "); got != "0.4.7" {
		t.Errorf("untrimmed version tag = %q, want 0.4.7", got)
	}
}

func TestLinuxArchiveName(t *testing.T) {
	if got := linuxArchiveName("amd64"); got != "openframe-cli_linux_amd64.tar.gz" {
		t.Errorf("amd64 archive = %q", got)
	}
	if got := linuxArchiveName("arm64"); got != "openframe-cli_linux_arm64.tar.gz" {
		t.Errorf("arm64 archive = %q", got)
	}
}

func TestReleaseAssetURL(t *testing.T) {
	// Real releases live at .../download/<bare-version>/... (e.g. 0.4.7); the
	// v-prefixed form 404s.
	got := releaseAssetURL("1.2.3", "openframe-cli_linux_amd64.tar.gz")
	want := "https://github.com/flamingo-stack/openframe-cli/releases/download/1.2.3/openframe-cli_linux_amd64.tar.gz"
	if got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
	if csum := releaseAssetURL("2.0.0", "checksums.txt"); !strings.HasSuffix(csum, "/2.0.0/checksums.txt") {
		t.Errorf("checksums url = %q", csum)
	}
}

// TestInstallScript_VerifiesAndInstalls locks the safety-critical shape of the
// install script: it must download, verify the SHA256 against the release
// checksums, and only then install the binary.
func TestInstallScript_VerifiesAndInstalls(t *testing.T) {
	s := installScript(
		"https://example.com/archive.tar.gz",
		"https://example.com/checksums.txt",
		"openframe-cli_linux_amd64.tar.gz",
	)
	for _, want := range []string{
		"set -e",
		`curl -fsSL -o archive.tar.gz "https://example.com/archive.tar.gz"`,
		`curl -fsSL -o checksums.txt "https://example.com/checksums.txt"`,
		"sha256sum archive.tar.gz",
		`grep " openframe-cli_linux_amd64.tar.gz$" checksums.txt`,
		`[ -n "$EXPECTED" ] && [ "$EXPECTED" = "$ACTUAL" ]`,
		`install -m 0755 openframe "$BIN_DIR/openframe"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("install script missing %q:\n%s", want, s)
		}
	}

	// The install must come AFTER the checksum check (never install unverified).
	verifyIdx := strings.Index(s, "sha256sum")
	installIdx := strings.Index(s, "install -m 0755")
	if verifyIdx < 0 || installIdx < 0 || verifyIdx > installIdx {
		t.Error("checksum verification must precede install")
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
