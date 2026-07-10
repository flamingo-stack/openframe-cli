package config

import (
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	chartUI "github.com/flamingo-stack/openframe-cli/internal/chart/ui"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

// Builder handles construction of installation configurations
type Builder struct {
	configService *Service
	operationsUI  *chartUI.OperationsUI
}

// NewBuilder creates a new configuration builder
func NewBuilder(operationsUI *chartUI.OperationsUI) *Builder {
	return &Builder{
		configService: NewService(),
		operationsUI:  operationsUI,
	}
}

// helmValues is the subset of the FLATTENED chart schema this builder consumes:
// the app-of-apps source ref lives at top-level repository.branch — the same key
// HelmValuesModifier writes (SetRepositoryBranch) and the chart itself reads.
// This struct is the single source of truth for branch resolution. The old
// nested deployment.oss.repository.branch schema is deliberately gone: reading
// it here ignored the branch the rest of the pipeline used and let stale
// legacy-schema files silently override an explicit --ref (audit F1/T1-1).
type helmValues struct {
	Repository struct {
		Branch string `yaml:"branch"`
	} `yaml:"repository"`
}

// getBranchFromHelmValues reads the Helm values file and extracts the repository branch
func (b *Builder) getBranchFromHelmValues() string {
	return b.getBranchFromHelmValuesPath("")
}

// getBranchFromHelmValuesPath reads a specific Helm values file and extracts the
// flattened repository.branch (empty means "use the default/flag ref").
func (b *Builder) getBranchFromHelmValuesPath(helmValuesPath string) string {
	if helmValuesPath == "" {
		pathResolver := NewPathResolver()
		helmValuesPath = pathResolver.GetHelmValuesFile()
	}

	// Read the YAML file
	data, err := os.ReadFile(helmValuesPath) // #nosec G304 -- helm values path resolved from config/CLI, read as invoking user
	if err != nil {
		// If we can't read the file, return empty string (will use default)
		return ""
	}

	var values helmValues
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		// If we can't parse the YAML, return empty string (will use default)
		return ""
	}

	return values.Repository.Branch
}

// BuildInstallConfig constructs the installation configuration
func (b *Builder) BuildInstallConfig(
	force, dryRun, verbose bool,
	clusterName, githubRepo, githubBranch, certDir string,
) (ChartInstallConfig, error) {
	// Use config service for certificate directory
	if certDir == "" {
		certDir = b.configService.GetCertificateDirectory()
	}

	// Create app-of-apps configuration if GitHub repo is provided
	var appOfAppsConfig *models.AppOfAppsConfig
	if githubRepo != "" {
		appOfAppsConfig = models.NewAppOfAppsConfig()
		appOfAppsConfig.GitHubRepo = githubRepo
		appOfAppsConfig.GitHubBranch = githubBranch
		appOfAppsConfig.CertDir = certDir

		// Repository is public, no credentials needed

		// After credentials are provided, check for branch override from Helm values
		helmBranch := b.getBranchFromHelmValues()
		if helmBranch != "" {
			if verbose {
				pterm.Info.Printf("📥 Using branch '%s' from Helm values\n", helmBranch)
			}
			appOfAppsConfig.GitHubBranch = helmBranch
		} else if verbose {
			pterm.Info.Printf("📥 Using default branch '%s'\n", appOfAppsConfig.GitHubBranch)
		}
	}

	return b.configService.BuildInstallConfig(
		force, dryRun, verbose,
		clusterName,
		appOfAppsConfig,
	), nil
}

// BuildInstallConfigWithCustomHelmPath constructs the installation configuration using a custom helm values file
func (b *Builder) BuildInstallConfigWithCustomHelmPath(
	force, dryRun, verbose, nonInteractive bool,
	clusterName, githubRepo, githubBranch, certDir, helmValuesPath string,
) (ChartInstallConfig, error) {
	// Use config service for certificate directory
	if certDir == "" {
		certDir = b.configService.GetCertificateDirectory()
	}

	// Create app-of-apps configuration if GitHub repo is provided
	var appOfAppsConfig *models.AppOfAppsConfig
	if githubRepo != "" {
		appOfAppsConfig = models.NewAppOfAppsConfig()
		appOfAppsConfig.GitHubRepo = githubRepo
		appOfAppsConfig.GitHubBranch = githubBranch
		appOfAppsConfig.CertDir = certDir

		// Repository is public, no credentials needed

		// Set the custom helm values file path if provided
		if helmValuesPath != "" {
			appOfAppsConfig.ValuesFile = helmValuesPath
		}

		// Check for a branch override from the custom Helm values path
		// (flattened schema: top-level repository.branch). When --ref was
		// explicit, buildConfiguration already pinned it into this file, so
		// reading it back here keeps the clone and the child Applications'
		// targetRevision on the same ref.
		helmBranch := b.getBranchFromHelmValuesPath(helmValuesPath)
		if helmBranch != "" {
			if verbose {
				pterm.Info.Printf("📥 Using branch '%s' from Helm values\n", helmBranch)
			}
			appOfAppsConfig.GitHubBranch = helmBranch
		} else if verbose {
			pterm.Info.Printf("📥 Using default branch '%s'\n", appOfAppsConfig.GitHubBranch)
		}
	}

	config := b.configService.BuildInstallConfig(
		force, dryRun, verbose,
		clusterName,
		appOfAppsConfig,
	)

	// Set Silent flag based on NonInteractive mode
	config.Silent = nonInteractive
	config.NonInteractive = nonInteractive
	// CRDs are installed by the Argo CD Helm chart itself (crds.install=true).
	// SkipCRDs is retained only so the readiness check still waits for the
	// Application CRD to appear before app-of-apps runs.
	config.SkipCRDs = false

	return config, nil
}
