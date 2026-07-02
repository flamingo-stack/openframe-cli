package k3d

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
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
		cmd := exec.Command("wsl", "-d", "Ubuntu", "command", "-v", "k3d")
		return cmd.Run() == nil
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
	case "windows":
		return k.installWindows()
	default:
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

func (k *K3dInstaller) installWindows() error {
	fmt.Println("Installing k3d inside WSL2...")

	// Install k3d inside WSL2 Ubuntu using a script with retry logic and fallback version
	// The official install script can fail with 504 errors from GitHub
	installScript := `#!/bin/bash
set -e

# Check if k3d is already installed
if command -v k3d &> /dev/null; then
    echo "k3d already installed in WSL2"
    exit 0
fi

echo "Installing k3d..."

# Fallback version if we can't fetch latest from GitHub
FALLBACK_VERSION="v5.7.5"

# Function to install k3d with retries
install_k3d() {
    local max_retries=3
    local retry_delay=5

    for i in $(seq 1 $max_retries); do
        echo "Attempt $i of $max_retries..."

        # Try the official install script first
        if curl -fsSL --retry 3 --retry-delay 2 https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash 2>/dev/null; then
            return 0
        fi

        echo "Official install script failed, trying direct binary download..."

        # Try direct binary download with specific version as fallback
        local version
        version=$(curl -fsSL --retry 3 --retry-delay 2 https://api.github.com/repos/k3d-io/k3d/releases/latest 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

        if [ -z "$version" ]; then
            echo "Could not fetch latest version, using fallback: $FALLBACK_VERSION"
            version="$FALLBACK_VERSION"
        fi

        local arch="amd64"
        if [ "$(uname -m)" = "aarch64" ]; then
            arch="arm64"
        fi

        local download_url="https://github.com/k3d-io/k3d/releases/download/${version}/k3d-linux-${arch}"
        echo "Downloading k3d ${version} for ${arch}..."

        if curl -fsSL --retry 3 --retry-delay 2 -o /tmp/k3d "$download_url" && chmod +x /tmp/k3d && sudo mv /tmp/k3d /usr/local/bin/k3d; then
            return 0
        fi

        if [ $i -lt $max_retries ]; then
            echo "Retrying in ${retry_delay} seconds..."
            sleep $retry_delay
            retry_delay=$((retry_delay * 2))
        fi
    done

    return 1
}

if install_k3d; then
    echo "k3d installed successfully"
else
    echo "Failed to install k3d after multiple attempts"
    exit 1
fi
`

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install k3d in WSL2: %w", err)
	}

	// Create Windows wrapper
	if err := k.createK3dWrapper(); err != nil {
		return fmt.Errorf("failed to create k3d wrapper: %w", err)
	}

	fmt.Println("✓ k3d installed successfully in WSL2!")
	return nil
}

func (k *K3dInstaller) createK3dWrapper() error {
	fmt.Println("Creating k3d command for Windows...")

	// Create a batch file wrapper that calls k3d in WSL2
	wrapperDir := os.Getenv("USERPROFILE") + "\\bin"
	_ = os.MkdirAll(wrapperDir, 0750) // #nosec G703 -- wrapper dir path from USERPROFILE env + constant name, runs as invoking user

	wrapperPath := wrapperDir + "\\k3d.bat"
	wrapperContent := `@echo off
wsl -d Ubuntu k3d %*
`

	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil { // #nosec G306 G703 -- wrapper path from USERPROFILE env + constant name; script must be executable
		return fmt.Errorf("failed to create k3d wrapper: %w", err)
	}

	// Add to PATH if not already there
	addPathScript := fmt.Sprintf(`
$binDir = "%s"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
    $env:Path = "$env:Path;$binDir"
    Write-Host "Added $binDir to PATH"
} else {
    Write-Host "PATH already contains $binDir"
}
`, wrapperDir)

	cmd := exec.Command("powershell", "-Command", addPathScript) // #nosec G204 G702 -- shell string built from constant/program-derived values, not untrusted input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // Ignore errors

	// Update PATH for current process so k3d can be found immediately
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, wrapperDir) {
		newPath := currentPath + ";" + wrapperDir
		_ = os.Setenv("PATH", newPath)
		fmt.Printf("Updated current process PATH to include: %s\n", wrapperDir)
	}

	fmt.Printf("✓ k3d wrapper created at: %s\n", wrapperPath)
	return nil
}

// containsPath checks if a PATH string contains a specific directory
func containsPath(pathEnv, dir string) bool {
	paths := strings.Split(pathEnv, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), strings.TrimSpace(dir)) {
			return true
		}
	}
	return false
}

func (k *K3dInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
	// Completely silence output during installation
	return cmd.Run()
}

