package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/git"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/helm"
	chartUI "github.com/flamingo-stack/openframe-cli/internal/chart/ui"
	"github.com/flamingo-stack/openframe-cli/internal/chart/ui/configuration"
	"github.com/flamingo-stack/openframe-cli/internal/chart/ui/templates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/files"
	"github.com/pterm/pterm"
	"k8s.io/client-go/rest"
)

// ChartService handles high-level chart operations
type ChartService struct {
	executor       executor.CommandExecutor
	clusterService types.ClusterAccess
	configService  *config.Service
	operationsUI   *chartUI.OperationsUI
	displayService *chartUI.DisplayService
	helmManager    *helm.HelmManager
	gitRepository  *git.Repository
}

// NewChartService creates a new chart service with the given rest.Config
// The config is used to create the Kubernetes client for native API operations
func NewChartService(clusterAccess types.ClusterAccess, kubeConfig *rest.Config, dryRun, verbose bool) (*ChartService, error) {
	// Create executor
	chartExec := executor.NewRealCommandExecutor(dryRun, verbose)

	// Initialize configuration service
	configService := config.NewService()
	if err := configService.Initialize(); err != nil {
		pterm.Debug.Printf("config service initialization failed: %v\n", err)
	}

	// Create HelmManager with the rest.Config
	helmManager, err := helm.NewHelmManager(chartExec, kubeConfig, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create HelmManager: %w", err)
	}

	return &ChartService{
		executor:       chartExec,
		clusterService: clusterAccess,
		configService:  configService,
		operationsUI:   chartUI.NewOperationsUI(),
		displayService: chartUI.NewDisplayService(),
		helmManager:    helmManager,
		gitRepository:  git.NewRepository(),
	}, nil
}

// NewChartServiceDeferred creates a chart service without initializing HelmManager
// The HelmManager will be initialized later after cluster selection
func NewChartServiceDeferred(clusterAccess types.ClusterAccess, dryRun, verbose bool) (*ChartService, error) {
	// Create executor
	chartExec := executor.NewRealCommandExecutor(dryRun, verbose)

	// Initialize configuration service
	configService := config.NewService()
	if err := configService.Initialize(); err != nil {
		pterm.Debug.Printf("config service initialization failed: %v\n", err)
	}

	return &ChartService{
		executor:       chartExec,
		clusterService: clusterAccess,
		configService:  configService,
		operationsUI:   chartUI.NewOperationsUI(),
		displayService: chartUI.NewDisplayService(),
		helmManager:    nil, // Will be initialized after cluster selection
		gitRepository:  git.NewRepository(),
	}, nil
}

// initializeHelmManager initializes the HelmManager with the given rest.Config
// This is called after cluster selection in deferred mode
func (cs *ChartService) initializeHelmManager(kubeConfig *rest.Config, verbose bool) error {
	chartExec := executor.NewRealCommandExecutor(false, verbose)
	helmManager, err := helm.NewHelmManager(chartExec, kubeConfig, verbose)
	if err != nil {
		return fmt.Errorf("failed to create HelmManager: %w", err)
	}
	cs.helmManager = helmManager
	return nil
}

func (cs *ChartService) InstallWithContext(ctx context.Context, req types.InstallationRequest) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("chart installation cancelled: %w", ctx.Err())
	default:
	}

	// Create installation workflow with direct dependencies
	fileCleanup := files.NewFileCleanup()
	fileCleanup.SetCleanupOnSuccessOnly(true) // Only clean temporary files after successful ArgoCD sync

	workflow := &InstallationWorkflow{
		chartService:   cs,
		clusterService: cs.clusterService,
		fileCleanup:    fileCleanup,
	}

	// Execute workflow with context
	return workflow.ExecuteWithContext(ctx, req)
}

// InstallWithContextDeferred performs installation with deferred HelmManager initialization
// This is used when KubeConfig is not available upfront (e.g., standalone chart install)
func (cs *ChartService) InstallWithContextDeferred(ctx context.Context, req types.InstallationRequest) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("chart installation cancelled: %w", ctx.Err())
	default:
	}

	// Create installation workflow with direct dependencies
	fileCleanup := files.NewFileCleanup()
	fileCleanup.SetCleanupOnSuccessOnly(true) // Only clean temporary files after successful ArgoCD sync

	workflow := &InstallationWorkflow{
		chartService:   cs,
		clusterService: cs.clusterService,
		fileCleanup:    fileCleanup,
	}

	// Execute workflow with deferred initialization
	return workflow.ExecuteWithContextDeferred(ctx, req)
}

