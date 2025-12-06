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
	fmt.Println("Installing kubectl natively on Windows...")

	// Try package managers first
	if commandExists("choco") {
		fmt.Println("Installing kubectl via Chocolatey...")
		cmd := exec.Command("choco", "install", "kubernetes-cli", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ kubectl installed successfully via Chocolatey!")
			return nil
		}
		fmt.Println("Chocolatey installation failed, trying other methods...")
	}

	if commandExists("winget") {
		fmt.Println("Installing kubectl via winget...")
		cmd := exec.Command("winget", "install", "--id", "Kubernetes.kubectl", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ kubectl installed successfully via winget!")
			return nil
		}
		fmt.Println("winget installation failed, trying other methods...")
	}

	// Fallback to direct binary download
	return k.installWindowsBinary()
}

func (k *KubectlInstaller) installWindowsBinary() error {
	fmt.Println("Downloading kubectl.exe binary...")

	binDir := os.Getenv("USERPROFILE") + "\\bin"
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	kubectlPath := binDir + "\\kubectl.exe"

	// Get latest stable version and download
	downloadCmd := `
$ProgressPreference = 'SilentlyContinue'
$version = (Invoke-WebRequest -Uri "https://dl.k8s.io/release/stable.txt" -UseBasicParsing).Content.Trim()
Invoke-WebRequest -Uri "https://dl.k8s.io/release/$version/bin/windows/amd64/kubectl.exe" -OutFile "` + kubectlPath + `"
`

	cmd := exec.Command("powershell", "-Command", downloadCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download kubectl.exe: %w", err)
	}

	// Add to PATH
	k.addToPath(binDir)

	fmt.Printf("✓ kubectl.exe installed successfully at: %s\n", kubectlPath)
	return nil
}

func (k *KubectlInstaller) addToPath(binDir string) {
	addPathScript := fmt.Sprintf(`
$binDir = "%s"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
    Write-Host "Added $binDir to PATH"
} else {
    Write-Host "PATH already contains $binDir"
}
`, binDir)

	cmd := exec.Command("powershell", "-Command", addPathScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // Ignore errors

	// Update current process PATH
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, binDir) {
		os.Setenv("PATH", currentPath+";"+binDir)
	}
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