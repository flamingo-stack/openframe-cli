package kubectl

import (
	"context"
	"encoding/base64"
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
	if commandExists("apt") {
		return k.installUbuntu()
	} else if commandExists("yum") {
		return k.installRedHat()
	} else if commandExists("dnf") {
		return k.installFedora()
	} else if commandExists("pacman") {
		return k.installArch()
	} else {
		return k.installBinary()
	}
}

func (k *KubectlInstaller) installUbuntu() error {
	fmt.Println("Installing kubectl on Ubuntu/Debian...")
	
	commands := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl"},
	}

	for _, cmdArgs := range commands {
		if err := k.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add Kubernetes GPG key
	gpgCmd := "curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-archive-keyring.gpg"
	if err := k.runShellCommand(gpgCmd); err != nil {
		return fmt.Errorf("failed to add Kubernetes GPG key: %w", err)
	}

	// Add Kubernetes repository
	repoCmd := `echo "deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list`
	if err := k.runShellCommand(repoCmd); err != nil {
		return fmt.Errorf("failed to add Kubernetes repository: %w", err)
	}

	// Install kubectl
	installCommands := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "-y", "kubectl"},
	}

	for _, cmdArgs := range installCommands {
		if err := k.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	return nil
}

func (k *KubectlInstaller) installRedHat() error {
	fmt.Println("Installing kubectl on CentOS/RHEL...")
	
	// Create Kubernetes repository
	repoContent := `[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl`

	repoCmd := fmt.Sprintf("echo '%s' | sudo tee /etc/yum.repos.d/kubernetes.repo", repoContent)
	if err := k.runShellCommand(repoCmd); err != nil {
		return fmt.Errorf("failed to add Kubernetes repository: %w", err)
	}

	// Install kubectl
	if err := k.runCommand("sudo", "yum", "install", "-y", "kubectl", "--disableexcludes=kubernetes"); err != nil {
		return fmt.Errorf("failed to install kubectl: %w", err)
	}

	return nil
}

func (k *KubectlInstaller) installFedora() error {
	fmt.Println("Installing kubectl on Fedora...")
	
	// Create Kubernetes repository
	repoContent := `[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl`

	repoCmd := fmt.Sprintf("echo '%s' | sudo tee /etc/yum.repos.d/kubernetes.repo", repoContent)
	if err := k.runShellCommand(repoCmd); err != nil {
		return fmt.Errorf("failed to add Kubernetes repository: %w", err)
	}

	// Install kubectl
	if err := k.runCommand("sudo", "dnf", "install", "-y", "kubectl", "--disableexcludes=kubernetes"); err != nil {
		return fmt.Errorf("failed to install kubectl: %w", err)
	}

	return nil
}

func (k *KubectlInstaller) installArch() error {
	fmt.Println("Installing kubectl on Arch Linux...")
	
	if err := k.runCommand("sudo", "pacman", "-S", "--noconfirm", "kubectl"); err != nil {
		return fmt.Errorf("failed to install kubectl: %w", err)
	}

	return nil
}

func (k *KubectlInstaller) installBinary() error {
	fmt.Println("Installing kubectl via direct binary download...")
	
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	commands := []string{
		fmt.Sprintf("curl -LO \"https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/%s/kubectl\"", arch),
		"sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl",
		"rm kubectl",
	}

	for _, cmd := range commands {
		if err := k.runShellCommand(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmd, err)
		}
	}

	return nil
}

func (k *KubectlInstaller) installWindows() error {
	fmt.Println("Installing kubectl inside WSL2...")

	// Install kubectl inside WSL2 Ubuntu
	// NOTE: We do NOT use 'set -e' here because WSL2 DNS can be unreliable
	// and we want to retry downloads with proper error handling
	//
	// IMPORTANT: The script is base64-encoded before passing to WSL to avoid
	// shell character interpretation issues. When passing multi-line scripts
	// via 'wsl bash -c <script>', special characters like >, &, | can be
	// interpreted by Windows command line parsing before reaching bash.
	installScript := `#!/bin/bash

# Check if kubectl is already installed
if command -v kubectl >/dev/null 2>&1; then
    echo "kubectl already installed in WSL2"
    exit 0
fi

echo "Installing kubectl..."

# Wait for DNS to be available (WSL2 networking can take time to stabilize)
echo "Waiting for DNS to be available..."
DNS_READY=0
for i in $(seq 1 30); do
    if nslookup dl.k8s.io >/dev/null 2>&1; then
        echo "DNS is ready"
        DNS_READY=1
        break
    fi
    echo "DNS not ready, waiting... (attempt $i/30)"
    sleep 2
done

if [ "$DNS_READY" = "0" ]; then
    echo "ERROR: DNS resolution failed after 60 seconds"
    echo "WSL2 networking may not be properly configured"
    echo "Try running: wsl --shutdown and then restart WSL"
    exit 6
fi

# Download kubectl with retries
DOWNLOAD_OK=0
for i in $(seq 1 5); do
    echo "Downloading kubectl (attempt $i/5)..."
    # First get the stable version
    KUBECTL_VERSION=$(curl -fsSL --retry 3 --retry-delay 5 https://dl.k8s.io/release/stable.txt 2>/dev/null)
    if [ -z "$KUBECTL_VERSION" ]; then
        echo "Failed to get kubectl version, retrying..."
        sleep 5
        continue
    fi

    # Download the binary
    if curl -fsSL --retry 3 --retry-delay 5 -o kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"; then
        DOWNLOAD_OK=1
        break
    fi
    echo "Download failed, retrying in 5 seconds..."
    sleep 5
done

if [ "$DOWNLOAD_OK" = "0" ]; then
    echo "ERROR: Failed to download kubectl after 5 attempts"
    echo "Network connectivity issues detected"
    exit 6
fi

# Install kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Clean up
rm -f kubectl

echo "kubectl installed successfully"
`

	// Base64-encode the script to avoid shell character interpretation issues
	// when passing through Windows -> WSL -> bash argument chain
	encoded := base64.StdEncoding.EncodeToString([]byte(installScript))
	wrapperCmd := fmt.Sprintf("echo %s | base64 -d | bash", encoded)

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", wrapperCmd)
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

	// Write the helper script to WSL2 using base64 encoding to avoid shell escaping issues
	writeScript := fmt.Sprintf(`#!/bin/bash
cat > /tmp/kubectl-wrapper.sh << 'EOFSCRIPT'
%s
EOFSCRIPT
sudo mv /tmp/kubectl-wrapper.sh /usr/local/bin/kubectl-wrapper.sh
sudo chmod +x /usr/local/bin/kubectl-wrapper.sh
`, helperScript)

	// Base64-encode the script to avoid shell character interpretation issues
	encoded := base64.StdEncoding.EncodeToString([]byte(writeScript))
	wrapperCmd := fmt.Sprintf("echo %s | base64 -d | bash", encoded)

	cmd := exec.Command("wsl", "-d", "Ubuntu", "bash", "-c", wrapperCmd)
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