package prerequisites

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	fw "github.com/flamingo-stack/openframe-cli/internal/prerequisites"
)

// ClusterSet returns the prerequisites required before creating or managing a
// local (k3d) cluster, expressed against the shared prerequisites framework:
// Docker (running), k3d, and helm. kubectl is not required — the CLI talks to
// Kubernetes via client-go, not the kubectl binary.
//
// On macOS/Linux the framework auto-installs any that are missing; on Windows it
// reports each missing tool with its manual setup guidance.
func ClusterSet() fw.Set {
	dockerInstaller := docker.NewDockerInstaller()
	k3dInstaller := k3d.NewK3dInstaller()
	helmInstaller := helm.NewHelmInstaller()

	return fw.Set{
		Name: "cluster",
		Items: []fw.Prerequisite{
			{
				// Docker must be running, not merely installed.
				Name:        "Docker",
				IsSatisfied: docker.IsDockerRunning,
				Install:     asCtxInstall(dockerInstaller.Install),
				DocsURL:     dockerInstaller.GetInstallHelp(),
				// When the binary is present but the daemon is down, say so instead
				// of the framework's default "not installed" — the fix a user needs
				// (start the daemon) is different from installing it.
				Detail: func() string {
					if dockerInstaller.IsInstalled() {
						return "installed but not running — start Docker Desktop or the Docker daemon"
					}
					return "" // genuinely absent: let the generic "not installed" wording stand
				},
			},
			toolPrerequisite("k3d", k3dInstaller.IsInstalled, k3dInstaller.Install, k3dInstaller.GetInstallHelp),
			toolPrerequisite("helm", helmInstaller.IsInstalled, helmInstaller.Install, helmInstaller.GetInstallHelp),
		},
	}
}

// toolPrerequisite adapts the uniform tool-installer API (IsInstalled/Install/
// GetInstallHelp) to a framework Prerequisite.
func toolPrerequisite(name string, isInstalled func() bool, install func() error, help func() string) fw.Prerequisite {
	return fw.Prerequisite{
		Name:        name,
		IsSatisfied: isInstalled,
		Install:     asCtxInstall(install),
		DocsURL:     help(),
	}
}

// asCtxInstall adapts a no-arg Install() to the framework's ctx-aware signature.
func asCtxInstall(install func() error) func(context.Context) error {
	return func(context.Context) error { return install() }
}
