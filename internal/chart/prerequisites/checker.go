package prerequisites

import (
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/certificates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/memory"
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
				Name:        "Helm",
				Command:     "helm",
				IsInstalled: func() bool { return helm.NewHelmInstaller().IsInstalled() },
				InstallHelp: func() string { return helm.NewHelmInstaller().GetInstallHelp() },
			},
			{
				Name:        "Memory",
				Command:     "memory",
				IsInstalled: func() bool { return memory.NewMemoryChecker().IsInstalled() },
				InstallHelp: func() string { return memory.NewMemoryChecker().GetInstallHelp() },
			},
			{
				Name:        "Certificates",
				Command:     "certificates",
				IsInstalled: func() bool { return certificates.NewCertificateInstaller().IsInstalled() },
				InstallHelp: func() string { return certificates.NewCertificateInstaller().GetInstallHelp() },
			},
		},
	}
}

// CheckAll verifies whether all registered requirements are installed.
// It returns true along with a nil slice if everything is present, or
// false along with the names of the missing requirements otherwise.
func (pc *PrerequisiteChecker) CheckAll() (bool, []string) {
	missing := make([]string, 0, len(pc.requirements))
	for _, req := range pc.requirements {
		if !req.IsInstalled() {
			missing = append(missing, req.Name)
		}
	}

	if len(missing) == 0 {
		return true, nil
	}
	return false, missing
}
