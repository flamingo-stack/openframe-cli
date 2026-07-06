package helm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/flamingo-stack/openframe-cli/internal/shared/wsllauncher"
)

type HelmInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isHelmInstalled() bool {
	// On Windows, check helm in WSL2
	if runtime.GOOS == "windows" {
		return wsllauncher.CommandAvailable("helm")
	}

	if !commandExists("helm") {
		return false
	}
	// Check helm with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "helm", "version")
	err := cmd.Run()
	return err == nil
}

func helmInstallHelp() string {
	return platform.InstallHint("helm")
}

func NewHelmInstaller() *HelmInstaller {
	return &HelmInstaller{}
}

func (h *HelmInstaller) IsInstalled() bool {
	return isHelmInstalled()
}

func (h *HelmInstaller) GetInstallHelp() string {
	return helmInstallHelp()
}

func (h *HelmInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return h.installMacOS()
	case "linux":
		return h.installLinux()
	default:
		// Windows is unsupported here by design: the CLI forwards into WSL and
		// runs as linux, so native-Windows install code is never reached.
		return fmt.Errorf("automatic helm installation not supported on %s", runtime.GOOS)
	}
}

func (h *HelmInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic helm installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}

	fmt.Println("Installing helm via Homebrew...")
	cmd := exec.Command("brew", "install", "helm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install helm: %w", err)
	}

	return nil
}

func (h *HelmInstaller) installLinux() error {
	return h.installVerified()
}

// installVerified downloads the pinned Helm .tar.gz, verifies its SHA256, extracts
// the helm binary, and installs it into ~/.openframe/bin (no sudo). This replaces
// the unverified `curl get-helm-3 | bash` install (audit T0.3).
func (h *HelmInstaller) installVerified() error {
	binDir, err := download.UserBinDir()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Printf("Downloading verified helm %s...\n", download.Helm.Version)
	path, err := (download.Downloader{}).InstallPinnedTool(ctx, download.Helm, binDir)
	if err != nil {
		return fmt.Errorf("verified helm install failed: %w", err)
	}
	download.PrependToPath(binDir)
	fmt.Printf("Installed verified helm %s to %s\n", download.Helm.Version, path)
	return nil
}
