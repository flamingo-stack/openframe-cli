package helm

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	uispinner "github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HelmManager handles Helm operations
type HelmManager struct {
	executor      executor.CommandExecutor
	kubeConfig    *rest.Config         // Stores the cluster connection config
	dynamicClient dynamic.Interface    // Dynamic client for programmatic resource management
	kubeClient    kubernetes.Interface // Typed client for Deployment checks
	verbose       bool                 // Enable verbose logging
}

// NewHelmManager creates a new Helm manager with the given rest.Config
// The config is used to create the Kubernetes client for native API operations
func NewHelmManager(exec executor.CommandExecutor, config *rest.Config, verbose bool) (*HelmManager, error) {
	if config == nil {
		// Return a minimal HelmManager that can still execute helm commands
		// but will use kubectl fallback for deployment verification
		if verbose {
			pterm.Warning.Println("Creating HelmManager without rest.Config - native Go client will be unavailable")
		}
		return &HelmManager{
			executor: exec,
			verbose:  verbose,
		}, nil
	}

	// CRITICAL FIX: Bypass TLS Verification for local k3d clusters
	// Uses Insecure=true with CA data cleared, preserving client cert authentication.
	// Applied here as defense-in-depth in case the caller's config doesn't have it set.
	config = sharedconfig.ApplyInsecureTLSConfig(config)

	if verbose {
		pterm.Debug.Println("TLS verification bypassed for local k3d cluster (Insecure=true, auth preserved)")
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		// Log the error but continue with kubectl fallback capability
		if verbose {
			pterm.Warning.Printf("Failed to create Kubernetes core client (will use kubectl fallback): %v\n", err)
		}
		return &HelmManager{
			executor:   exec,
			kubeConfig: config,
			verbose:    verbose,
		}, nil
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to create Kubernetes dynamic client: %v\n", err)
		}
		// Still return with coreClient available
		return &HelmManager{
			executor:   exec,
			kubeConfig: config,
			kubeClient: coreClient,
			verbose:    verbose,
		}, nil
	}

	if verbose {
		pterm.Debug.Println("HelmManager initialized with native Go Kubernetes clients")
	}

	return &HelmManager{
		executor:      exec,
		kubeConfig:    config,
		dynamicClient: dynamicClient,
		kubeClient:    coreClient,
		verbose:       verbose,
	}, nil
}

// getHelmEnv returns environment variables for Helm to use writable directories
// This is especially important in CI environments where home directory may not have write permissions
func (h *HelmManager) getHelmEnv() map[string]string {
	// Define the directories - these are WSL/Linux paths
	// On Windows, helm runs inside WSL via the helm-wrapper.sh script which sets these
	helmDirs := map[string]string{
		"HELM_CACHE_HOME":  "/tmp/helm/cache",
		"HELM_CONFIG_HOME": "/tmp/helm/config",
		"HELM_DATA_HOME":   "/tmp/helm/data",
	}

	// Only create directories on non-Windows platforms
	// On Windows, the directories are created inside WSL by the wrapper script
	if runtime.GOOS != "windows" {
		for _, dir := range helmDirs {
			if err := os.MkdirAll(dir, 0750); err != nil {
				pterm.Debug.Printf("failed to pre-create helm dir %s: %v\n", dir, err)
			}
		}
	}

	return helmDirs
}

// IsHelmInstalled checks if Helm is available
func (h *HelmManager) IsHelmInstalled(ctx context.Context) error {
	_, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    []string{"version", "--short"},
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return errors.ErrHelmNotAvailable
	}
	return nil
}

// IsChartInstalled checks if a chart is already installed
func (h *HelmManager) IsChartInstalled(ctx context.Context, releaseName, namespace string) (bool, error) {
	args := []string{"list", "-q", "-n", namespace}
	if releaseName != "" {
		args = append(args, "-f", releaseName)
	}

	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return false, err
	}

	releases := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, release := range releases {
		if strings.TrimSpace(release) == releaseName {
			return true, nil
		}
	}

	return false, nil
}

// UninstallRelease removes a Helm release from a namespace. Missing releases are
// treated as success (--ignore-not-found). kubeContext, when non-empty, targets
// a specific kube-context (matching how installs pin the context).
func (h *HelmManager) UninstallRelease(ctx context.Context, releaseName, namespace, kubeContext string) error {
	args := []string{"uninstall", releaseName, "-n", namespace, "--ignore-not-found", "--wait"}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	_, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("helm uninstall %s: %w", releaseName, err)
	}
	return nil
}