// InstallationWorkflow orchestrates the installation process
type InstallationWorkflow struct {
	chartService   *ChartService
	clusterService types.ClusterAccess
	fileCleanup    *files.FileCleanup
}

func (w *InstallationWorkflow) ExecuteWithContext(parentCtx context.Context, req types.InstallationRequest) error {
	// parentCtx is already signal-cancelled (the root runs via ExecuteContext),
	// so Ctrl-C / SIGTERM cancels it directly — no local signal handler needed.
	// A derived cancellable context lets us stop remaining work early.
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Step 1: Determine configuration mode and run appropriate workflow
	var chartConfig *types.ChartConfiguration
	if req.DryRun {
		var err error
		chartConfig, err = w.dryRunConfiguration()
		if err != nil {
			return err
		}
		// dry-run writes a real values file too, so register it for cleanup.
		if chartConfig.TempHelmValuesPath != "" {
			if backupErr := w.fileCleanup.RegisterTempFile(chartConfig.TempHelmValuesPath); backupErr != nil {
				pterm.Warning.Printf("Failed to register temp file for cleanup: %v\n", backupErr)
			}
		}
		pterm.Info.Println("Using existing configuration (dry-run mode)")
	} else if req.NonInteractive {
		// NON-INTERACTIVE (CI/CD): use the existing helm-values.yaml as-is.
		pterm.Warning.Println("Running in non-interactive mode using existing helm-values.yaml")
		var err error
		chartConfig, err = w.loadExistingConfiguration()
		if err != nil {
			return fmt.Errorf("non-interactive configuration failed: %w", err)
		}
	} else {
		// FULLY INTERACTIVE (existing behavior)
		var err error
		chartConfig, err = w.runConfigurationWizard()
		if err != nil {
			return fmt.Errorf("configuration wizard failed: %w", err)
		}

		// Register temporary file for cleanup
		if chartConfig.TempHelmValuesPath != "" {
			if backupErr := w.fileCleanup.RegisterTempFile(chartConfig.TempHelmValuesPath); backupErr != nil {
				pterm.Warning.Printf("Failed to register temp file for cleanup: %v\n", backupErr)
			}
		}
	}

	// Step 2: Select cluster
	clusterName, err := w.selectCluster(req.Args, req.Verbose)
	if err != nil || clusterName == "" {
		return err
	}

	// Step 3: Confirm installation on the selected cluster (skip in non-interactive mode)
	if !req.NonInteractive {
		if !w.confirmInstallationOnCluster(clusterName) {
			pterm.Info.Println("Installation cancelled.")
			return fmt.Errorf("installation cancelled by user")
		}
	}

	// Step 4: Regenerate certificates after configuration and cluster selection
	// Skip certificate regeneration in non-interactive mode
	if !req.NonInteractive {
		// Non-fatal: failures are logged inside the method, continue regardless.
		_ = w.regenerateCertificates()
	} else {
		pterm.Warning.Println("Skipping certificate regeneration (non-interactive mode)")
	}

	// Step 5: Build configuration
	config, err := w.buildConfiguration(req, clusterName, chartConfig)
	if err != nil {
		chartErr := errors.WrapAsChartError("configuration", "build", err).WithCluster(clusterName)
		return sharedErrors.HandleGlobalError(chartErr, req.Verbose)
	}

	// Step 6: Execute installation with retry support
	err = w.performInstallationWithRetry(ctx, config)

	// Step 7: Clean up generated files based on installation result
	if err != nil {
		// Installation failed - clean up temporary files immediately
		if cleanupErr := w.fileCleanup.RestoreFiles(req.Verbose); cleanupErr != nil {
			pterm.Warning.Printf("Failed to clean up files after error: %v\n", cleanupErr)
		}
		return err
	}

	// Check if cancelled by signal (CTRL-C) — the context is signal-cancelled.
	if ctx.Err() != nil {
		// User interrupted - clean up temporary files silently
		_ = w.fileCleanup.RestoreFiles(false) // Always clean up silently on interruption
		return fmt.Errorf("installation cancelled by user")
	}

	// Step 8: ArgoCD sync is already handled by installer.InstallCharts
	// The installer waits for all ArgoCD applications after installing app-of-apps

	// Step 9: Installation successful - clean up temporary files
	if cleanupErr := w.fileCleanup.RestoreFilesOnSuccess(req.Verbose); cleanupErr != nil {
		pterm.Warning.Printf("Failed to clean up files after successful installation: %v\n", cleanupErr)
	}

	return nil
}

