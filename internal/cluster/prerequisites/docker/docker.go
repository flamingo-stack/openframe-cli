package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/pterm/pterm"
)

type DockerInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// isDockerInstalled reports whether the docker CLI is present.
//
// No Windows branch: on Windows the root command forwards the whole CLI into
// WSL before any command runs, so this code only executes as a Linux process
// (see wsllauncher). The old branch probed WSL from the outside and hardcoded
// the "Ubuntu" distro, contradicting the distro-agnostic launcher.
func isDockerInstalled() bool {
	return commandExists("docker")
}

// IsDockerRunning reports whether the docker daemon answers. See
// isDockerInstalled for why there is no Windows branch.
func IsDockerRunning() bool {
	if !commandExists("docker") {
		return false
	}
	// Check if Docker daemon is accessible by running docker ps with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "ps")
	err := cmd.Run()
	return err == nil
}

func dockerInstallHelp() string {
	return platform.InstallHint("docker")
}

func NewDockerInstaller() *DockerInstaller {
	return &DockerInstaller{}
}

func (d *DockerInstaller) IsInstalled() bool {
	return isDockerInstalled()
}

func (d *DockerInstaller) GetInstallHelp() string {
	return dockerInstallHelp()
}

func (d *DockerInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return d.installMacOS()
	case "linux":
		return d.installLinux()
	default:
		// Windows is unsupported here by design: the CLI forwards into WSL and
		// runs as linux, so native-Windows install code is never reached.
		return fmt.Errorf("automatic Docker installation not supported on %s", runtime.GOOS)
	}
}

func (d *DockerInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic Docker installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}

	pterm.Info.Println("Installing Docker Desktop via Homebrew...")
	cmd := exec.Command("brew", "install", "--cask", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Docker Desktop: %w", err)
	}

	pterm.Info.Println("Starting Docker Desktop...")
	cmd = exec.Command("open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		pterm.Warning.Printfln("Could not start Docker Desktop automatically: %v", err)
		pterm.Info.Println("Please start Docker Desktop manually from Applications")
	}

	return nil
}

func (d *DockerInstaller) installLinux() error {
	switch {
	case commandExists("apk"):
		return d.installAlpine()
	case commandExists("apt"):
		return d.installUbuntu()
	case commandExists("yum"):
		return d.installRedHat()
	case commandExists("dnf"):
		return d.installFedora()
	case commandExists("pacman"):
		return d.installArch()
	default:
		return fmt.Errorf("no supported package manager found. Please install Docker manually from https://docs.docker.com/engine/install/")
	}
}

// installAlpine installs Docker on Alpine Linux following
// https://wiki.alpinelinux.org/wiki/Docker — apk add docker, then enable and
// start the OpenRC service. Alpine's default user is often root (and may not
// ship sudo, e.g. in WSL/containers), so sudo is only prefixed when needed.
// Enabling/starting the service is best-effort: under WSL or containers OpenRC
// may not be the init system, but `apk add docker` already provides the engine,
// which can be started directly (see StartDocker).
func (d *DockerInstaller) installAlpine() error {
	pterm.Info.Println("Installing Docker on Alpine Linux...")

	run := func(args ...string) error {
		if os.Geteuid() != 0 && commandExists("sudo") {
			args = append([]string{"sudo"}, args...)
		}
		return d.runCommand(args[0], args[1:]...)
	}

	if err := run("apk", "add", "--no-cache", "docker"); err != nil {
		return fmt.Errorf("failed to install Docker with apk: %w", err)
	}
	if err := run("rc-update", "add", "docker", "default"); err != nil {
		pterm.Warning.Printfln("Could not enable the docker service (rc-update): %v", err)
	}
	if err := run("rc-service", "docker", "start"); err != nil {
		pterm.Warning.Printfln("Could not start the docker service (rc-service): %v", err)
	}

	// Add the current user to the docker group (Alpine uses addgroup, not usermod).
	if user := os.Getenv("USER"); user != "" && user != "root" {
		if err := run("addgroup", user, "docker"); err != nil {
			pterm.Warning.Printfln("Could not add user to the docker group: %v", err)
		} else {
			pterm.Info.Println("Log out and back in for Docker group permissions to take effect")
		}
	}

	return nil
}

func (d *DockerInstaller) installUbuntu() error {
	pterm.Info.Println("Installing Docker on Ubuntu/Debian...")

	commands := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg", "lsb-release"},
	}

	for _, cmdArgs := range commands {
		if err := d.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add Docker's official GPG key
	gpgCmd := "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg"
	if err := d.runShellCommand(gpgCmd); err != nil {
		return fmt.Errorf("failed to add Docker GPG key: %w", err)
	}

	// Add Docker repository
	repoCmd := `echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`
	if err := d.runShellCommand(repoCmd); err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}

	// Install Docker
	installCommands := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io"},
		{"sudo", "systemctl", "enable", "docker"},
		{"sudo", "systemctl", "start", "docker"},
	}

	for _, cmdArgs := range installCommands {
		if err := d.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add user to docker group
	user := os.Getenv("USER")
	if user != "" {
		if err := d.runCommand("sudo", "usermod", "-aG", "docker", user); err != nil {
			pterm.Warning.Printfln("Could not add user to docker group: %v", err)
		} else {
			pterm.Info.Println("You may need to log out and back in for Docker group permissions to take effect")
		}
	}

	return nil
}

