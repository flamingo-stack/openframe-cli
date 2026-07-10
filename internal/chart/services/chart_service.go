package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
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
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
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
	// kubeConfig is the rest.Config the HelmManager was built with — the single
	// install target. The ArgoCD wait manager is constructed from it too, so the
	// helm CLI, the native checks, and the readiness wait all watch the same
	// cluster (audit F4).
	kubeConfig *rest.Config
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
		kubeConfig:     kubeConfig,
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
	cs.kubeConfig = kubeConfig
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

// InstallWithContextDeferred performs installation with deferred HelmManager
// initialization — used when KubeConfig is not available upfront (standalone
// chart install). Same workflow as InstallWithContext: the nil HelmManager on a
// service built by NewChartServiceDeferred triggers the in-workflow resolution.
func (cs *ChartService) InstallWithContextDeferred(ctx context.Context, req types.InstallationRequest) error {
	return cs.InstallWithContext(ctx, req)
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
		// NON-INTERACTIVE (CI/CD). Which values are used (existing file vs
		// chart defaults) is announced by loadExistingConfiguration — claiming
		// "using existing openframe-helm-values.yaml" here contradicted the
		// missing-file warning two lines later (verification finding N1).
		pterm.Info.Println("Running in non-interactive mode")
		var err error
		chartConfig, err = w.loadExistingConfiguration(req.RequireExistingValues)
		if err != nil {
			return fmt.Errorf("non-interactive configuration failed: %w", err)
		}
		// Register the temp values file for cleanup (the dry-run and interactive
		// paths do the same); otherwise the OS temp dir accumulates one per run.
		if chartConfig.TempHelmValuesPath != "" {
			if backupErr := w.fileCleanup.RegisterTempFile(chartConfig.TempHelmValuesPath); backupErr != nil {
				pterm.Warning.Printf("Failed to register temp file for cleanup: %v\n", backupErr)
			}
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

	// Step 2: Resolve the install target. An explicit rest.Config from the
	// command layer (--context, or the interactive kube-context selector) IS
	// the target — running k3d cluster selection on top of it demanded a
	// cluster name that --context had already made redundant (verification
	// finding N2: `app install -c <ctx> --non-interactive` was unusable) and
	// double-prompted interactive users (kube-context, then k3d cluster).
	var clusterName string
	if req.KubeConfig == nil {
		var err error
		clusterName, err = w.selectCluster(req.Args, req.NonInteractive, req.Verbose)
		if err != nil {
			return err
		}
		if clusterName == "" {
			// selectCluster prints why (no clusters found, or the interactive
			// selection was cancelled) but returns no error; surface a non-zero exit
			// so callers and CI don't read a no-op install as success.
			return fmt.Errorf("no cluster selected — nothing was installed")
		}
	} else if req.KubeContext != "" {
		// ClusterName stays empty: every helm call targets req.KubeContext
		// (helmKubeContext gives it precedence) and the ArgoCD wait manager is
		// built from the same rest.Config (F4 one-target rule).
		pterm.Info.Printf("Install target: kube-context %q\n", req.KubeContext)
	}

	// Step 2.5 (deferred mode): no HelmManager yet — the caller had no
	// rest.Config upfront (standalone install), so resolve the selected
	// cluster's config now, through the injected ClusterAccess interface
	// (req 18/19). This used to live in a ~120-line copy of this whole
	// workflow (ExecuteWithContextDeferred) that drifted from this one; the
	// nil-check replaces the fork (audit B7).
	if w.chartService.helmManager == nil {
		kubeConfig := req.KubeConfig
		if kubeConfig == nil {
			resolved, kerr := w.clusterService.GetRestConfig(clusterName)
			if kerr != nil {
				return fmt.Errorf("failed to get rest.Config for cluster %s: %w", clusterName, kerr)
			}
			kubeConfig = resolved
		}
		if ierr := w.chartService.initializeHelmManager(kubeConfig, req.Verbose); ierr != nil {
			return fmt.Errorf("failed to initialize HelmManager: %w", ierr)
		}
	}

	// Step 3: Confirm installation (skipped in non-interactive and dry-run modes)
	if !req.NonInteractive && !req.DryRun {
		target := clusterName
		if target == "" {
			target = req.KubeContext
		}
		if !w.confirmInstallationOnCluster(target) {
			pterm.Info.Println("Installation cancelled.")
			return fmt.Errorf("installation cancelled by user")
		}
	}

	// Step 4: Regenerate certificates (skipped in non-interactive and dry-run modes)
	if !req.NonInteractive && !req.DryRun {
		// Non-fatal: failures are logged inside the method, continue regardless.
		_ = w.regenerateCertificates()
	} else if req.DryRun {
		pterm.Info.Println("Skipping certificate regeneration (dry-run)")
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

	// A dry run that ends without an explicit statement is indistinguishable
	// from a real run (verification report, minor observation).
	if req.DryRun {
		pterm.Success.Println("Dry run complete — nothing was changed.")
	}

	return nil
}

// selectCluster handles cluster selection
func (w *InstallationWorkflow) selectCluster(args []string, nonInteractive, verbose bool) (string, error) {
	clusterSelector := NewClusterSelector(w.clusterService, w.chartService.operationsUI)
	return clusterSelector.SelectCluster(args, nonInteractive, verbose)
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

// loadExistingConfiguration loads existing openframe-helm-values.yaml for non-interactive mode
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
		BaseHelmValuesPath: config.DefaultHelmValuesFile,
		TempHelmValuesPath: tempFilePath,
		ExistingValues:     baseValues,
		ModifiedSections:   make([]string, 0),
	}, nil
}

func (w *InstallationWorkflow) loadExistingConfiguration(requireValuesFile bool) (*types.ChartConfiguration, error) {
	modifier := templates.NewHelmValuesModifier()

	if _, err := os.Stat(config.DefaultHelmValuesFile); err != nil {
		// Upgrades REQUIRE the values file: proceeding with an empty map makes
		// `helm upgrade` replace the release values with chart defaults —
		// silently wiping registry credentials and ingress settings when run
		// from the wrong directory (audit F3/T1-2).
		if requireValuesFile {
			return nil, fmt.Errorf(
				"%s not found in the current directory — upgrading with no values file would deploy chart DEFAULTS and wipe the existing configuration; run from the directory containing the values file: %w",
				config.DefaultHelmValuesFile, err)
		}
		// Fresh non-interactive install/bootstrap: chart defaults are a valid
		// starting point (a clean machine has no values file yet), but say so
		// loudly instead of silently pretending a file was used.
		pterm.Warning.Printf("%s not found in the current directory — deploying chart defaults\n", config.DefaultHelmValuesFile)
	} else {
		pterm.Info.Printf("Using existing %s\n", config.DefaultHelmValuesFile)
	}

	// Load existing openframe-helm-values.yaml (empty map when absent — allowed
	// only on the fresh-install path above)
	values, err := modifier.LoadOrCreateBaseValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", config.DefaultHelmValuesFile, err)
	}

	// Create temporary file from the existing values (same as interactive mode)
	tempFilePath, err := modifier.CreateTemporaryValuesFile(values)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary values file: %w", err)
	}

	result := &types.ChartConfiguration{
		BaseHelmValuesPath: config.DefaultHelmValuesFile,
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

	// When the operator explicitly pins a ref (--ref), write it into
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

	cfg, err := configBuilder.BuildInstallConfigWithCustomHelmPath(
		req.Force, req.DryRun, req.Verbose, req.NonInteractive, clusterName,
		githubRepo, req.GitHubBranch, req.CertDir,
		chartConfig.TempHelmValuesPath,
	)
	if err != nil {
		return cfg, err
	}
	// One target per install: an explicit kube-context resolved at the command
	// layer overrides the ClusterName-derived context in every helm call.
	cfg.KubeContext = req.KubeContext
	cfg.SyncStragglersOnStall = req.SyncStragglersOnStall
	return cfg, nil
}

