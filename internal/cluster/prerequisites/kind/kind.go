package kind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// KindInstaller handles installation of kind and related tools on Windows
type KindInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// isKindInstalled checks if kind is installed natively on Windows
func isKindInstalled() bool {
	// On Windows, check for native kind.exe in PATH
	if runtime.GOOS == "windows" {
		_, err := exec.LookPath("kind.exe")
		if err == nil {
			return true
		}
		// Also check without .exe extension
		_, err = exec.LookPath("kind")
		return err == nil
	}

	// On non-Windows, check kind with timeout
	if !commandExists("kind") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kind", "version")
	err := cmd.Run()
	return err == nil
}

func kindInstallHelp() string {
	switch runtime.GOOS {
	case "darwin":
		return "kind: Run 'brew install kind' or download from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
	case "linux":
		return "kind: Download from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
	case "windows":
		return "kind: Run 'choco install kind' or 'winget install Kubernetes.kind' or download from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
	default:
		return "kind: Please install kind from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
	}
}

func NewKindInstaller() *KindInstaller {
	return &KindInstaller{}
}

func (k *KindInstaller) IsInstalled() bool {
	return isKindInstalled()
}

func (k *KindInstaller) GetInstallHelp() string {
	return kindInstallHelp()
}

func (k *KindInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return k.installMacOS()
	case "linux":
		return k.installLinux()
	case "windows":
		return k.installWindows()
	default:
		return fmt.Errorf("automatic kind installation not supported on %s", runtime.GOOS)
	}
}

func (k *KindInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("Homebrew is required for automatic kind installation on macOS. Please install brew first: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "kind")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install kind: %w", err)
	}

	return nil
}

func (k *KindInstaller) installLinux() error {
	return k.installBinaryLinux()
}

func (k *KindInstaller) installBinaryLinux() error {
	fmt.Println("Installing kind via direct binary download...")

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	commands := []string{
		fmt.Sprintf("curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.25.0/kind-linux-%s", arch),
		"chmod +x ./kind",
		"sudo mv ./kind /usr/local/bin/kind",
	}

	for _, cmdStr := range commands {
		cmd := exec.Command("bash", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmdStr, err)
		}
	}

	return nil
}

// installWindows installs kind and related tools natively on Windows
func (k *KindInstaller) installWindows() error {
	fmt.Println("Installing kind and related tools natively on Windows...")

	// Try to install using package managers first (Chocolatey or winget)
	if err := k.tryPackageManagerInstall(); err == nil {
		fmt.Println("kind installed successfully via package manager!")
		return nil
	}

	// Fallback to direct binary download
	fmt.Println("Package managers not available, using direct binary download...")
	return k.installWindowsBinary()
}

// tryPackageManagerInstall attempts to install kind using Chocolatey or winget
func (k *KindInstaller) tryPackageManagerInstall() error {
	// Try Chocolatey first
	if commandExists("choco") {
		fmt.Println("Installing kind via Chocolatey...")
		cmd := exec.Command("choco", "install", "kind", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		fmt.Println("Chocolatey installation failed, trying other methods...")
	}

	// Try winget
	if commandExists("winget") {
		fmt.Println("Installing kind via winget...")
		cmd := exec.Command("winget", "install", "--id", "Kubernetes.kind", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		fmt.Println("winget installation failed, trying other methods...")
	}

	// Try Scoop
	if commandExists("scoop") {
		fmt.Println("Installing kind via Scoop...")
		cmd := exec.Command("scoop", "install", "kind")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		fmt.Println("Scoop installation failed, trying other methods...")
	}

	return fmt.Errorf("no package manager available")
}

// installWindowsBinary downloads and installs kind.exe directly
func (k *KindInstaller) installWindowsBinary() error {
	fmt.Println("Downloading kind.exe binary...")

	// Determine architecture
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture for Windows: %s", arch)
	}

	// Create bin directory in user profile
	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	kindPath := filepath.Join(binDir, "kind.exe")
	downloadURL := fmt.Sprintf("https://kind.sigs.k8s.io/dl/v0.25.0/kind-windows-%s", arch)

	// Download kind.exe using PowerShell
	downloadCmd := fmt.Sprintf(`
$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri "%s" -OutFile "%s"
`, downloadURL, kindPath)

	cmd := exec.Command("powershell", "-Command", downloadCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download kind.exe: %w", err)
	}

	// Add to PATH if not already there
	if err := k.addToWindowsPath(binDir); err != nil {
		fmt.Printf("Warning: Could not add %s to PATH: %v\n", binDir, err)
		fmt.Printf("Please add %s to your PATH manually\n", binDir)
	}

	// Update current process PATH
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, binDir) {
		os.Setenv("PATH", currentPath+";"+binDir)
	}

	fmt.Printf("kind.exe installed successfully at: %s\n", kindPath)
	return nil
}

