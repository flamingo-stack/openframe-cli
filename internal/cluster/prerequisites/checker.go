package prerequisites

import (
	"os"
	"runtime"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/kubectl"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/wsl"
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
	requirements := []Requirement{}

	// On native Windows (not inside WSL), check for WSL first
	// since all other tools depend on it
	wslInstaller := wsl.NewWSLInstaller()
	if wslInstaller.IsApplicable() {
		requirements = append(requirements, Requirement{
			Name:        "WSL2",
			Command:     "wsl",
			IsInstalled: func() bool { return wslInstaller.IsInstalled() },
			InstallHelp: func() string { return wslInstaller.GetInstallHelp() },
		})
	}

	// Add common requirements
	requirements = append(requirements,
		Requirement{
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
		Requirement{
			Name:        "kubectl",
			Command:     "kubectl",
			IsInstalled: func() bool { return kubectl.NewKubectlInstaller().IsInstalled() },
			InstallHelp: func() string { return kubectl.NewKubectlInstaller().GetInstallHelp() },
		},
		Requirement{
			Name:        "k3d",
			Command:     "k3d",
			IsInstalled: func() bool { return k3d.NewK3dInstaller().IsInstalled() },
			InstallHelp: func() string { return k3d.NewK3dInstaller().GetInstallHelp() },
		},
		Requirement{
			Name:        "helm",
			Command:     "helm",
			IsInstalled: func() bool { return helm.NewHelmInstaller().IsInstalled() },
			InstallHelp: func() string { return helm.NewHelmInstaller().GetInstallHelp() },
		},
	)

	return &PrerequisiteChecker{
		requirements: requirements,
	}
}

// IsWindowsNative returns true if running on native Windows (not inside WSL)
func IsWindowsNative() bool {
	return runtime.GOOS == "windows" && !wsl.IsRunningInWSL()
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

func (pc *PrerequisiteChecker) GetInstallInstructions(missingTools []string) []string {
	var instructions []string

	for _, tool := range missingTools {
		for _, req := range pc.requirements {
			if strings.EqualFold(req.Name, tool) {
				instructions = append(instructions, req.InstallHelp())
				break
			}
		}
	}

	return instructions
}

func CheckPrerequisites() error {
	installer := NewInstaller()
	// Check if we're in a CI environment (GitHub Actions, GitLab CI, CircleCI, etc.)
	nonInteractive := os.Getenv("CI") != "" ||
		os.Getenv("GITHUB_ACTIONS") != "" ||
		os.Getenv("GITLAB_CI") != "" ||
		os.Getenv("CIRCLECI") != ""

	return installer.CheckAndInstallNonInteractive(nonInteractive)
}
