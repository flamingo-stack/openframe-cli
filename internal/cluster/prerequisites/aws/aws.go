package aws

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
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

// installLinux installs the AWS CLI v2 via the platform package manager. The
// v2 package name differs per distro: Debian 12+/Ubuntu 24.04+ ship v2 as
// "awscli", Fedora/RHEL as "awscli2", Arch as "aws-cli-v2".
func (a *AwsInstaller) installLinux() error {
	type pm struct {
		name string
		args []string
	}
	managers := []pm{
		{"apt", []string{"apt", "install", "-y", "awscli"}},
		{"dnf", []string{"dnf", "install", "-y", "awscli2"}},
		{"yum", []string{"yum", "install", "-y", "awscli2"}},
		{"pacman", []string{"pacman", "-S", "--noconfirm", "aws-cli-v2"}},
	}
	for _, m := range managers {
		if !commandExists(m.name) {
			continue
		}
		// Package installs need root; sudo -n keeps this non-interactive (the
		// prerequisite flow already runs under a user confirmation).
		if err := runCommand("sudo", append([]string{"-n"}, m.args...)...); err == nil {
			if installedIsV2() {
				return nil
			}
			// Older repos (e.g. Ubuntu 22.04) ship legacy v1, whose
			// `aws eks get-token` emits an auth API kubectl no longer accepts.
			return fmt.Errorf("the distro package installed AWS CLI v1, but the EKS flow needs v2. %s", a.GetInstallHelp())
		}
	}
	return fmt.Errorf("could not install the AWS CLI automatically. %s", a.GetInstallHelp())
}

// installedIsV2 reports whether the aws binary on PATH is the v2 CLI. v1
// prints its version to stderr, v2 to stdout — check both.
func installedIsV2() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "aws", "--version").CombinedOutput()
	return err == nil && strings.HasPrefix(string(out), "aws-cli/2.")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
	return cmd.Run()
}
