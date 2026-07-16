package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/git"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
)

// AppOfApps handles app-of-apps installation logic
type AppOfApps struct {
	helmManager  *helm.HelmManager
	gitRepo      *git.Repository
	pathResolver *config.PathResolver
}

// NewAppOfApps creates a new app-of-apps service
func NewAppOfApps(helmManager *helm.HelmManager, gitRepo *git.Repository, pathResolver *config.PathResolver) *AppOfApps {
	return &AppOfApps{
		helmManager:  helmManager,
		gitRepo:      gitRepo,
		pathResolver: pathResolver,
	}
}

// Install installs app-of-apps from GitHub repository using git clone
func (a *AppOfApps) Install(ctx context.Context, config config.ChartInstallConfig) error {
	// Validate configuration
	if config.AppOfApps == nil {
		return errors.NewValidationError("app-of-apps", "nil", "configuration is required")
	}

	appConfig := config.AppOfApps
	if appConfig.GitHubRepo == "" {
		return errors.NewValidationError("GitHubRepo", "empty", "GitHub repository URL is required for app-of-apps installation")
	}
	if appConfig.GitHubBranch == "" {
		appConfig.GitHubBranch = "main" // Default to main branch
	}

	// Say what is being DEPLOYED (the resolved ref), not "used" — the old
	// wording read as if it reflected the cluster's current ref, which made a
	// dry-run against a cluster on another ref confusing (verification report,
	// minor observation).
	pterm.Info.Printf("Deploying ref '%s'...\n", appConfig.GitHubBranch)

	// Clone the repository to a temporary directory. On a cold cache this is a
	// full clone over the network and used to run without any indicator.
	var cloneSpinner *spinner.Spinner
	if !config.Silent && !config.NonInteractive {
		cloneSpinner = spinner.Start(fmt.Sprintf("Cloning the OpenFrame chart repository (ref %s)...", appConfig.GitHubBranch))
	}
	cloneResult, err := a.gitRepo.CloneChartRepository(ctx, appConfig)
	if err != nil {
		if cloneSpinner != nil {
			cloneSpinner.Fail("Could not clone the chart repository")
		}
		// Check if this is a branch not found error
		if strings.Contains(err.Error(), "branch") && strings.Contains(err.Error(), "does not exist") {
			// Return the proper error type
			return sharedErrors.NewBranchNotFoundError(appConfig.GitHubBranch)
		}
		return errors.NewRecoverableChartError("clone", "Git repository", err, 10*time.Second).WithCluster(config.ClusterName)
	}
	if cloneSpinner != nil {
		cloneSpinner.Success("Chart repository cloned")
	}

	// Ensure cleanup happens after installation completes (success or failure)
	defer func() {
		a.gitRepo.Cleanup(cloneResult.TempDir)
	}()

	// Get file paths
	valuesFile := a.pathResolver.GetHelmValuesFile()
	if appConfig.ValuesFile != "" {
		valuesFile = appConfig.ValuesFile
	}

	certFile, keyFile := a.pathResolver.GetCertificateFiles()

	// Create a modified config with the local chart path
	// Deep copy the AppOfApps config to avoid modifying the original
	localAppOfApps := *config.AppOfApps
	localAppOfApps.ChartPath = cloneResult.ChartPath
	localAppOfApps.ValuesFile = valuesFile
	localConfig := config
	localConfig.AppOfApps = &localAppOfApps

	// Show details only in verbose mode
	if config.Verbose {
		pterm.Info.Printf("   Chart path: %s\n", cloneResult.ChartPath)
		pterm.Info.Printf("   Values file: %s\n", valuesFile)
	}

	// Use helm manager to install app-of-apps
	err = a.helmManager.InstallAppOfAppsFromLocal(ctx, localConfig, certFile, keyFile)
	if err != nil {
		return errors.WrapAsChartError("installation", "app-of-apps", err).WithCluster(config.ClusterName)
	}

	return nil
}

// IsInstalled checks if app-of-apps is installed
func (a *AppOfApps) IsInstalled(ctx context.Context, namespace string) (bool, error) {
	return a.helmManager.IsChartInstalled(ctx, "app-of-apps", namespace)
}

// GetStatus returns the status of app-of-apps installation
func (a *AppOfApps) GetStatus(ctx context.Context, namespace string) (models.ChartInfo, error) {
	return a.helmManager.GetChartStatus(ctx, "app-of-apps", namespace)
}
