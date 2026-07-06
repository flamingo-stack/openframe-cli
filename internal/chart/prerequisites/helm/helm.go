package helm

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
	case "windows":
		return h.installWindows()
	default:
		return fmt.Errorf("automatic Helm installation not supported on %s", runtime.GOOS)
	}
}

func (h *HelmInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic Helm installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "helm")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Helm: %w", err)
	}

	return nil
}

func (h *HelmInstaller) installLinux() error {
	return h.installVerified()
}

// installVerified downloads the pinned Helm .tar.gz, verifies its SHA256, extracts
// the helm binary, and installs it into ~/.openframe/bin (no sudo). This replaces
// the unverified apt/`curl get-helm-3 | bash` installs (audit T0.3).
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

func (h *HelmInstaller) installWindows() error {
	fmt.Println("Installing Helm inside WSL2...")

	// Install Helm inside WSL2 Ubuntu using the official install script
	installScript := `#!/bin/bash
set -e

# Check if helm is already installed
if command -v helm &> /dev/null; then
    echo "Helm already installed in WSL2"
    exit 0
fi

echo "Installing Helm..."

# Use the official Helm install script (redirect stderr to suppress progress output)
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash 2>/dev/null || curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

echo "Helm installed successfully"
`

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Helm in WSL2: %w", err)
	}

	// Create Windows wrapper
	if err := h.createHelmWrapper(); err != nil {
		return fmt.Errorf("failed to create Helm wrapper: %w", err)
	}

	fmt.Println("✓ Helm installed successfully in WSL2!")
	return nil
}

func (h *HelmInstaller) createHelmWrapper() error {
	fmt.Println("Creating helm command for Windows...")

	// First, create a bash helper script in WSL2 that converts Windows paths
	helperScript := `#!/bin/bash
# Helper script to run helm with Windows path conversion

# Set Helm environment variables to use writable directories
# This is especially important in CI environments where home directory may not have write permissions
export HELM_CACHE_HOME="/tmp/helm/cache"
export HELM_CONFIG_HOME="/tmp/helm/config"
export HELM_DATA_HOME="/tmp/helm/data"

# Create directories if they don't exist
mkdir -p "$HELM_CACHE_HOME" "$HELM_CONFIG_HOME" "$HELM_DATA_HOME"

args=()
for arg in "$@"; do
    # Check if argument looks like a Windows path (contains : after first char)
    if [[ "$arg" =~ ^[A-Za-z]: ]]; then
        # Convert Windows path to WSL path
        converted=$(wslpath -a "$arg" 2>/dev/null || echo "$arg")
        args+=("$converted")
    else
        args+=("$arg")
    fi
done

# Execute helm with converted arguments
exec helm "${args[@]}"
`

	// Write the helper script to WSL2 (write to temp location first, then move with sudo)
	writeCmd := fmt.Sprintf(`
cat > /tmp/helm-wrapper.sh << 'EOFSCRIPT'
%s
EOFSCRIPT
sudo mv /tmp/helm-wrapper.sh /usr/local/bin/helm-wrapper.sh
sudo chmod +x /usr/local/bin/helm-wrapper.sh
`, helperScript)

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", writeCmd) // #nosec G204 -- shell string built from constant/program-derived values, not untrusted input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create helm helper script in WSL2: %w", err)
	}

	// Create a batch file wrapper that calls the helper script
	wrapperDir := os.Getenv("USERPROFILE") + "\\bin"
	_ = os.MkdirAll(wrapperDir, 0750) // #nosec G703 -- wrapper dir path from USERPROFILE env + constant name, runs as invoking user

	wrapperPath := wrapperDir + "\\helm.bat"

	// Simple batch wrapper that calls the bash helper
	wrapperContent := `@echo off
wsl -d Ubuntu /usr/local/bin/helm-wrapper.sh %*
`

	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil { // #nosec G306 G703 -- wrapper path from USERPROFILE env + constant name; script must be executable
		return fmt.Errorf("failed to create helm wrapper: %w", err)
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

	pathCmd := exec.Command("powershell", "-Command", addPathScript) // #nosec G204 G702 -- shell string built from constant/program-derived values, not untrusted input
	pathCmd.Stdout = os.Stdout
	pathCmd.Stderr = os.Stderr
	_ = pathCmd.Run() // Ignore errors

	// Update PATH for current process so helm can be found immediately
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, wrapperDir) {
		newPath := currentPath + ";" + wrapperDir
		_ = os.Setenv("PATH", newPath)
		fmt.Printf("Updated current process PATH to include: %s\n", wrapperDir)
	}

	fmt.Printf("✓ Helm wrapper created at: %s\n", wrapperPath)
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
