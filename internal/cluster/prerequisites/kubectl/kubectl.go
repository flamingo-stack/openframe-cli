package kubectl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type KubectlInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isKubectlInstalled() bool {
	// On Windows, check kubectl in WSL2
	if runtime.GOOS == "windows" {
		cmd := exec.Command("wsl", "-d", "Ubuntu", "command", "-v", "kubectl")
		return cmd.Run() == nil
	}

	if !commandExists("kubectl") {
		return false
	}
	// Check kubectl with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", "version", "--client")
	err := cmd.Run()
	return err == nil
}

func kubectlInstallHelp() string {
	switch runtime.GOOS {
	case "darwin":
		return "kubectl: Run 'brew install kubectl' or download from https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/"
	case "linux":
		return "kubectl: Install using your package manager or from https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/"
	case "windows":
		return "kubectl: Download from https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/"
	default:
		return "kubectl: Please install kubectl from https://kubernetes.io/docs/tasks/tools/"
	}
}

func NewKubectlInstaller() *KubectlInstaller {
	return &KubectlInstaller{}
}

func (k *KubectlInstaller) IsInstalled() bool {
	return isKubectlInstalled()
}

func (k *KubectlInstaller) GetInstallHelp() string {
	return kubectlInstallHelp()
}

func (k *KubectlInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return k.installMacOS()
	case "linux":
		return k.installLinux()
	case "windows":
		return k.installWindows()
	default:
		return fmt.Errorf("automatic kubectl installation not supported on %s", runtime.GOOS)
	}
}

func (k *KubectlInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("Homebrew is required for automatic kubectl installation on macOS. Please install brew first: https://brew.sh")
	}

	fmt.Println("Installing kubectl via Homebrew...")
	cmd := exec.Command("brew", "install", "kubectl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install kubectl: %w", err)
	}

	return nil
}

func (k *KubectlInstaller) installLinux() error {
	// Use direct binary download for all Linux distros - more reliable than package managers
	// which may have outdated repositories or require additional configuration
	return k.installBinary()
}

func (k *KubectlInstaller) installBinary() error {
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Use a shell script with retry logic and fallback version
	installCmd := fmt.Sprintf(`
FALLBACK_VERSION="v1.31.0"
ARCH="%s"

install_kubectl() {
    local max_retries=3
    local retry_delay=5

    for i in $(seq 1 $max_retries); do
        # Try to get the latest stable version
        local version
        version=$(curl -fsSL --retry 3 --retry-delay 2 https://dl.k8s.io/release/stable.txt 2>/dev/null || echo "")

        if [ -z "$version" ]; then
            version="$FALLBACK_VERSION"
        fi

        local download_url="https://dl.k8s.io/release/${version}/bin/linux/${ARCH}/kubectl"

        if curl -fsSL --retry 3 --retry-delay 2 -o /tmp/kubectl "$download_url" && \
           chmod +x /tmp/kubectl && \
           sudo mv /tmp/kubectl /usr/local/bin/kubectl; then
            return 0
        fi

        if [ $i -lt $max_retries ]; then
            sleep $retry_delay
            retry_delay=$((retry_delay * 2))
        fi
    done

    return 1
}

install_kubectl
`, arch)

	if err := k.runShellCommand(installCmd); err != nil {
		return fmt.Errorf("failed to install kubectl via binary download: %w", err)
	}

	return nil
}

func (k *KubectlInstaller) installWindows() error {
	fmt.Println("Installing kubectl inside WSL2...")

	// Install kubectl inside WSL2 Ubuntu
	installScript := `#!/bin/bash
set -e

# Check if kubectl is already installed
if command -v kubectl &> /dev/null; then
    echo "kubectl already installed in WSL2"
    exit 0
fi

echo "Installing kubectl..."

# Download the latest stable kubectl binary (silent mode to avoid progress output)
curl -fsSLO "https://dl.k8s.io/release/$(curl -fsSL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Install kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Clean up
rm kubectl

echo "kubectl installed successfully"
`

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install kubectl in WSL2: %w", err)
	}

	// Create Windows wrapper
	if err := k.createKubectlWrapper(); err != nil {
		return fmt.Errorf("failed to create kubectl wrapper: %w", err)
	}

	fmt.Println("✓ kubectl installed successfully in WSL2!")
	return nil
}

func (k *KubectlInstaller) createKubectlWrapper() error {
	fmt.Println("Creating kubectl command for Windows...")

	// First, create a bash helper script in WSL2 that converts Windows paths
	helperScript := `#!/bin/bash
# Helper script to run kubectl with Windows path conversion

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

# Execute kubectl with converted arguments
exec kubectl "${args[@]}"
`

	// Write the helper script to WSL2 (write to temp location first, then move with sudo)
	writeCmd := fmt.Sprintf(`
cat > /tmp/kubectl-wrapper.sh << 'EOFSCRIPT'
%s
EOFSCRIPT
sudo mv /tmp/kubectl-wrapper.sh /usr/local/bin/kubectl-wrapper.sh
sudo chmod +x /usr/local/bin/kubectl-wrapper.sh
`, helperScript)

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", writeCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create kubectl helper script in WSL2: %w", err)
	}

	// Create a batch file wrapper that calls the helper script
	wrapperDir := os.Getenv("USERPROFILE") + "\\bin"
	os.MkdirAll(wrapperDir, 0755)

	wrapperPath := wrapperDir + "\\kubectl.bat"

	// Simple batch wrapper that calls the bash helper
	wrapperContent := `@echo off
wsl -d Ubuntu /usr/local/bin/kubectl-wrapper.sh %*
`

	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {
		return fmt.Errorf("failed to create kubectl wrapper: %w", err)
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

	pathCmd := exec.Command("powershell", "-Command", addPathScript)
	pathCmd.Stdout = os.Stdout
	pathCmd.Stderr = os.Stderr
	pathCmd.Run() // Ignore errors

	// Update PATH for current process so kubectl can be found immediately
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, wrapperDir) {
		newPath := currentPath + ";" + wrapperDir
		os.Setenv("PATH", newPath)
		fmt.Printf("Updated current process PATH to include: %s\n", wrapperDir)
	}

	fmt.Printf("✓ kubectl wrapper created at: %s\n", wrapperPath)
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

func (k *KubectlInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Completely silence output during installation
	return cmd.Run()
}

func (k *KubectlInstaller) runShellCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	// Completely silence output during installation
	return cmd.Run()
}