package prerequisites

import (
	"context"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/aws"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/gcloud"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/terraform"
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

// EKSSet returns the prerequisites for EKS clusters: terraform (provisioning
// engine) and the AWS CLI (kubeconfig exec auth). Docker/k3d are deliberately
// absent — a cloud cluster needs no local runtime. AWS credentials are
// preflighted by the EKS provider itself, where the error can name the
// profile in use.
func EKSSet() fw.Set {
	terraformInstaller := terraform.NewTerraformInstaller()
	awsInstaller := aws.NewAwsInstaller()

	return fw.Set{
		Name: "eks",
		Items: []fw.Prerequisite{
			toolPrerequisite("terraform", terraformInstaller.IsInstalled, terraformInstaller.Install, terraformInstaller.GetInstallHelp),
			toolPrerequisite("AWS CLI", awsInstaller.IsInstalled, awsInstaller.Install, awsInstaller.GetInstallHelp),
		},
	}
}

// GKESet returns the prerequisites for GKE clusters: terraform (provisioning
// engine), the gcloud CLI, and gke-gcloud-auth-plugin (kubeconfig exec auth).
// GCP credentials are preflighted by the GKE provider itself, where the error
// can name the project in use.
func GKESet() fw.Set {
	terraformInstaller := terraform.NewTerraformInstaller()
	gcloudInstaller := gcloud.NewGcloudInstaller()
	authPluginInstaller := gcloud.NewAuthPluginInstaller()

	return fw.Set{
		Name: "gke",
		Items: []fw.Prerequisite{
			toolPrerequisite("terraform", terraformInstaller.IsInstalled, terraformInstaller.Install, terraformInstaller.GetInstallHelp),
			toolPrerequisite("gcloud", gcloudInstaller.IsInstalled, gcloudInstaller.Install, gcloudInstaller.GetInstallHelp),
			toolPrerequisite("gke-gcloud-auth-plugin", authPluginInstaller.IsInstalled, authPluginInstaller.Install, authPluginInstaller.GetInstallHelp),
		},
	}
}

// SetForClusterType maps a cluster type to its prerequisite set: Docker/k3d/
// helm for local k3d clusters, terraform + the cloud CLI for the cloud types.
// An empty type means the local default. Unknown types are an error — unlike
// the create-time gate there is no later provider factory to catch them here.
func SetForClusterType(clusterType models.ClusterType) (fw.Set, error) {
	switch clusterType {
	case models.ClusterTypeK3d, "":
		return ClusterSet(), nil
	case models.ClusterTypeEKS:
		return EKSSet(), nil
	case models.ClusterTypeGKE:
		return GKESet(), nil
	default:
		return fw.Set{}, fmt.Errorf("unknown cluster type %q (expected k3d, eks, or gke)", clusterType)
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