// argoCDInstallArgs builds the `helm upgrade --install argo-cd` argument list.
// Pure and testable — the CRDs are installed by the chart itself
// (crds.install=true), so no crds flag is passed.
func argoCDInstallArgs(cfg config.ChartInstallConfig, valuesFilePath string) []string {
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--version=10.1.0",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "7m",
		"-f", valuesFilePath,
	}
	if cfg.ClusterName != "" {
		args = append(args, "--kube-context", fmt.Sprintf("k3d-%s", cfg.ClusterName))
	}
	if cfg.DryRun {
		args = append(args, "--dry-run")
	}
	return args
}

// installArgoCDHelm runs `helm upgrade --install argo-cd ... -f -`, feeding the
// embedded ArgoCD values via stdin so nothing is written to the user's
// filesystem (and there is no path to convert for WSL). Split out from
// InstallArgoCDWithProgress so the stdin / no-temp-file contract is unit-testable
// without the post-install verification and deployment waits.
func (h *HelmManager) installArgoCDHelm(ctx context.Context, cfg config.ChartInstallConfig) (*executor.CommandResult, error) {
	args := argoCDInstallArgs(cfg, "-")
	if cfg.Verbose {
		pterm.Debug.Printf("Executing: helm %s\n", strings.Join(args, " "))
	}
	return h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
		Stdin:   []byte(argocd.GetArgoCDValues()),
	})
}

