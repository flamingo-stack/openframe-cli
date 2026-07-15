// Package gcloud installs the Google Cloud CLI pieces the GKE provider needs:
// the gcloud CLI itself and the gke-gcloud-auth-plugin kubeconfig exec plugin.
package gcloud

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func runQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
	return cmd.Run()
}

// GcloudInstaller manages the gcloud CLI.
type GcloudInstaller struct{}

func NewGcloudInstaller() *GcloudInstaller { return &GcloudInstaller{} }

func (g *GcloudInstaller) IsInstalled() bool {
	if !commandExists("gcloud") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "gcloud", "--version").Run() == nil
}

func (g *GcloudInstaller) GetInstallHelp() string {
	return platform.InstallHint("gcloud")
}

// Install installs the gcloud CLI. On macOS Homebrew carries the SDK; on
// Linux the official install requires a distribution-specific repo setup, so
// the user is pointed at the docs instead of a fragile automated attempt.
func (g *GcloudInstaller) Install() error {
	switch runtime.GOOS {
	case "darwin":
		if !commandExists("brew") {
			return fmt.Errorf("automatic gcloud installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
		}
		if err := runQuiet("brew", "install", "--cask", "google-cloud-sdk"); err != nil {
			return fmt.Errorf("failed to install the Google Cloud SDK: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("automatic gcloud installation is not supported on %s. %s", runtime.GOOS, g.GetInstallHelp())
	}
}

// AuthPluginInstaller manages gke-gcloud-auth-plugin, the exec plugin GKE
// kubeconfigs authenticate through.
type AuthPluginInstaller struct{}

func NewAuthPluginInstaller() *AuthPluginInstaller { return &AuthPluginInstaller{} }

func (a *AuthPluginInstaller) IsInstalled() bool {
	return commandExists("gke-gcloud-auth-plugin")
}

func (a *AuthPluginInstaller) GetInstallHelp() string {
	return platform.InstallHint("gke-gcloud-auth-plugin")
}

// Install installs the plugin through gcloud's component manager (requires
// gcloud itself, which precedes this requirement in the GKE set).
func (a *AuthPluginInstaller) Install() error {
	if !commandExists("gcloud") {
		return fmt.Errorf("gke-gcloud-auth-plugin is installed via gcloud, which is missing")
	}
	if err := runQuiet("gcloud", "components", "install", "gke-gcloud-auth-plugin", "--quiet"); err != nil {
		return fmt.Errorf("failed to install gke-gcloud-auth-plugin (for package-manager gcloud installs, use the OS package instead — see %s): %w",
			"https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl", err)
	}
	return nil
}