// performInstallation executes the actual installation
func (w *InstallationWorkflow) performInstallation(ctx context.Context, config config.ChartInstallConfig) error {
	// Create installer directly without factory. The ArgoCD wait manager gets
	// the SAME rest.Config the HelmManager was built with (falling back to the
	// selected cluster's context) — never the kubeconfig's current context,
	// which may point at an entirely different cluster (audit F4).
	pathResolver := w.chartService.configService.GetPathResolver()
	argoCDService, err := NewArgoCDForTarget(w.chartService.helmManager, pathResolver, w.chartService.executor, w.chartService.kubeConfig, config.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD service for the install target: %w", err)
	}
	appOfAppsService := NewAppOfApps(w.chartService.helmManager, w.chartService.gitRepository, pathResolver)

	installer := &Installer{
		argoCDService:    argoCDService,
		appOfAppsService: appOfAppsService,
	}

	err = installer.InstallChartsWithContext(ctx, config)
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

	// Check prerequisites first. Treat a non-TTY environment as non-interactive
	// even without the flag, so CI never blocks on a Y/N prompt (this is the only
	// prerequisite gate now — the app command group no longer runs a second one).
	installer := prerequisites.NewInstaller()
	if err := installer.CheckAndInstallNonInteractive(req.NonInteractive || sharedUI.IsNonInteractive()); err != nil {
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
