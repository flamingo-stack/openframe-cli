package services

import (
	"context"
	stderrors "errors"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
)

// Installer orchestrates the chart installation process
type Installer struct {
	argoCDService    types.ArgoCDService
	appOfAppsService types.AppOfAppsService
	// refValidator preflights the chart ref before anything touches the
	// cluster; nil skips the preflight (tests, callers without a repo).
	refValidator types.GitRefValidator
}

// InstallChartsWithContext handles the complete chart installation process with context support
func (i *Installer) InstallChartsWithContext(ctx context.Context, config config.ChartInstallConfig) error {
	// A bad --ref must fail HERE, in seconds — not after ArgoCD has been
	// installed: the clone was the first place a typo'd ref surfaced, leaving
	// the cluster with ArgoCD deployed and no applications.
	if config.HasAppOfApps() && i.refValidator != nil {
		appConfig := *config.AppOfApps
		if appConfig.GitHubBranch == "" {
			appConfig.GitHubBranch = "main" // mirror the app-of-apps default
		}
		if err := i.refValidator.ValidateRef(ctx, &appConfig); err != nil {
			var bnfErr *sharedErrors.BranchNotFoundError
			if stderrors.As(err, &bnfErr) {
				return err // renders its own actionable panel; don't wrap
			}
			return errors.WrapAsChartError("preflight", "chart repository", err).WithCluster(config.ClusterName)
		}
	}

	// Install ArgoCD first
	if err := i.argoCDService.Install(ctx, config); err != nil {
		return errors.WrapAsChartError("installation", "ArgoCD", err).WithCluster(config.ClusterName)
	}

	// Install app-of-apps from GitHub repository if configured
	if config.HasAppOfApps() {
		if err := i.appOfAppsService.Install(ctx, config); err != nil {
			// Check if this is a branch not found error
			var bnfErr *sharedErrors.BranchNotFoundError
			if stderrors.As(err, &bnfErr) {
				return err // Return as-is, don't wrap
			}
			return errors.WrapAsChartError("installation", "app-of-apps", err).WithCluster(config.ClusterName)
		}

		// Wait for all ArgoCD applications to be ready after app-of-apps installation
		// Note: This is NOT a recoverable error - ArgoCD and app-of-apps are already installed,
		// so retrying would reinstall them unnecessarily. WaitForApplications has its own internal retry logic.
		if err := i.argoCDService.WaitForApplications(ctx, config); err != nil {
			// Create a new non-recoverable error (don't use WrapAsChartError which preserves existing ChartError's Recoverable flag).
			// The service layer already wrapped with the same operation/component;
			// rewrap its CAUSE, not the wrapper, or the message doubles up:
			// "waiting failed for ArgoCD applications on cluster X: waiting failed for ..."
			cause := err
			var ce *errors.ChartError
			if stderrors.As(err, &ce) && ce.Operation == "waiting" && ce.Component == "ArgoCD applications" {
				cause = ce.Cause
			}
			// WithNonRetryable, not just default-false: the wait diagnostics
			// embed pod logs whose text ("connection refused", "i/o timeout")
			// would otherwise pattern-match the retry policy's transient table.
			return errors.NewChartError("waiting", "ArgoCD applications", cause).WithCluster(config.ClusterName).WithNonRetryable()
		}
	}

	return nil
}
