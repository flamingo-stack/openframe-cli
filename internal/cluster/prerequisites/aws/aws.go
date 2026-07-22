package aws

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
)

// AwsInstaller installs the AWS CLI required by the EKS provider (kubeconfig
// auth runs through `aws eks get-token`). Credentials are NOT checked here —
// the EKS provider preflights them with `aws sts get-caller-identity` so the
// error can name the profile being used.
type AwsInstaller struct{}

func NewAwsInstaller() *AwsInstaller {
	return &AwsInstaller{}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (a *AwsInstaller) IsInstalled() bool {
	if !commandExists("aws") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "aws", "--version").Run() == nil
}

func (a *AwsInstaller) GetInstallHelp() string {
	return platform.InstallHint("aws")
}

// Install installs the AWS CLI via the platform package manager. There is no
// verified-download fallback: AWS ships the v2 CLI as a frequently-rotated
// installer bundle without stable per-version checksums to pin.
func (a *AwsInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		return a.installMacOS()
	case "linux":
		return a.installLinux()
	default:
		// Windows is unsupported here by design: the CLI forwards into WSL and
		// runs as linux, so native-Windows install code is never reached.
		return fmt.Errorf("automatic AWS CLI installation not supported on %s", runtime.GOOS)
	}
}

func (a *AwsInstaller) installMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic AWS CLI installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}
	if err := runCommand("brew", "install", "awscli"); err != nil {
		return fmt.Errorf("failed to install AWS CLI: %w", err)
	}
	return nil
}

func (a *AwsInstaller) installLinux() error {
	type pm struct {
		name string
		args []string
	}
	managers := []pm{
		{"apt", []string{"apt", "install", "-y", "awscli"}},
		{"pacman", []string{"pacman", "-S", "--noconfirm", "aws-cli-v2"}},
	}
	for _, m := range managers {
		if !commandExists(m.name) {
			continue
		}
		// Package installs need root; sudo -n keeps this non-interactive (the
		// prerequisite flow already runs under a user confirmation).
		if err := runCommand("sudo", append([]string{"-n"}, m.args...)...); err == nil {
			return nil
		}
	}
	return fmt.Errorf("could not install the AWS CLI automatically. %s", a.GetInstallHelp())
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
	return cmd.Run()
}