// InstallArgoCDWithProgress installs ArgoCD using Helm with progress indicators
func (h *HelmManager) InstallArgoCDWithProgress(ctx context.Context, config config.ChartInstallConfig) error {
	// Show progress for each step only if not in silent/non-interactive mode
	var spinner *uispinner.Spinner
	if !config.Silent && !config.NonInteractive {
		spinner = uispinner.Start("Installing ArgoCD...")
	} else {
		pterm.Info.Println("Installing ArgoCD...")
	}

	// Add ArgoCD repository silently
	_, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    []string{"repo", "add", "argo", "https://argoproj.github.io/argo-helm"},
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		// Ignore if already exists
		if !strings.Contains(err.Error(), "already exists") {
			if spinner != nil {
				spinner.Stop()
			}
			return fmt.Errorf("failed to add ArgoCD repository: %w", err)
		}
	}

	// Update repositories silently
	_, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    []string{"repo", "update"},
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to update Helm repositories: %w", err)
	}

	// First, verify kubectl can connect to the cluster with retries
	// Use explicit context if cluster name is provided (important for Windows/WSL)
	maxRetries := 10
	retryDelay := 3 // seconds
	var lastErr error

	// Build kubectl args with explicit context if cluster name is provided
	kubectlArgs := []string{"cluster-info"}
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		kubectlArgs = []string{"--context", contextName, "cluster-info"}
	}

	for i := 0; i < maxRetries; i++ {
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    kubectlArgs,
		})

		if err == nil && result.ExitCode == 0 {
			// Cluster is accessible
			break
		}

		lastErr = err
		if i < maxRetries-1 {
			if config.Verbose {
				pterm.Info.Printf("Waiting for cluster to be ready... (attempt %d/%d)\n", i+1, maxRetries)
			}
			// Check if context was cancelled
			select {
			case <-ctx.Done():
				if spinner != nil {
					spinner.Stop()
				}
				return ctx.Err()
			case <-time.After(time.Duration(retryDelay) * time.Second):
				// Continue to next retry
			}
		}
	}

	if lastErr != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to connect to cluster after %d retries: %w", maxRetries, lastErr)
	}

	// ArgoCD CRDs are installed by the Helm chart itself (crds.install=true, the
	// chart default), so they always match the chart's ArgoCD version. No separate
	// CRD fetch/apply is needed.

	// Installation details are now silent - just show in verbose mode
	if config.Verbose {
		pterm.Info.Printf("   Version: 10.1.0\n")
		pterm.Info.Printf("   Namespace: argocd\n")
		pterm.Info.Println("   Values: piped via stdin (-f -)")
	}

	// Explicitly create and verify the argocd namespace exists BEFORE Helm install
	// This addresses the race condition where Helm's --create-namespace may not complete properly
	// in Windows/WSL environments, leading to "namespace not found" errors during deployment verification
	if !config.DryRun {
		if err := h.ensureArgoCDNamespace(ctx, config.ClusterName, config.Verbose); err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			return fmt.Errorf("failed to ensure argocd namespace exists: %w", err)
		}
	}

	if config.DryRun && config.Verbose {
		pterm.Info.Println("Running in dry-run mode...")
	}

	result, err := h.installArgoCDHelm(ctx, config)
	if err != nil {
		// Check if the error is due to context cancellation (CTRL-C)
		if ctx.Err() == context.Canceled {
			if spinner != nil {
				spinner.Stop()
			}
			return ctx.Err() // Return context cancellation directly without extra messaging
		}

		if spinner != nil {
			spinner.Stop()
		}

		// Show diagnostic information about ArgoCD pods
		h.showArgoCDDiagnostics(ctx, config.ClusterName)

		// Include stdout and stderr output for better debugging
		// On Windows/WSL, stderr is redirected to stdout via 2>&1, so check both
		if result != nil {
			output := result.Stderr
			if output == "" {
				output = result.Stdout
			}
			if output != "" {
				return fmt.Errorf("failed to install ArgoCD: %w\nHelm output: %s", err, output)
			}
		}
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Log Helm output for debugging (helps identify if Helm actually created resources)
	if config.Verbose && result != nil {
		if result.Stdout != "" {
			pterm.Info.Println("Helm stdout:")
			pterm.Println(result.Stdout)
		}
		if result.Stderr != "" {
			pterm.Info.Println("Helm stderr:")
			pterm.Println(result.Stderr)
		}
	}

	// Verify the Helm release was actually created by checking helm list
	if err := h.verifyHelmRelease(ctx, "argo-cd", "argocd", config.ClusterName, config.Verbose); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("ArgoCD Helm install completed but release verification failed: %w", err)
	}

	// Wait for ArgoCD deployments to be created after Helm install
	// This addresses the race condition where Helm --wait returns before Kubernetes
	// has actually created the Deployment objects (common in k3d/CI environments)
	//
	// Use native Go client for all platforms (including Windows) for fast, reliable polling
	// The kubeClient uses the same kubeconfig that was used to create the cluster
	// On Windows/WSL2, always use kubectl because the native Go client can't reliably
	// reach the cluster running inside WSL due to networking bridge issues
	if h.kubeClient != nil && runtime.GOOS != "windows" {
		if err := h.waitForArgoCDDeployments(ctx, config.Verbose); err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			// Check if the error is due to context cancellation (CTRL-C)
			if ctx.Err() == context.Canceled {
				return ctx.Err()
			}
			pterm.Warning.Println("Helm install reported success but ArgoCD deployments were not found")
			pterm.Info.Println("This may indicate a Helm caching issue or cluster connectivity problem")
			return fmt.Errorf("ArgoCD Helm install completed but deployments were not created: %w", err)
		}
	} else {
		// Fallback to kubectl-based verification when native Go client is unavailable or on Windows
		if config.Verbose {
			if runtime.GOOS == "windows" {
				pterm.Info.Println("Using kubectl for deployment verification (Windows/WSL2 mode)")
			} else {
				pterm.Warning.Println("Native Go client unavailable, using kubectl for deployment verification")
			}
		}
		if err := h.waitForArgoCDDeploymentsKubectl(ctx, config.ClusterName, config.Verbose); err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			if ctx.Err() == context.Canceled {
				return ctx.Err()
			}
			pterm.Warning.Println("Helm install reported success but ArgoCD deployments were not found")
			pterm.Info.Println("This may indicate a Helm caching issue or cluster connectivity problem")
			return fmt.Errorf("ArgoCD Helm install completed but deployments were not created: %w", err)
		}
	}

	if spinner != nil {
		spinner.Stop()
	}

	return nil
}