// ExecuteWithContextDeferred runs the installation workflow with deferred HelmManager initialization
// This is used when KubeConfig is not available upfront (e.g., standalone chart install)
func (w *InstallationWorkflow) ExecuteWithContextDeferred(parentCtx context.Context, req types.InstallationRequest) error {
	// parentCtx is already signal-cancelled (root ExecuteContext); a derived
	// cancellable context is enough to stop remaining work on Ctrl-C / SIGTERM.
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Step 1: Determine configuration mode and run appropriate workflow
	var chartConfig *types.ChartConfiguration
	if req.DryRun {
		var err error
		chartConfig, err = w.dryRunConfiguration()
		if err != nil {
			return err
		}
		// dry-run writes a real values file too, so register it for cleanup.
		if chartConfig.TempHelmValuesPath != "" {
			if backupErr := w.fileCleanup.RegisterTempFile(chartConfig.TempHelmValuesPath); backupErr != nil {
				pterm.Warning.Printf("Failed to register temp file for cleanup: %v\n", backupErr)
			}
		}
		pterm.Info.Println("Using existing configuration (dry-run mode)")
	} else if req.NonInteractive {
		// NON-INTERACTIVE (CI/CD): use the existing helm-values.yaml as-is.
		pterm.Warning.Println("Running in non-interactive mode using existing helm-values.yaml")
		var err error
		chartConfig, err = w.loadExistingConfiguration()
		if err != nil {
			return fmt.Errorf("non-interactive configuration failed: %w", err)
		}
	} else {
		var err error
		chartConfig, err = w.runConfigurationWizard()
		if err != nil {
			return fmt.Errorf("configuration wizard failed: %w", err)
		}
		if chartConfig.TempHelmValuesPath != "" {
			if backupErr := w.fileCleanup.RegisterTempFile(chartConfig.TempHelmValuesPath); backupErr != nil {
				pterm.Warning.Printf("Failed to register temp file for cleanup: %v\n", backupErr)
			}
		}
	}

	// Step 2: Select cluster
	clusterName, err := w.selectCluster(req.Args, req.Verbose)
	if err != nil || clusterName == "" {
		return err
	}

	// Step 2.5: Get KubeConfig for the selected cluster and initialize HelmManager.
	// Resolved through the injected ClusterAccess interface so this workflow does
	// not depend on the concrete cluster service (req 18/19).
	kubeConfig, err := w.clusterService.GetRestConfig(clusterName)
	if err != nil {
		return fmt.Errorf("failed to get rest.Config for cluster %s: %w", clusterName, err)
	}
	if err := w.chartService.initializeHelmManager(kubeConfig, req.Verbose); err != nil {
		return fmt.Errorf("failed to initialize HelmManager: %w", err)
	}

	// Step 3: Confirm installation on the selected cluster (skip in non-interactive mode)
	if !req.NonInteractive {
		if !w.confirmInstallationOnCluster(clusterName) {
			pterm.Info.Println("Installation cancelled.")
			return fmt.Errorf("installation cancelled by user")
		}
	}

	// Step 4: Regenerate certificates
	if !req.NonInteractive {
		// Non-fatal: failures are logged inside the method, continue regardless.
		_ = w.regenerateCertificates()
	} else {
		pterm.Warning.Println("Skipping certificate regeneration (non-interactive mode)")
	}

	// Step 5: Build configuration
	config, err := w.buildConfiguration(req, clusterName, chartConfig)
	if err != nil {
		chartErr := errors.WrapAsChartError("configuration", "build", err).WithCluster(clusterName)
		return sharedErrors.HandleGlobalError(chartErr, req.Verbose)
	}

	// Step 6: Execute installation with retry support
	err = w.performInstallationWithRetry(ctx, config)

	// Step 7: Clean up generated files based on installation result
	if err != nil {
		if cleanupErr := w.fileCleanup.RestoreFiles(req.Verbose); cleanupErr != nil {
			pterm.Warning.Printf("Failed to clean up files after error: %v\n", cleanupErr)
		}
		return err
	}

	if ctx.Err() != nil {
		_ = w.fileCleanup.RestoreFiles(false)
		return fmt.Errorf("installation cancelled by user")
	}

	if cleanupErr := w.fileCleanup.RestoreFilesOnSuccess(req.Verbose); cleanupErr != nil {
		pterm.Warning.Printf("Failed to clean up files after successful installation: %v\n", cleanupErr)
	}

	return nil
}

