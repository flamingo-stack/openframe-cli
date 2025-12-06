package k3d

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type K3dInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isK3dInstalled() bool {
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
	switch runtime.GOOS {
	case "darwin":
		return "k3d: Run 'brew install k3d' or download from https://k3d.io/v5.4.6/#installation"
	case "linux":
		return "k3d: Run 'curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash' or download from https://k3d.io/v5.4.6/#installation"
	case "windows":
		return "k3d: Download from https://github.com/k3d-io/k3d/releases or use chocolatey 'choco install k3d'"
	default:
		return "k3d: Please install k3d from https://k3d.io/v5.4.6/#installation"
	}
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
		return fmt.Errorf("Homebrew is required for automatic k3d installation on macOS. Please install brew first: https://brew.sh")
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
		return k.installScript()
	}
}

func (k *K3dInstaller) installUbuntu() error {
	// k3d doesn't have official apt repository, so use the install script
	return k.installScript()
}

func (k *K3dInstaller) installRedHat() error {
	// k3d doesn't have official yum repository, so use the install script
	return k.installScript()
}

func (k *K3dInstaller) installFedora() error {
	// k3d doesn't have official dnf repository, so use the install script
	return k.installScript()
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
	return k.installScript()
}

func (k *K3dInstaller) installScript() error {
	// Use the official k3d install script
	installCmd := "curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"

	if err := k.runShellCommand(installCmd); err != nil {
		return fmt.Errorf("failed to install k3d via script: %w", err)
	}

	return nil
}

func (k *K3dInstaller) installBinary() error {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Get latest release version
	versionCmd := "curl -s https://api.github.com/repos/k3d-io/k3d/releases/latest | grep '\"tag_name\":' | sed -E 's/.*\"([^\"]+)\".*/\\1/'"

	commands := []string{
		fmt.Sprintf("VERSION=$(%s)", versionCmd),
		fmt.Sprintf("curl -Lo k3d https://github.com/k3d-io/k3d/releases/download/${VERSION}/k3d-linux-%s", arch),
		"chmod +x k3d",
		"sudo mv k3d /usr/local/bin/",
	}

	for _, cmd := range commands {
		if err := k.runShellCommand(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmd, err)
		}
	}

	return nil
}

func (k *K3dInstaller) installWindows() error {
	fmt.Println("Installing k3d natively on Windows...")

	// Try package managers first
	if commandExists("choco") {
		fmt.Println("Installing k3d via Chocolatey...")
		cmd := exec.Command("choco", "install", "k3d", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ k3d installed successfully via Chocolatey!")
			return nil
		}
		fmt.Println("Chocolatey installation failed, trying other methods...")
	}

	if commandExists("winget") {
		fmt.Println("Installing k3d via winget...")
		cmd := exec.Command("winget", "install", "--id", "k3d-io.k3d", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("✓ k3d installed successfully via winget!")
			return nil
		}
		fmt.Println("winget installation failed, trying other methods...")
	}

	// Fallback to direct binary download
	return k.installWindowsBinary()
}

func (k *K3dInstaller) installWindowsBinary() error {
	fmt.Println("Downloading k3d.exe binary...")

	binDir := os.Getenv("USERPROFILE") + "\\bin"
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	k3dPath := binDir + "\\k3d.exe"

	// Download k3d using PowerShell
	downloadCmd := `
$ProgressPreference = 'SilentlyContinue'
$releases = Invoke-RestMethod -Uri "https://api.github.com/repos/k3d-io/k3d/releases/latest"
$version = $releases.tag_name
$downloadUrl = "https://github.com/k3d-io/k3d/releases/download/$version/k3d-windows-amd64.exe"
Write-Host "Downloading k3d $version..."
Invoke-WebRequest -Uri $downloadUrl -OutFile "` + k3dPath + `"
Write-Host "k3d installed successfully!"
`

	cmd := exec.Command("powershell", "-Command", downloadCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download k3d.exe: %w", err)
	}

	// Add to PATH
	k.addToPath(binDir)

	fmt.Printf("✓ k3d.exe installed successfully at: %s\n", k3dPath)
	return nil
}

func (k *K3dInstaller) addToPath(binDir string) {
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

func (k *K3dInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Completely silence output during installation
	return cmd.Run()
}

func (k *K3dInstaller) runShellCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	// Completely silence output during installation
	return cmd.Run()
}
