// Package infracost installs the OPTIONAL infracost CLI, which powers the
// monthly cost estimate in the cloud dry-run preview. It is never a required
// prerequisite: the CLI only OFFERS to install it, and estimates additionally
// need the user to run `infracost auth login` once (free API key for
// infracost's own pricing API — no cloud credentials are involved).
package infracost

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/pterm/pterm"
)

type Installer struct{}

func NewInstaller() *Installer { return &Installer{} }

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// IsInstalled reports whether a working infracost binary is reachable,
// preferring the CLI-managed bin dir.
func (i *Installer) IsInstalled() bool {
	if binDir, err := download.UserBinDir(); err == nil {
		download.PrependToPath(binDir)
	}
	if !commandExists("infracost") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "infracost", "--version").Run() == nil
}

func (i *Installer) GetInstallHelp() string {
	return platform.InstallHint("infracost")
}

// Install downloads the pinned infracost release, verifies its SHA256, and
// installs it into ~/.openframe/bin. The archive member layout is irregular
// (see the pin comment), hence the direct InstallVerifiedTarGz call.
func (i *Installer) Install() error {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
	default:
		return fmt.Errorf("automatic infracost installation not supported on %s", runtime.GOOS)
	}

	asset, ok := download.Infracost.Asset(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return fmt.Errorf("no verified infracost %s asset for %s/%s", download.Infracost.Version, runtime.GOOS, runtime.GOARCH)
	}
	binDir, err := download.UserBinDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		return fmt.Errorf("creating %s: %w", binDir, err)
	}

	member := fmt.Sprintf("infracost-%s-%s", runtime.GOOS, runtime.GOARCH)
	dest := filepath.Join(binDir, "infracost")
	if runtime.GOOS == "windows" {
		member = "infracost.exe"
		dest += ".exe"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Printf("Downloading verified infracost %s...\n", download.Infracost.Version)
	if err := (download.Downloader{}).InstallVerifiedTarGz(ctx, asset, member, dest, 0o750); err != nil {
		return fmt.Errorf("verified infracost install failed: %w", err)
	}
	download.PrependToPath(binDir)
	pterm.Success.Printf("Installed verified infracost %s to %s\n", download.Infracost.Version, dest)
	pterm.Info.Println("To enable estimates, get a free API key once: infracost auth login")
	return nil
}
