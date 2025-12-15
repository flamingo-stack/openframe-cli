package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type WSLInstaller struct{}

func NewWSLInstaller() *WSLInstaller {
	return &WSLInstaller{}
}

// IsApplicable returns true if WSL check is relevant (Windows only, not inside WSL)
func (w *WSLInstaller) IsApplicable() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Don't check WSL if we're already running inside WSL
	return !IsRunningInWSL()
}

// IsRunningInWSL detects if we're running inside WSL
func IsRunningInWSL() bool {
	// Check for WSL-specific environment variables
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	if os.Getenv("WSLENV") != "" {
		return true
	}

	// Check /proc/version for Microsoft/WSL indicators (Linux check)
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/version")
		if err == nil {
			version := strings.ToLower(string(data))
			if strings.Contains(version, "microsoft") || strings.Contains(version, "wsl") {
				return true
			}
		}
	}

	return false
}

// IsInstalled checks if WSL is installed and functional
func (w *WSLInstaller) IsInstalled() bool {
	if !w.IsApplicable() {
		return true // Not applicable means we consider it "installed" (skip)
	}

	// Check if wsl.exe exists
	_, err := exec.LookPath("wsl.exe")
	if err != nil {
		return false
	}

	// Check if WSL is functional by running --status
	cmd := exec.Command("wsl", "--status")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// IsUbuntuInstalled checks if Ubuntu distribution is installed in WSL
func (w *WSLInstaller) IsUbuntuInstalled() bool {
	if !w.IsInstalled() {
		return false
	}

	// Check if Ubuntu distribution exists
	cmd := exec.Command("wsl", "-d", "Ubuntu", "echo", "ok")
	return cmd.Run() == nil
}

// GetInstallHelp returns installation instructions for WSL
func (w *WSLInstaller) GetInstallHelp() string {
	return "WSL2: Run 'wsl --install' in an elevated PowerShell/Command Prompt, then restart your computer"
}

// Install installs WSL2 on Windows
func (w *WSLInstaller) Install() error {
	if !w.IsApplicable() {
		return nil // Nothing to do
	}

	// Check if wsl.exe exists but WSL is not properly configured
	_, err := exec.LookPath("wsl.exe")
	if err != nil {
		// WSL not present at all, need to install
		return w.installWSL()
	}

	// WSL binary exists, check if it's functional
	cmd := exec.Command("wsl", "--status")
	if err := cmd.Run(); err != nil {
		// WSL exists but not functional, try to install/enable
		return w.installWSL()
	}

	// WSL is installed and functional
	fmt.Println("WSL2 is already installed and functional")

	// Ensure WSL2 is the default version
	setDefaultCmd := exec.Command("wsl", "--set-default-version", "2")
	setDefaultCmd.Run() // Ignore errors, might already be set

	return nil
}

func (w *WSLInstaller) installWSL() error {
	fmt.Println("Installing WSL2...")
	fmt.Println("Note: This may require administrator privileges and a system restart")

	// Install WSL2 without a distribution first
	cmd := exec.Command("wsl", "--install", "--no-distribution")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check if it's an elevation error
		return fmt.Errorf("failed to install WSL2. Please run as Administrator or manually run: wsl --install")
	}

	// Set WSL2 as the default version
	setDefaultCmd := exec.Command("wsl", "--set-default-version", "2")
	setDefaultCmd.Run() // Ignore errors

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("WSL2 installation initiated!")
	fmt.Println("")
	fmt.Println("IMPORTANT: You MUST restart your computer")
	fmt.Println("for WSL2 to become fully functional.")
	fmt.Println("")
	fmt.Println("After restart, run this command again to")
	fmt.Println("continue with the installation.")
	fmt.Println("========================================")

	// Return a special error to indicate restart is needed
	return &RestartRequiredError{Message: "System restart required to complete WSL2 installation"}
}

// InstallUbuntu installs Ubuntu distribution in WSL2
func (w *WSLInstaller) InstallUbuntu() error {
	if !w.IsInstalled() {
		return fmt.Errorf("WSL2 must be installed first")
	}

	// Check if Ubuntu is already installed
	if w.IsUbuntuInstalled() {
		fmt.Println("Ubuntu is already installed in WSL2")
		return nil
	}

	fmt.Println("Installing Ubuntu in WSL2...")
	cmd := exec.Command("wsl", "--install", "-d", "Ubuntu", "--no-launch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check if error is because distribution already exists
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ERROR_ALREADY_EXISTS") {
			fmt.Println("Ubuntu already exists in WSL2")
			return nil
		}
		return fmt.Errorf("failed to install Ubuntu: %w", err)
	}

	fmt.Println("Ubuntu installed successfully in WSL2")
	return nil
}

// RestartRequiredError indicates that a system restart is needed
type RestartRequiredError struct {
	Message string
}

func (e *RestartRequiredError) Error() string {
	return e.Message
}

// IsRestartRequired checks if the error indicates a restart is needed
func IsRestartRequired(err error) bool {
	_, ok := err.(*RestartRequiredError)
	return ok
}