// selectCluster handles cluster selection
func (w *InstallationWorkflow) selectCluster(args []string, verbose bool) (string, error) {
	clusterSelector := NewClusterSelector(w.clusterService, w.chartService.operationsUI)
	return clusterSelector.SelectCluster(args, verbose)
}

// confirmInstallationOnCluster prompts for user confirmation with specific cluster name
func (w *InstallationWorkflow) confirmInstallationOnCluster(clusterName string) bool {
	confirmed, err := w.chartService.operationsUI.ConfirmInstallationOnCluster(clusterName)
	if err != nil {
		// Treat a prompt error/interruption as "not confirmed"; the caller turns
		// that into a cancellation error that exits non-zero (no os.Exit here).
		return false
	}
	return confirmed
}

// regenerateCertificates refreshes certificates after user confirmation
func (w *InstallationWorkflow) regenerateCertificates() error {
	installer := prerequisites.NewInstaller()
	return installer.RegenerateCertificatesOnly()
}

// runConfigurationWizard runs the configuration wizard to get user preferences
func (w *InstallationWorkflow) runConfigurationWizard() (*types.ChartConfiguration, error) {
	wizard := configuration.NewConfigurationWizard()

	// Configure Helm values from current directory
	config, err := wizard.ConfigureHelmValues()
	if err != nil {
		return nil, fmt.Errorf("helm values configuration failed: %w", err)
	}

	return config, nil
}

// loadExistingConfiguration loads existing helm-values.yaml for non-interactive mode
// dryRunConfiguration builds the chart configuration for a dry-run. Like every
// other mode, it writes the base helm values to a real temporary file and
// points TempHelmValuesPath at it — previously dry-run set a fixed
// "helm-values-tmp.yaml" that nothing ever wrote, so the app-of-apps step ran
// `helm --dry-run -f helm-values-tmp.yaml` against a non-existent file. The
// caller registers the returned path for cleanup.
func (w *InstallationWorkflow) dryRunConfiguration() (*types.ChartConfiguration, error) {
	modifier := templates.NewHelmValuesModifier()
	baseValues, err := modifier.LoadOrCreateBaseValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load base values for dry-run: %w", err)
	}
	tempFilePath, err := modifier.CreateTemporaryValuesFile(baseValues)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary values file for dry-run: %w", err)
	}
	return &types.ChartConfiguration{
		BaseHelmValuesPath: "helm-values.yaml",
		TempHelmValuesPath: tempFilePath,
		ExistingValues:     baseValues,
		ModifiedSections:   make([]string, 0),
	}, nil
}

func (w *InstallationWorkflow) loadExistingConfiguration() (*types.ChartConfiguration, error) {
	modifier := templates.NewHelmValuesModifier()

	// Load existing helm-values.yaml
	values, err := modifier.LoadOrCreateBaseValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load helm-values.yaml: %w", err)
	}

	// Create temporary file from the existing values (same as interactive mode)
	tempFilePath, err := modifier.CreateTemporaryValuesFile(values)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary values file: %w", err)
	}

	result := &types.ChartConfiguration{
		BaseHelmValuesPath: "helm-values.yaml",
		TempHelmValuesPath: tempFilePath, // Use temporary file like interactive mode
		ExistingValues:     values,
		ModifiedSections:   []string{},
	}

	// Validate the OSS configuration (no-op today: OSS uses a public repo).
	validator := NewConfigurationValidator()
	if err := validator.ValidateConfiguration(result); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return result, nil
}

