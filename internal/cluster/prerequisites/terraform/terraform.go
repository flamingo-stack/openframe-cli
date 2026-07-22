package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/pterm/pterm"
)

// TerraformInstaller installs the Terraform CLI used by the cloud cluster
// providers. Unlike docker/k3d/helm it always installs the pinned, verified
// binary into ~/.openframe/bin — package managers no longer carry current
// Terraform (the homebrew-core formula is disabled since the BUSL change), so
// the verified download is the primary path, not the fallback.
type TerraformInstaller struct{}

func NewTerraformInstaller() *TerraformInstaller {
	return &TerraformInstaller{}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Minimum terraform version the generated root modules require
// (required_version = ">= 1.15.0" in the templates). An older system
// terraform must count as NOT installed, so the pinned binary gets installed
// into ~/.openframe/bin — which then wins PATH resolution.
const minMajor, minMinor = 1, 15

// IsInstalled reports whether a terraform binary that satisfies the
// templates' version constraint is reachable. The CLI-managed bin dir is
// prepended to PATH so a previously installed pinned binary is found even in
// a fresh shell.
func (t *TerraformInstaller) IsInstalled() bool {
	if binDir, err := download.UserBinDir(); err == nil {
		download.PrependToPath(binDir)
	}
	if !commandExists("terraform") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "terraform", "version", "-json").Output()
	if err != nil {
		return false
	}
	var v struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return false
	}
	return versionSatisfies(v.TerraformVersion, minMajor, minMinor)
}

// versionSatisfies reports whether version ("1.15.8") is >= major.minor.
// Unparseable versions are treated as too old — the safe direction, since it
// triggers a pinned install rather than a mid-provisioning failure.
func versionSatisfies(version string, major, minor int) bool {
	parts := strings.SplitN(strings.TrimPrefix(version, "v"), ".", 3)
	if len(parts) < 2 {
		return false
	}
	gotMajor, err1 := strconv.Atoi(parts[0])
	gotMinor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return gotMajor > major || (gotMajor == major && gotMinor >= minor)
}

func (t *TerraformInstaller) GetInstallHelp() string {
	return platform.InstallHint("terraform")
}

// Install downloads the pinned Terraform release, verifies its SHA256, and
// installs it into ~/.openframe/bin (no sudo). Unlike docker/k3d (which run
// inside WSL on Windows), terraform is installed natively on all three
// platforms — the provisioning engine invokes it directly.
func (t *TerraformInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		return t.installVerified()
	default:
		return fmt.Errorf("automatic terraform installation not supported on %s", runtime.GOOS)
	}
}

func (t *TerraformInstaller) installVerified() error {
	binDir, err := download.UserBinDir()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Printf("Downloading verified terraform %s...\n", download.Terraform.Version)
	path, err := (download.Downloader{}).InstallPinnedTool(ctx, download.Terraform, binDir)
	if err != nil {
		return fmt.Errorf("verified terraform install failed: %w", err)
	}

	download.PrependToPath(binDir)
	pterm.Success.Printf("Installed verified terraform %s to %s\n", download.Terraform.Version, path)
	pterm.Info.Printf("To use terraform directly in your shell, add %s to PATH: export PATH=\"%s:$PATH\"\n", binDir, binDir)
	return nil
}