func (d *DockerInstaller) installRedHat() error {
	pterm.Info.Println("Installing Docker on CentOS/RHEL...")

	commands := [][]string{
		{"sudo", "yum", "install", "-y", "yum-utils"},
		{"sudo", "yum-config-manager", "--add-repo", "https://download.docker.com/linux/centos/docker-ce.repo"},
		{"sudo", "yum", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io"},
		{"sudo", "systemctl", "enable", "docker"},
		{"sudo", "systemctl", "start", "docker"},
	}

	for _, cmdArgs := range commands {
		if err := d.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add user to docker group
	user := os.Getenv("USER")
	if user != "" {
		if err := d.runCommand("sudo", "usermod", "-aG", "docker", user); err != nil {
			pterm.Warning.Printfln("Could not add user to docker group: %v", err)
		} else {
			pterm.Info.Println("You may need to log out and back in for Docker group permissions to take effect")
		}
	}

	return nil
}

func (d *DockerInstaller) installFedora() error {
	pterm.Info.Println("Installing Docker on Fedora...")

	commands := [][]string{
		{"sudo", "dnf", "install", "-y", "dnf-plugins-core"},
		{"sudo", "dnf", "config-manager", "--add-repo", "https://download.docker.com/linux/fedora/docker-ce.repo"},
		{"sudo", "dnf", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io"},
		{"sudo", "systemctl", "enable", "docker"},
		{"sudo", "systemctl", "start", "docker"},
	}

	for _, cmdArgs := range commands {
		if err := d.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add user to docker group
	user := os.Getenv("USER")
	if user != "" {
		if err := d.runCommand("sudo", "usermod", "-aG", "docker", user); err != nil {
			pterm.Warning.Printfln("Could not add user to docker group: %v", err)
		} else {
			pterm.Info.Println("You may need to log out and back in for Docker group permissions to take effect")
		}
	}

	return nil
}

func (d *DockerInstaller) installArch() error {
	pterm.Info.Println("Installing Docker on Arch Linux...")

	commands := [][]string{
		{"sudo", "pacman", "-S", "--noconfirm", "docker"},
		{"sudo", "systemctl", "enable", "docker"},
		{"sudo", "systemctl", "start", "docker"},
	}

	for _, cmdArgs := range commands {
		if err := d.runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	// Add user to docker group
	user := os.Getenv("USER")
	if user != "" {
		if err := d.runCommand("sudo", "usermod", "-aG", "docker", user); err != nil {
			pterm.Warning.Printfln("Could not add user to docker group: %v", err)
		} else {
			pterm.Info.Println("You may need to log out and back in for Docker group permissions to take effect")
		}
	}

	return nil
}

func (d *DockerInstaller) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 G702 -- explicit argv, no shell; command and args are internal, not untrusted input
	// Completely silence output during installation
	return cmd.Run()
}

func (d *DockerInstaller) runShellCommand(command string) error {
	cmd := exec.Command("bash", "-c", command) // #nosec G204 -- shell string built from constant/program-derived values, not untrusted input
	// Completely silence output during installation
	return cmd.Run()
}

// StartDocker attempts to start Docker based on the operating system
func StartDocker() error {
	switch runtime.GOOS {
	case "darwin":
		return startDockerMacOS()
	case "linux":
		return startDockerLinux()
	case "windows":
		// Unreachable in the supported flow (the CLI runs inside WSL as linux);
		// the previous WSL-from-Windows starter hardcoded the Ubuntu distro.
		return fmt.Errorf("starting Docker from the native Windows launcher is not supported — run openframe inside WSL")
	default:
		return fmt.Errorf("starting Docker is not supported on %s", runtime.GOOS)
	}
}

func startDockerMacOS() error {
	// Try to start Docker Desktop on macOS
	cmd := exec.Command("open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		// Try alternative command
		cmd = exec.Command("open", "/Applications/Docker.app")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Docker Desktop: %w", err)
		}
	}
	return nil
}

func startDockerLinux() error {
	// Try to start Docker daemon on Linux
	// First check if systemctl exists (systemd)
	if commandExists("systemctl") {
		cmd := exec.Command("sudo", "systemctl", "start", "docker")
		if err := cmd.Run(); err != nil {
			// Try without sudo in case user has permissions
			cmd = exec.Command("systemctl", "start", "docker")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to start Docker daemon with systemctl: %w", err)
			}
		}
		return nil
	}

	// OpenRC (Alpine)
	if commandExists("rc-service") {
		if exec.Command("rc-service", "docker", "start").Run() == nil {
			return nil
		}
		if err := exec.Command("sudo", "rc-service", "docker", "start").Run(); err != nil {
			return fmt.Errorf("failed to start Docker daemon with rc-service: %w", err)
		}
		return nil
	}

	// Try service command (older systems)
	if commandExists("service") {
		cmd := exec.Command("sudo", "service", "docker", "start")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Docker daemon with service: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unable to start Docker daemon: no supported init system found")
}

// WaitForDocker waits for the Docker daemon to become available. The budget is
// generous because a cold Docker Desktop start on macOS routinely exceeds the
// old 30s ceiling; the poll returns as soon as the daemon answers, so healthy
// setups pay nothing extra.
func WaitForDocker() error {
	maxAttempts := 120 // seconds
	for i := 0; i < maxAttempts; i++ {
		if IsDockerRunning() {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for Docker to start")
}
