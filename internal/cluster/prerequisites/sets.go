package prerequisites

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/kubectl"
	fw "github.com/flamingo-stack/openframe-cli/internal/prerequisites"
)

// ClusterSet returns the prerequisites required before creating or managing a
// local (k3d) cluster, expressed against the shared prerequisites framework:
// Docker (running), kubectl, k3d, and helm.
//
// On macOS/Linux the framework auto-installs any that are missing; on Windows it
// reports each missing tool with its manual setup guidance.
func ClusterSet() fw.Set {
	dockerInstaller := docker.NewDockerInstaller()
	kubectlInstaller := kubectl.NewKubectlInstaller()
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
			},
			toolPrerequisite("kubectl", kubectlInstaller.IsInstalled, kubectlInstaller.Install, kubectlInstaller.GetInstallHelp),
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
