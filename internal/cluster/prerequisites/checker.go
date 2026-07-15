package prerequisites

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/aws"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

type PrerequisiteChecker struct {
	requirements []Requirement
}

type Requirement struct {
	Name        string
	Command     string
	IsInstalled func() bool
	InstallHelp func() string
	Install     func() error
}

// NewPrerequisiteChecker returns the requirement set for local k3d clusters.
func NewPrerequisiteChecker() *PrerequisiteChecker {
	return &PrerequisiteChecker{
		requirements: []Requirement{
			{
				Name:        "Docker",
				Command:     "docker",
				IsInstalled: func() bool { return docker.IsDockerRunning() },
				InstallHelp: func() string {
					if !docker.NewDockerInstaller().IsInstalled() {
						return docker.NewDockerInstaller().GetInstallHelp()
					}
					return "Docker is installed but not running. Please start Docker Desktop or the Docker daemon."
				},
				Install: func() error { return docker.NewDockerInstaller().Install() },
			},
			{
				Name:        "k3d",
				Command:     "k3d",
				IsInstalled: func() bool { return k3d.NewK3dInstaller().IsInstalled() },
				InstallHelp: func() string { return k3d.NewK3dInstaller().GetInstallHelp() },
				Install:     func() error { return k3d.NewK3dInstaller().Install() },
			},
			{
				Name:        "helm",
				Command:     "helm",
				IsInstalled: func() bool { return helm.NewHelmInstaller().IsInstalled() },
				InstallHelp: func() string { return helm.NewHelmInstaller().GetInstallHelp() },
				Install:     func() error { return helm.NewHelmInstaller().Install() },
			},
		},
	}
}

// NewEKSPrerequisiteChecker returns the requirement set for EKS clusters:
// terraform (provisioning engine) and the AWS CLI (kubeconfig exec auth).
// Docker/k3d are deliberately absent — a cloud cluster needs no local runtime.
// AWS credentials are preflighted by the EKS provider itself, where the error
// can name the profile in use.
func NewEKSPrerequisiteChecker() *PrerequisiteChecker {
	return &PrerequisiteChecker{
		requirements: []Requirement{
			{
				Name:        "terraform",
				Command:     "terraform",
				IsInstalled: func() bool { return terraform.NewTerraformInstaller().IsInstalled() },
				InstallHelp: func() string { return terraform.NewTerraformInstaller().GetInstallHelp() },
				Install:     func() error { return terraform.NewTerraformInstaller().Install() },
			},
			{
				Name:        "AWS CLI",
				Command:     "aws",
				IsInstalled: func() bool { return aws.NewAwsInstaller().IsInstalled() },
				InstallHelp: func() string { return aws.NewAwsInstaller().GetInstallHelp() },
				Install:     func() error { return aws.NewAwsInstaller().Install() },
			},
		},
	}
}

func (pc *PrerequisiteChecker) CheckAll() (bool, []string) {
	var missing []string
	allPresent := true

	for _, req := range pc.requirements {
		if !req.IsInstalled() {
			missing = append(missing, req.Name)
			allPresent = false
		}
	}

	return allPresent, missing
}

func CheckPrerequisites() error {
	// A CI environment or a non-terminal stdin must not hit an interactive prompt.
	return NewInstaller().CheckAndInstallNonInteractive(ui.IsNonInteractive())
}

// CheckForClusterType runs the prerequisite gate for the given cluster type:
// Docker/k3d/helm for local k3d clusters, terraform + AWS CLI for EKS. GKE has
// no backend yet, so it passes through and fails at the provider factory
// instead of demanding tools it will never use.
func CheckForClusterType(clusterType models.ClusterType) error {
	switch clusterType {
	case models.ClusterTypeK3d, "":
		return CheckPrerequisites()
	case models.ClusterTypeEKS:
		return NewInstallerWithChecker(NewEKSPrerequisiteChecker()).CheckAndInstallNonInteractive(ui.IsNonInteractive())
	default:
		return nil
	}
}