// InstallAppOfAppsFromLocal installs the app-of-apps chart from a local path
func (h *HelmManager) InstallAppOfAppsFromLocal(ctx context.Context, config config.ChartInstallConfig, certFile, keyFile string) error {
	// Validate configuration
	if config.AppOfApps == nil {
		return fmt.Errorf("app-of-apps configuration is required")
	}

	appConfig := config.AppOfApps
	if appConfig.ChartPath == "" {
		return fmt.Errorf("chart path is required for app-of-apps installation")
	}

	// On Windows, validate WSL Ubuntu is accessible before proceeding
	// This provides early, clear error messages instead of cryptic failures later
	if runtime.GOOS == "windows" {
		if !executor.IsWSLAvailable() {
			return fmt.Errorf("WSL is not available on this system. Helm requires WSL2 with Ubuntu to run on Windows.\n" +
				"Please install WSL2: wsl --install")
		}
		if !executor.IsWSLUbuntuAvailable() {
			return fmt.Errorf("WSL Ubuntu distribution is not accessible.\n" +
				"This could mean:\n" +
				"  1. Ubuntu is not installed (run: wsl --install -d Ubuntu)\n" +
				"  2. Ubuntu is not running (run: wsl -d Ubuntu)\n" +
				"  3. Ubuntu is still initializing (wait a few seconds and retry)\n" +
				"Check status with: wsl --list --verbose")
		}
		if h.verbose {
			pterm.Debug.Println("WSL Ubuntu is accessible, proceeding with helm installation")
		}
	}

	// Verify cluster connectivity before running helm (important after idle periods)
	// This helps diagnose issues where WSL networking may have gone stale
	if err := h.verifyClusterConnectivity(ctx, config); err != nil {
		return fmt.Errorf("cluster connectivity check failed before app-of-apps installation: %w", err)
	}

	// Convert Windows paths to WSL paths if needed (for Helm running in WSL2)
	chartPath := appConfig.ChartPath
	valuesFilePath := appConfig.ValuesFile
	certFilePath := certFile
	keyFilePath := keyFile

	if runtime.GOOS == "windows" {
		var err error

		// Convert chart path
		if chartPath != "" {
			chartPath, err = h.convertWindowsPathToWSL(appConfig.ChartPath)
			if err != nil {
				return fmt.Errorf("failed to convert chart path for WSL: %w", err)
			}
		}

		// Convert values file path
		if valuesFilePath != "" {
			valuesFilePath, err = h.convertWindowsPathToWSL(appConfig.ValuesFile)
			if err != nil {
				return fmt.Errorf("failed to convert values file path for WSL: %w", err)
			}
		}

		// Convert certificate file paths
		if certFile != "" {
			certFilePath, err = h.convertWindowsPathToWSL(certFile)
			if err != nil {
				return fmt.Errorf("failed to convert cert file path for WSL: %w", err)
			}
		}

		if keyFile != "" {
			keyFilePath, err = h.convertWindowsPathToWSL(keyFile)
			if err != nil {
				return fmt.Errorf("failed to convert key file path for WSL: %w", err)
			}
		}
	}

	// Install app-of-apps using the local chart path
	args := []string{
		"upgrade", "--install", "app-of-apps", chartPath,
		"--namespace", appConfig.Namespace,
		"--wait",
		"--timeout", appConfig.Timeout,
		"-f", valuesFilePath,
	}

	// Only add certificate files if they exist and are not empty paths
	if certFile != "" && keyFile != "" {
		// Check if files actually exist before adding them (use original Windows paths for os.Stat)
		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				args = append(args,
					// OSS mode certificates (use WSL paths for Helm)
					"--set-file", fmt.Sprintf("deployment.oss.ingress.localhost.tls.cert=%s", certFilePath),
					"--set-file", fmt.Sprintf("deployment.oss.ingress.localhost.tls.key=%s", keyFilePath),
					// SaaS mode certificates (use WSL paths for Helm)
					"--set-file", fmt.Sprintf("deployment.saas.ingress.localhost.tls.cert=%s", certFilePath),
					"--set-file", fmt.Sprintf("deployment.saas.ingress.localhost.tls.key=%s", keyFilePath),
				)
			}
		}
	}

	// Add explicit kube-context if cluster name is provided (important for Windows/WSL)
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		args = append(args, "--kube-context", contextName)
	}

	if config.DryRun {
		args = append(args, "--dry-run")
	}

	// Execute helm command with local chart path
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})

	if err != nil {
		// Check if the error is due to context cancellation (CTRL-C)
		if ctx.Err() == context.Canceled {
			return ctx.Err() // Return context cancellation directly without extra messaging
		}

		// Include stderr output for better debugging
		if result != nil && result.Stderr != "" {
			return fmt.Errorf("failed to install app-of-apps: %w\nHelm output: %s", err, result.Stderr)
		}
		return fmt.Errorf("failed to install app-of-apps: %w", err)
	}

	return nil
}

