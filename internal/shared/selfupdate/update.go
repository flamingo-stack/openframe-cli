package selfupdate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"golang.org/x/mod/semver"
)

// binaryName is the executable member packed inside each release archive
// (GoReleaser builds `binary: openframe`, wrap_in_directory: false, so the
// member sits at the archive root).
const binaryName = "openframe"

// archiveName returns the release archive filename for the given platform,
// matching GoReleaser's name_template (openframe-cli_<os>_<arch>.tar.gz).
func archiveName(goos, goarch string) string {
	return fmt.Sprintf("openframe-cli_%s_%s.tar.gz", goos, goarch)
}

// normalizeVersion returns a semver-comparable, v-prefixed version, or "" when
// v is not a comparable release version (e.g. the "dev" ldflags default or a
// malformed tag).
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "dev" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semver.IsValid(v) {
		return ""
	}
	return semver.Canonical(v)
}

// IsNewer reports whether latest is a strictly higher release than current.
// Either version being unparseable (dev build, malformed tag) yields false, so
// development builds never consider themselves out of date.
func IsNewer(current, latest string) bool {
	c, l := normalizeVersion(current), normalizeVersion(latest)
	if c == "" || l == "" {
		return false
	}
	return semver.Compare(l, c) > 0
}

// Status is the outcome of a version check.
type Status struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Available  bool   `json:"updateAvailable"`
	ReleaseURL string `json:"releaseUrl,omitempty"`
	DevBuild   bool   `json:"devBuild"`
}

// Updater checks for and applies self-updates.
type Updater struct {
	Current      string // running version (cmd.DefaultVersionInfo.Version)
	Client       Client
	GOOS, GOARCH string // default to runtime values; overridable in tests
	exePath      string // overrides the resolved executable path in tests
}

func (u Updater) goos() string {
	if u.GOOS != "" {
		return u.GOOS
	}
	return runtime.GOOS
}

func (u Updater) goarch() string {
	if u.GOARCH != "" {
		return u.GOARCH
	}
	return runtime.GOARCH
}

// Check queries a release and compares it to the running version. When tag is
// non-empty it targets that exact release instead of "latest".
func (u Updater) Check(ctx context.Context, tag string) (Status, Release, error) {
	var (
		rel Release
		err error
	)
	if tag != "" {
		rel, err = u.Client.ForTag(ctx, tag)
	} else {
		rel, err = u.Client.Latest(ctx)
	}
	if err != nil {
		return Status{}, Release{}, err
	}
	st := Status{
		Current:    u.Current,
		Latest:     rel.TagName,
		ReleaseURL: rel.HTMLURL,
		DevBuild:   normalizeVersion(u.Current) == "",
		Available:  IsNewer(u.Current, rel.TagName),
	}
	return st, rel, nil
}

// Apply downloads the target release's archive, verifies it against the release
// checksums, and atomically replaces the running executable. It keeps a backup
// and rolls back if the freshly installed binary fails a `--version` smoke
// test. progress, if non-nil, receives human-readable step messages.
func (u Updater) Apply(ctx context.Context, rel Release, progress func(string)) error {
	if u.goos() == "windows" {
		return fmt.Errorf("self-update is not supported for the native Windows launcher; " +
			"re-run the installer to update it (the WSL-side Linux binary updates itself)")
	}
	log := func(s string) {
		if progress != nil {
			progress(s)
		}
	}

	exePath, err := u.resolveExe()
	if err != nil {
		return err
	}
	if err := dirWritable(filepath.Dir(exePath)); err != nil {
		return fmt.Errorf("cannot update %s: %w", exePath, err)
	}

	name := archiveName(u.goos(), u.goarch())
	assetURL, ok := rel.assetURL(name)
	if !ok {
		return fmt.Errorf("release %s has no asset %s", rel.TagName, name)
	}
	log(fmt.Sprintf("Verifying checksum for %s...", name))
	sum, err := u.Client.fetchChecksum(ctx, rel, name)
	if err != nil {
		return err
	}

	// Stage the new binary next to the current one (same filesystem → atomic
	// rename), verified against the release checksum before it lands.
	newPath := exePath + ".new"
	defer func() { _ = os.Remove(newPath) }() // no-op once renamed into place
	log(fmt.Sprintf("Downloading %s %s...", binaryName, rel.TagName))
	dl := download.Downloader{}
	if err := dl.InstallVerifiedTarGz(ctx, download.PinnedAsset{URL: assetURL, SHA256: sum}, binaryName, newPath, 0o755); err != nil {
		return err
	}

	// Smoke-test the staged binary before committing to it.
	log("Verifying the downloaded binary...")
	if err := smokeTest(ctx, newPath); err != nil {
		return fmt.Errorf("the downloaded binary failed to run, keeping the current version: %w", err)
	}

	// Swap with a backup so a failed install can be rolled back.
	backup := exePath + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(exePath, backup); err != nil {
		return fmt.Errorf("backing up the current binary: %w", err)
	}
	if err := os.Rename(newPath, exePath); err != nil {
		_ = os.Rename(backup, exePath) // roll back
		return fmt.Errorf("installing the new binary (rolled back): %w", err)
	}
	_ = os.Remove(backup)
	log(fmt.Sprintf("Installed %s.", rel.TagName))
	return nil
}

// resolveExe returns the path of the binary to replace: the test override when
// set, otherwise the running executable with symlinks resolved.
func (u Updater) resolveExe() (string, error) {
	if u.exePath != "" {
		return u.exePath, nil
	}
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating the running binary: %w", err)
	}
	if resolved, rerr := filepath.EvalSymlinks(p); rerr == nil {
		p = resolved
	}
	return p, nil
}

// dirWritable reports whether files can be created in dir (i.e. the binary
// living there can be replaced) by probing with a temp file. This surfaces
// elevated / package-managed installs with a clear error instead of a mid-swap
// failure.
func dirWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".of-update-probe-*")
	if err != nil {
		return fmt.Errorf("%s is not writable (installed with elevated privileges or by a package manager?): %w", dir, err)
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

// smokeTest runs "<path> --version" and fails if the binary does not execute
// cleanly, so a corrupt or wrong-platform download is caught before the swap.
func smokeTest(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, path, "--version").CombinedOutput(); err != nil {
		return fmt.Errorf("%w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