// addToWindowsPath adds a directory to the Windows user PATH
func (k *KindInstaller) addToWindowsPath(dir string) error {
	addPathScript := fmt.Sprintf(`
$binDir = "%s"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
    Write-Host "Added $binDir to PATH"
} else {
    Write-Host "PATH already contains $binDir"
}
`, dir)

	cmd := exec.Command("powershell", "-Command", addPathScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

// InstallKubectlWindows installs kubectl.exe natively on Windows
func InstallKubectlWindows() error {
	fmt.Println("Installing kubectl natively on Windows...")

	// Check if already installed
	if _, err := exec.LookPath("kubectl.exe"); err == nil {
		fmt.Println("kubectl.exe already installed")
		return nil
	}
	if _, err := exec.LookPath("kubectl"); err == nil {
		fmt.Println("kubectl already installed")
		return nil
	}

	// Try package managers first
	if commandExists("choco") {
		fmt.Println("Installing kubectl via Chocolatey...")
		cmd := exec.Command("choco", "install", "kubernetes-cli", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	if commandExists("winget") {
		fmt.Println("Installing kubectl via winget...")
		cmd := exec.Command("winget", "install", "--id", "Kubernetes.kubectl", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to direct download
	return installKubectlWindowsBinary()
}

func installKubectlWindowsBinary() error {
	fmt.Println("Downloading kubectl.exe binary...")

	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	kubectlPath := filepath.Join(binDir, "kubectl.exe")

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
	addToPath(binDir)

	fmt.Printf("kubectl.exe installed successfully at: %s\n", kubectlPath)
	return nil
}

// InstallHelmWindows installs helm.exe natively on Windows
func InstallHelmWindows() error {
	fmt.Println("Installing helm natively on Windows...")

	// Check if already installed
	if _, err := exec.LookPath("helm.exe"); err == nil {
		fmt.Println("helm.exe already installed")
		return nil
	}
	if _, err := exec.LookPath("helm"); err == nil {
		fmt.Println("helm already installed")
		return nil
	}

	// Try package managers first
	if commandExists("choco") {
		fmt.Println("Installing helm via Chocolatey...")
		cmd := exec.Command("choco", "install", "kubernetes-helm", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	if commandExists("winget") {
		fmt.Println("Installing helm via winget...")
		cmd := exec.Command("winget", "install", "--id", "Helm.Helm", "-e", "--accept-source-agreements", "--accept-package-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to direct download
	return installHelmWindowsBinary()
}

func installHelmWindowsBinary() error {
	fmt.Println("Downloading helm.exe binary...")

	binDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
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
	addToPath(binDir)

	helmPath := filepath.Join(binDir, "helm.exe")
	fmt.Printf("helm.exe installed successfully at: %s\n", helmPath)
	return nil
}

func addToPath(binDir string) {
	addPathScript := fmt.Sprintf(`
$binDir = "%s"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binDir", "User")
}
`, binDir)

	cmd := exec.Command("powershell", "-Command", addPathScript)
	cmd.Run() // Ignore errors

	// Update current process PATH
	currentPath := os.Getenv("PATH")
	if !containsPath(currentPath, binDir) {
		os.Setenv("PATH", currentPath+";"+binDir)
	}
}

// IsKubectlInstalledNative checks if kubectl is installed natively on Windows
func IsKubectlInstalledNative() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	if _, err := exec.LookPath("kubectl.exe"); err == nil {
		return true
	}
	_, err := exec.LookPath("kubectl")
	return err == nil
}

// IsHelmInstalledNative checks if helm is installed natively on Windows
func IsHelmInstalledNative() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	if _, err := exec.LookPath("helm.exe"); err == nil {
		return true
	}
	_, err := exec.LookPath("helm")
	return err == nil
}