// GetChartStatus returns the status of a chart
func (h *HelmManager) GetChartStatus(ctx context.Context, releaseName, namespace string) (models.ChartInfo, error) {
	args := []string{"status", releaseName, "-n", namespace, "--output", "json"}

	_, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return models.ChartInfo{}, fmt.Errorf("failed to get chart status: %w", err)
	}

	// Parse JSON output and return chart info
	// For now, return basic info
	return models.ChartInfo{
		Name:      releaseName,
		Namespace: namespace,
		Status:    "deployed", // Parse from JSON
		Version:   "1.0.0",    // Parse from JSON
	}, nil
}

// convertWindowsPathToWSL converts a Windows path to a WSL path format
// Example: C:\Users\foo\file.txt -> /mnt/c/Users/foo/file.txt
// This is necessary when passing file paths from Windows to commands running in WSL2
//
// IMPORTANT: Uses `wsl wslpath` command for reliable conversion that handles:
// - Windows 8.3 short filenames (e.g., RUNNER~1 -> runneradmin)
// - Proper path escaping and special characters
// Falls back to manual conversion if wslpath is not available.
func (h *HelmManager) convertWindowsPathToWSL(windowsPath string) (string, error) {
	if windowsPath == "" {
		return "", fmt.Errorf("empty path provided")
	}

	// First, convert relative paths to absolute paths
	// This is necessary because wslpath requires absolute paths and
	// WSL won't be able to find files using Windows relative paths
	absPath, err := filepath.Abs(windowsPath)
	if err != nil {
		// If we can't get absolute path, try with original
		absPath = windowsPath
	}

	// Expand Windows 8.3 short filenames to long path names
	// For example: C:\Users\RUNNER~1\... -> C:\Users\runneradmin\...
	// This is critical because WSL doesn't understand Windows short filenames
	expandedPath, err := expandShortPath(absPath)
	if err == nil && expandedPath != "" {
		absPath = expandedPath
		if h.verbose {
			pterm.Debug.Printf("Expanded short path: %s -> %s\n", windowsPath, absPath)
		}
	}

	// Try using WSL's wslpath command for reliable conversion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert backslashes to forward slashes before passing to wslpath
	// This prevents backslashes from being interpreted as escape characters
	// when WSL passes arguments to the Linux command
	// wslpath accepts both forward and backward slashes
	windowsPathForWSL := strings.ReplaceAll(absPath, "\\", "/")

	// Must specify -d Ubuntu to use the correct distribution
	// Without this, wsl may try to use a non-existent default distribution
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "wslpath", "-u", windowsPathForWSL},
	})

	if err == nil && result != nil && result.ExitCode == 0 {
		wslPath := strings.TrimSpace(result.Stdout)
		if wslPath != "" {
			if h.verbose {
				pterm.Debug.Printf("Converted path via wslpath: %s -> %s\n", windowsPath, wslPath)
			}
			return wslPath, nil
		}
	}

	// Log WSL errors for debugging
	if err != nil {
		// Check if this is a WSL-specific error
		var wslErr *executor.WSLError
		if stderrors.As(err, &wslErr) {
			if h.verbose {
				pterm.Warning.Printf("WSL error during path conversion: %s\n", wslErr.Error())
				pterm.Info.Printf("Falling back to manual path conversion\n")
			}
		} else if h.verbose {
			pterm.Debug.Printf("wslpath command failed: %v\n", err)
		}
	} else if result != nil && result.ExitCode != 0 {
		if h.verbose {
			pterm.Debug.Printf("wslpath returned exit code %d, stderr: %s\n", result.ExitCode, result.Stderr)
		}
	}

	// Fallback to manual conversion if wslpath is not available
	if h.verbose {
		pterm.Debug.Printf("Using manual path conversion for: %s\n", absPath)
	}

	// Replace backslashes with forward slashes (use absPath which is already absolute)
	path := strings.ReplaceAll(absPath, "\\", "/")

	// Convert drive letter (e.g., C: -> /mnt/c)
	if len(path) >= 2 && path[1] == ':' {
		driveLetter := strings.ToLower(string(path[0]))
		// Remove the drive letter and colon, then prepend /mnt/<drive>
		path = "/mnt/" + driveLetter + path[2:]
	}

	return path, nil
}
