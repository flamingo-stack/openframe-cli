package prerequisites

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
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
}

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
			},
			{
				Name:        "k3d",
				Command:     "k3d",
				IsInstalled: func() bool { return k3d.NewK3dInstaller().IsInstalled() },
				InstallHelp: func() string { return k3d.NewK3dInstaller().GetInstallHelp() },
			},
			{
				Name:        "helm",
				Command:     "helm",
				IsInstalled: func() bool { return helm.NewHelmInstaller().IsInstalled() },
				InstallHelp: func() string { return helm.NewHelmInstaller().GetInstallHelp() },
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

// CheckForClusterType runs the prerequisite gate for the given cluster type.
// Docker/k3d/helm are only required for local k3d clusters; cloud types bring
// their own prerequisite sets with their backends (none implemented yet), so
// they pass through and fail later at the provider factory instead of
// demanding tools they will never use.
func CheckForClusterType(clusterType models.ClusterType) error {
	switch clusterType {
	case models.ClusterTypeK3d, "":
		return CheckPrerequisites()
	default:
		return nil
	}
}
