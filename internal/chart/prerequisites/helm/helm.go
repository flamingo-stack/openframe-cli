package helm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type HelmInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isHelmInstalled() bool {
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
	switch runtime.GOOS {
	case "darwin":
		return "Helm: Run 'brew install helm' or download from https://helm.sh/docs/intro/install/"
	case "linux":
		return "Helm: Run 'curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash' or download from https://helm.sh/docs/intro/install/"
	case "windows":
		return "Helm: Download from https://helm.sh/docs/intro/install/ or install via chocolatey 'choco install kubernetes-helm'"
	default:
		return "Helm: Please install Helm from https://helm.sh/docs/intro/install/"
	}
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
		return fmt.Errorf("Homebrew is required for automatic Helm installation on macOS. Please install brew first: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "helm")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Helm: %w", err)
	}

	return nil
}

func (h *HelmInstaller) installLinux() error {
	if commandExists("apt") {
		return h.installUbuntu()
	} else if commandExists("yum") {
		return h.installRedHat()
	} else if commandExists("dnf") {
		return h.installFedora()
	} else if commandExists("pacman") {
		return h.installArch()
	} else {
		return h.installScript()
	}
}

func (h *HelmInstaller) installUbuntu() error {
	commands := []string{
		"curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null",
		"sudo apt-get install apt-transport-https --yes",
		"echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main\" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list",
		"sudo apt-get update",
		"sudo apt-get install helm",
	}

	for _, cmd := range commands {
		if err := h.runShellCommand(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmd, err)
		}
	}

	return nil
}

func (h *HelmInstaller) installRedHat() error {
	commands := []string{
		"curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash",
	}

	for _, cmd := range commands {
		if err := h.runShellCommand(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmd, err)
		}
	}

	return nil
}

func (h *HelmInstaller) installFedora() error {
	if err := h.runCommand("sudo", "dnf", "install", "-y", "helm"); err != nil {
		// If dnf package not available, fall back to script
		return h.installScript()
	}
	return nil
}

func (h *HelmInstaller) installArch() error {
	if err := h.runCommand("sudo", "pacman", "-S", "--noconfirm", "helm"); err != nil {
		return fmt.Errorf("failed to install Helm: %w", err)
	}
	return nil
}

func (h *HelmInstaller) installScript() error {
	// Use the official Helm install script
	installCmd := "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"

	if err := h.runShellCommand(installCmd); err != nil {
		return fmt.Errorf("failed to install Helm via script: %w", err)
	}

	return nil
}

func (h *HelmInstaller) installWindows() error {
	fmt.Println("Installing Helm natively on Windows...")

	// Try package managers first
	if commandExists("choco") {
		fmt.Println("Installing Helm via Chocolatey...")
		cmd := exec.Command("choco", "install", "kubernetes-helm", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ Helm installed successfully via Chocolatey!")
			return nil
		}
		fmt.Println("Chocolatey installation failed, trying other methods...")
	}

	if commandExists("winget") {
		fmt.Println("Installing Helm via winget...")
		cmd := exec.Command("winget", "install", "--id", "Helm.Helm", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ Helm installed successfully via winget!")
			return nil
		}
		fmt.Println("winget installation failed, trying other methods...")
	}

	// Fallback to direct binary download
	return h.installWindowsBinary()
}

func (h *HelmInstaller) installWindowsBinary() error {
	fmt.Println("Downloading helm.exe binary...")

	binDir := os.Getenv("USERPROFILE") + "\\bin"
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Download Helm using PowerShell
	downloadCmd := `
$ProgressPreference = 'SilentlyContinue'
$helmVersion = "v3.16.3"
$helmUrl = "https://get.helm.sh/helm-$helmVersion-windows-amd64.zip"
$tempDir = [System.IO.Path]::GetTempPath()
$zipPath = Join-Path $tempDir "helm.zip"
$extractPath = Join-Path $tempDir "helm"

Write-Host "Downloading Helm..."
Invoke-WebRequest -Uri $helmUrl -OutFile $zipPath

Write-Host "Extracting Helm..."
Expand-Archive -Path $zipPath -DestinationPath $extractPath -Force

$binDir = "` + binDir + `"
Copy-Item -Path (Join-Path $extractPath "windows-amd64\helm.exe") -Destination (Join-Path $binDir "helm.exe") -Force

Write-Host "Cleaning up..."
Remove-Item -Path $zipPath -Force -ErrorAction SilentlyContinue
Remove-Item -Path $extractPath -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "Helm installed successfully!"
`

	cmd := exec.Command("powershell", "-Command", downloadCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download and install helm.exe: %w", err)
	}

	// Add to PATH
	h.addToPath(binDir)

	helmPath := binDir + "\\helm.exe"
	fmt.Printf("✓ helm.exe installed successfully at: %s\n", helmPath)
	return nil
}

func (h *HelmInstaller) addToPath(binDir string) {
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

func (h *HelmInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Completely silence output during installation
	return cmd.Run()
}

func (h *HelmInstaller) runShellCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	// Completely silence output during installation
	return cmd.Run()
}
