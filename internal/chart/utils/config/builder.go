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

// HelmValues represents the structure of the Helm values file
type HelmValues struct {
	Deployment struct {
		OSS struct {
			Enabled    bool `yaml:"enabled"`
			Repository struct {
				Branch string `yaml:"branch"`
			} `yaml:"repository"`
		} `yaml:"oss"`
	} `yaml:"deployment"`
}

// getBranchFromHelmValues reads the Helm values file and extracts the OSS branch
func (b *Builder) getBranchFromHelmValues() string {
	return b.getBranchFromHelmValuesPath("")
}

// getBranchFromHelmValuesPath reads a specific Helm values file and extracts the OSS repository branch
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

	var values HelmValues
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		// If we can't parse the YAML, return empty string (will use default)
		return ""
	}

	if values.Deployment.OSS.Repository.Branch != "" {
		return values.Deployment.OSS.Repository.Branch
	}

	return "" // Return empty string if no branch found
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
		// (OSS Tenant: deployment.oss.repository.branch).
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