// buildConfiguration constructs the installation configuration
func (w *InstallationWorkflow) buildConfiguration(req types.InstallationRequest, clusterName string, chartConfig *types.ChartConfiguration) (config.ChartInstallConfig, error) {
	configBuilder := config.NewBuilder(w.chartService.operationsUI)

	// Only the OSS (oss-tenant) deployment is supported: use the request's repo
	// (defaults to the public OSS repository) with no embedded credentials.
	githubRepo := req.GitHubRepo

	// When the operator explicitly pins a ref (--ref/--github-branch), write it into
	// the temp helm-values' repository.branch BEFORE the builder reads it back. This
	// makes the explicit ref win over the values-file branch and keeps BOTH the
	// app-of-apps clone and the child Applications' targetRevision on that ref
	// (otherwise the values-file branch silently overrides --ref).
	if ref := strings.TrimSpace(req.GitHubBranch); req.GitHubRefExplicit && ref != "" && chartConfig.TempHelmValuesPath != "" {
		modifier := templates.NewHelmValuesModifier()
		values := chartConfig.ExistingValues
		if values == nil {
			loaded, lerr := modifier.LoadExistingValues(chartConfig.TempHelmValuesPath)
			if lerr != nil {
				return config.ChartInstallConfig{}, fmt.Errorf("pinning ref %q: %w", ref, lerr)
			}
			values = loaded
		}
		modifier.SetRepositoryBranch(values, ref)
		if werr := modifier.WriteValues(values, chartConfig.TempHelmValuesPath); werr != nil {
			return config.ChartInstallConfig{}, fmt.Errorf("pinning ref %q into helm values: %w", ref, werr)
		}
		pterm.Info.Printf("Pinning platform to ref %q\n", ref)
	}

	return configBuilder.BuildInstallConfigWithCustomHelmPath(
		req.Force, req.DryRun, req.Verbose, req.NonInteractive, clusterName,
		githubRepo, req.GitHubBranch, req.CertDir,
		chartConfig.TempHelmValuesPath,
	)
}

// performInstallation executes the actual installation
func (w *InstallationWorkflow) performInstallation(ctx context.Context, config config.ChartInstallConfig) error {
	// Create installer directly without factory
	pathResolver := w.chartService.configService.GetPathResolver()
	argoCDService := NewArgoCD(w.chartService.helmManager, pathResolver, w.chartService.executor)
	appOfAppsService := NewAppOfApps(w.chartService.helmManager, w.chartService.gitRepository, pathResolver)

	installer := &Installer{
		argoCDService:    argoCDService,
		appOfAppsService: appOfAppsService,
	}

	err := installer.InstallChartsWithContext(ctx, config)
	if err != nil {
		// Check if this is a branch not found error
		var bnfErr *sharedErrors.BranchNotFoundError
		if stderrors.As(err, &bnfErr) {
			return err // Return as-is, don't wrap
		}
		return errors.WrapAsChartError("installation", "chart", err).WithCluster(config.ClusterName)
	}
	return nil
}

// performInstallationWithRetry executes installation with retry policy
func (w *InstallationWorkflow) performInstallationWithRetry(parentCtx context.Context, config config.ChartInstallConfig) error {
	retryPolicy := sharedErrors.InstallationRetryPolicy()
	retryExecutor := sharedErrors.NewRetryExecutor(retryPolicy)
	// No retry callback - let the spinner handle progress indication

	// Combine parent context (for CTRL-C) with timeout
	ctx, cancel := context.WithTimeout(parentCtx, 60*time.Minute)
	defer cancel()

	return retryExecutor.Execute(ctx, func() error {
		// Check if cancelled before attempting installation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return w.performInstallation(ctx, config)
	})
}

// InstallChartsWithConfigContext installs charts with the given configuration and context support
// If KubeConfig is nil, it will be obtained after cluster selection (for standalone chart install)
func InstallChartsWithConfigContext(ctx context.Context, req types.InstallationRequest) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("chart installation cancelled: %w", ctx.Err())
	default:
	}

	// Check prerequisites first
	installer := prerequisites.NewInstaller()
	if err := installer.CheckAndInstallNonInteractive(req.NonInteractive); err != nil {
		return err
	}

	// Check context again after prerequisites
	select {
	case <-ctx.Done():
		return fmt.Errorf("chart installation cancelled: %w", ctx.Err())
	default:
	}

	// If KubeConfig is provided (e.g., from bootstrap), use it directly
	// Otherwise, defer to the chart service to get it after cluster selection
	if req.KubeConfig != nil {
		// Create a chart service with the KubeConfig and perform the installation with context
		chartService, err := NewChartService(req.ClusterAccess, req.KubeConfig, req.DryRun, req.Verbose)
		if err != nil {
			return fmt.Errorf("failed to create chart service: %w", err)
		}
		return chartService.InstallWithContext(ctx, req)
	}

	// No KubeConfig provided - use deferred initialization. This path selects a
	// cluster and resolves its rest.Config, so it needs cluster access injected
	// by the caller (req 18/19 keeps this out of internal/cluster).
	if req.ClusterAccess == nil {
		return fmt.Errorf("cluster access is required to install without an explicit kubeconfig")
	}
	chartService, err := NewChartServiceDeferred(req.ClusterAccess, req.DryRun, req.Verbose)
	if err != nil {
		return fmt.Errorf("failed to create chart service: %w", err)
	}
	return chartService.InstallWithContextDeferred(ctx, req)
}
