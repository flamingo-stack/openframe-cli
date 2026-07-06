package k3d

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/flamingo-stack/openframe-cli/internal/shared/wsllauncher"
	"github.com/pterm/pterm"
)

type K3dInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isK3dInstalled() bool {
	// On Windows, check k3d in WSL2
	if runtime.GOOS == "windows" {
		return wsllauncher.CommandAvailable("k3d")
	}

	if !commandExists("k3d") {
		return false
	}
	// Check k3d with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "k3d", "version")
	err := cmd.Run()
	return err == nil
}

func k3dInstallHelp() string {
	return platform.InstallHint("k3d")
}

func NewK3dInstaller() *K3dInstaller {
	return &K3dInstaller{}
}

func (k *K3dInstaller) IsInstalled() bool {
	return isK3dInstalled()
}

func (k *K3dInstaller) GetInstallHelp() string {
	return k3dInstallHelp()
}

func (k *K3dInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return k.installMacOS()
	case "linux":
		return k.installLinux()
	default:
		// Windows is unsupported here by design: the CLI forwards into WSL and
		// runs as linux, so native-Windows install code is never reached.
		return fmt.Errorf("automatic k3d installation not supported on %s", runtime.GOOS)
	}
}

func (k *K3dInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic k3d installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "k3d")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install k3d: %w", err)
	}

	return nil
}

func (k *K3dInstaller) installLinux() error {
	if commandExists("apt") {
		return k.installUbuntu()
	} else if commandExists("yum") {
		return k.installRedHat()
	} else if commandExists("dnf") {
		return k.installFedora()
	} else if commandExists("pacman") {
		return k.installArch()
	} else {
		return k.installVerified()
	}
}

func (k *K3dInstaller) installUbuntu() error {
	// k3d doesn't have official apt repository, so use the install script
	return k.installVerified()
}

func (k *K3dInstaller) installRedHat() error {
	// k3d doesn't have official yum repository, so use the install script
	return k.installVerified()
}

func (k *K3dInstaller) installFedora() error {
	// k3d doesn't have official dnf repository, so use the install script
	return k.installVerified()
}

func (k *K3dInstaller) installArch() error {
	// Try AUR package first, fall back to script
	if commandExists("yay") {
		if err := k.runCommand("yay", "-S", "--noconfirm", "k3d-bin"); err == nil {
			return nil
		}
	}

	if commandExists("paru") {
		if err := k.runCommand("paru", "-S", "--noconfirm", "k3d-bin"); err == nil {
			return nil
		}
	}

	// Fall back to install script
	return k.installVerified()
}

// installVerified downloads the pinned k3d binary, verifies its SHA256, and
// installs it into the CLI-managed user bin directory (~/.openframe/bin) with
// no sudo. This replaces the previous unverified "curl | bash" / "curl -o
// /tmp/k3d && sudo mv" install (audit I5/M1).
func (k *K3dInstaller) installVerified() error {
	binDir, err := download.UserBinDir()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Printf("Downloading verified k3d %s...\n", download.K3d.Version)
	path, err := (download.Downloader{}).InstallPinnedTool(ctx, download.K3d, binDir)
	if err != nil {
		return fmt.Errorf("verified k3d install failed: %w", err)
	}

	download.PrependToPath(binDir)
	pterm.Success.Printf("Installed verified k3d %s to %s\n", download.K3d.Version, path)
	pterm.Info.Printf("To use k3d directly in your shell, add %s to PATH: export PATH=\"%s:$PATH\"\n", binDir, binDir)
	return nil
}

func (k *K3dInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
	// Completely silence output during installation
	return cmd.Run()
}
