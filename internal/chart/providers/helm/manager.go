package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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
	"github.com/pterm/pterm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

// HelmManager handles Helm operations
type HelmManager struct {
	executor      executor.CommandExecutor
	kubeConfig    *rest.Config                      // Stores the cluster connection config
	dynamicClient dynamic.Interface                 // Dynamic client for programmatic resource management
	kubeClient    kubernetes.Interface              // Typed client for Deployment checks
	crdClient     apiextensionsclient.Interface     // CRD client for checking CRD existence
	verbose       bool                              // Enable verbose logging
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

	// Create CRD client for checking CRD existence
	crdClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to create CRD client: %v\n", err)
		}
		// Still return with other clients available
		return &HelmManager{
			executor:      exec,
			kubeConfig:    config,
			dynamicClient: dynamicClient,
			kubeClient:    coreClient,
			verbose:       verbose,
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
		crdClient:     crdClient,
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
			os.MkdirAll(dir, 0755)
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

// InstallArgoCD installs ArgoCD using Helm with exact commands specified
func (h *HelmManager) InstallArgoCD(ctx context.Context, config config.ChartInstallConfig) error {
	// Add ArgoCD Helm repository
	_, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    []string{"repo", "add", "argo", "https://argoproj.github.io/argo-helm"},
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("failed to add ArgoCD repository: %w", err)
	}

	// Update repositories
	_, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    []string{"repo", "update"},
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("failed to update Helm repositories: %w", err)
	}

	// Create a temporary file with ArgoCD values
	tmpFile, err := os.CreateTemp("", "argocd-values-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary values file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the ArgoCD values to the temporary file
	if _, err := tmpFile.WriteString(argocd.GetArgoCDValues(config.ArgoCDConfig)); err != nil {
		return fmt.Errorf("failed to write values to temporary file: %w", err)
	}
	tmpFile.Close()

	// Convert Windows path to WSL path if needed (for Helm running in WSL2)
	valuesFilePath := tmpFile.Name()
	if runtime.GOOS == "windows" {
		valuesFilePath, err = h.convertWindowsPathToWSL(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("failed to convert values file path for WSL: %w", err)
		}
	}

	// Install ArgoCD with upgrade --install
	// CRDs are handled separately via native Go client, so we tell Helm to skip them
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--version=9.3.4",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "7m",
		"-f", valuesFilePath,
		"--set", "crds.install=false",
	}

	// Add explicit kube-context if cluster name is provided (important for Windows/WSL)
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		args = append(args, "--kube-context", contextName)
	}

	if config.DryRun {
		args = append(args, "--dry-run")
	}

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

	return nil
}

// InstallArgoCDWithProgress installs ArgoCD using Helm with progress indicators
func (h *HelmManager) InstallArgoCDWithProgress(ctx context.Context, config config.ChartInstallConfig) error {
	// Show progress for each step only if not in silent/non-interactive mode
	var spinner *pterm.SpinnerPrinter
	if !config.Silent && !config.NonInteractive {
		spinner, _ = pterm.DefaultSpinner.Start("Installing ArgoCD...")
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

	// Install ArgoCD CRDs unless skipped
	// CRITICAL: CRDs must be installed and verified BEFORE Helm upgrade runs
	// This eliminates the race condition where Helm tries to create CRD-based resources
	// before the CRD definitions are available to the API server
	if !config.SkipCRDs {
		if config.Verbose {
			pterm.Info.Println("Installing ArgoCD CRDs using native Go client...")
		}

		// Verify clients are initialized (should be set in constructor)
		if h.dynamicClient == nil {
			if spinner != nil {
				spinner.Stop()
			}
			return fmt.Errorf("dynamic client not initialized; ensure HelmManager was created with valid rest.Config")
		}

		// Install CRDs programmatically using client-go dynamic client
		crdUrls := []string{
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.2/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.2/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.13.2/manifests/crds/appproject-crd.yaml",
		}

		for _, crdUrl := range crdUrls {
			if err := h.applyManifestFromURL(ctx, crdUrl); err != nil {
				if spinner != nil {
					spinner.Stop()
				}
				return fmt.Errorf("failed to install ArgoCD CRDs: %w", err)
			}
		}

		if config.Verbose {
			pterm.Success.Println("ArgoCD CRDs applied successfully via API")
		}

		// Wait for CRDs to be available BEFORE running Helm
		// This ensures the Kubernetes API server recognizes the CRD types
		if h.crdClient != nil {
			if config.Verbose {
				pterm.Info.Println("Waiting for ArgoCD CRDs to be available...")
			}
			if err := h.waitForArgoCDCRD(ctx, config.Verbose); err != nil {
				if spinner != nil {
					spinner.Stop()
				}
				return fmt.Errorf("failed waiting for ArgoCD CRDs to become available: %w", err)
			}
		}
	} else if config.Verbose {
		pterm.Info.Println("Skipping ArgoCD CRDs installation (--skip-crds)")
	}

	// Create a temporary file with ArgoCD values
	tmpFile, err := os.CreateTemp("", "argocd-values-*.yaml")
	if err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to create temporary values file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the ArgoCD values to the temporary file
	if _, err := tmpFile.WriteString(argocd.GetArgoCDValues(config.ArgoCDConfig)); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to write values to temporary file: %w", err)
	}
	tmpFile.Close()

	// Convert Windows path to WSL path if needed (for Helm running in WSL2)
	valuesFilePath := tmpFile.Name()
	if runtime.GOOS == "windows" {
		valuesFilePath, err = h.convertWindowsPathToWSL(tmpFile.Name())
		if err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			return fmt.Errorf("failed to convert values file path for WSL: %w", err)
		}
	}

	// Installation details are now silent - just show in verbose mode
	if config.Verbose {
		pterm.Info.Printf("   Version: 9.3.4\n")
		pterm.Info.Printf("   Namespace: argocd\n")
		pterm.Info.Printf("   Values file (Windows): %s\n", tmpFile.Name())
		if runtime.GOOS == "windows" {
			pterm.Info.Printf("   Values file (WSL): %s\n", valuesFilePath)
		}
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

	// Install ArgoCD with upgrade --install
	// CRDs are handled separately via native Go client, so we tell Helm to skip them
	// This prevents the race condition where Helm tries to install CRDs that we already installed
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--version=9.3.4",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "7m",
		"-f", valuesFilePath,
		"--set", "crds.install=false",
	}

	// Add explicit kube-context if cluster name is provided (important for Windows/WSL)
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		args = append(args, "--kube-context", contextName)
	}

	if config.DryRun {
		args = append(args, "--dry-run")
		if config.Verbose {
			pterm.Info.Println("Running in dry-run mode...")
		}
	}

	// Show command being executed
	if config.Verbose {
		pterm.Debug.Printf("Executing: helm %s\n", strings.Join(args, " "))
	}

	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
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
		if wslErr, ok := err.(*executor.WSLError); ok {
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

// applyManifestFromURL fetches a multi-document YAML manifest and applies its resources
// using the dynamic client. This is used for CRD installation without relying on kubectl.
func (h *HelmManager) applyManifestFromURL(ctx context.Context, url string) error {
	// 1. Fetch the YAML manifest content
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch manifest: received status code %d from %s", resp.StatusCode, url)
	}

	manifestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read manifest body: %w", err)
	}

	// 2. Split the manifest into individual documents (resources)
	resources := strings.Split(string(manifestBytes), "---")

	if h.dynamicClient == nil {
		return fmt.Errorf("dynamic client not initialized; cannot apply manifest")
	}

	for _, resourceYAML := range resources {
		if strings.TrimSpace(resourceYAML) == "" {
			continue // Skip empty documents
		}

		// 3. Unmarshal YAML into an unstructured object
		var unstructuredObj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(resourceYAML), &unstructuredObj); err != nil {
			return fmt.Errorf("failed to unmarshal YAML resource: %w", err)
		}

		// Skip if resource is empty after unmarshalling (e.g., just comments)
		if unstructuredObj.Object == nil {
			continue
		}

		// 4. Determine GroupVersionResource (GVR) for the dynamic client
		gvk := unstructuredObj.GroupVersionKind()
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: strings.ToLower(gvk.Kind) + "s", // Heuristic: pluralize Kind
		}

		// 5. Apply the resource (try Create first, then handle conflict with Update)
		namespace := unstructuredObj.GetNamespace()

		var resourceInterface dynamic.ResourceInterface
		if namespace == "" {
			// For cluster-scoped resources (like CRDs)
			resourceInterface = h.dynamicClient.Resource(gvr)
		} else {
			// For namespaced resources
			resourceInterface = h.dynamicClient.Resource(gvr).Namespace(namespace)
		}

		// Attempt to create the resource
		_, err = resourceInterface.Create(ctx, &unstructuredObj, metav1.CreateOptions{})

		// If creation fails due to conflict (already exists), attempt to update (replace)
		if err != nil && strings.Contains(err.Error(), "already exists") {
			// Get the existing resource to obtain its resourceVersion
			existing, getErr := resourceInterface.Get(ctx, unstructuredObj.GetName(), metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("failed to get existing resource %s/%s: %w", gvk.Kind, unstructuredObj.GetName(), getErr)
			}
			unstructuredObj.SetResourceVersion(existing.GetResourceVersion())
			_, err = resourceInterface.Update(ctx, &unstructuredObj, metav1.UpdateOptions{})
		}

		if err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w", gvk.Kind, unstructuredObj.GetName(), err)
		}

		if h.verbose {
			pterm.Debug.Printf("Applied resource: %s/%s\n", gvk.Kind, unstructuredObj.GetName())
		}
	}

	return nil
}

// waitForArgoCDDeployments waits for ArgoCD workloads to be created in the cluster
// This addresses the race condition where Helm's --wait returns before Kubernetes
// has actually created the Deployment/StatefulSet objects (common in k3d/CI environments)
//
// NOTE: CRDs are now installed and verified BEFORE Helm runs (see InstallArgoCDWithProgress),
// so this function focuses only on verifying the workloads exist.
//
// ArgoCD v3.x (Helm chart 8.x) deploys the application-controller as a StatefulSet,
// while server and repo-server remain as Deployments.
func (h *HelmManager) waitForArgoCDDeployments(ctx context.Context, verbose bool) error {
	if h.kubeClient == nil {
		return fmt.Errorf("Kubernetes core client not initialized")
	}

	// Wait for API port to be available before making API calls
	// This prevents flooding a dead port with requests on Windows/WSL2
	if err := h.waitForAPIPort(ctx, 45*time.Second); err != nil {
		return fmt.Errorf("API port never opened: %w", err)
	}

	// List of expected Deployments (server and repo-server)
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
	}

	// List of expected StatefulSets (application-controller in ArgoCD v3.x)
	expectedStatefulSets := []string{
		"argocd-application-controller",
	}

	// CRITICAL: Use extended timeout since cluster operations can be slow
	// Native API calls are fast (~ms), so we use frequent polling with longer total timeout
	timeout := 90 * time.Second      // 90 seconds for slow CI/Windows environments
	retryInterval := 1 * time.Second // Fast polling interval (native API is ~ms per call)

	pterm.Info.Println("Waiting for ArgoCD workloads via NATIVE API...")

	// Use wait.PollUntilContextTimeout for resilient polling
	return wait.PollUntilContextTimeout(ctx, retryInterval, timeout, false, func(ctx context.Context) (bool, error) {

		missingWorkloads := []string{}

		// Check Deployments
		for _, name := range expectedDeployments {
			_, err := h.kubeClient.AppsV1().Deployments("argocd").Get(ctx, name, metav1.GetOptions{})

			if k8serrors.IsNotFound(err) {
				missingWorkloads = append(missingWorkloads, "deployment/"+name)
			} else if err != nil {
				// If it's a transient API error (not 'Not Found'), log and retry
				pterm.Warning.Printf("Transient API error checking deployment %s: %v\n", name, err)
				return false, nil
			}
		}

		// Check StatefulSets (application-controller in ArgoCD v3.x)
		for _, name := range expectedStatefulSets {
			_, err := h.kubeClient.AppsV1().StatefulSets("argocd").Get(ctx, name, metav1.GetOptions{})

			if k8serrors.IsNotFound(err) {
				missingWorkloads = append(missingWorkloads, "statefulset/"+name)
			} else if err != nil {
				// If it's a transient API error (not 'Not Found'), log and retry
				pterm.Warning.Printf("Transient API error checking statefulset %s: %v\n", name, err)
				return false, nil
			}
		}

		if len(missingWorkloads) == 0 {
			pterm.Success.Println("All ArgoCD workloads found.")
			return true, nil // Success: All workloads exist.
		}

		if verbose {
			pterm.Debug.Printf("Still missing workloads: %v\n", missingWorkloads)
		}

		return false, nil // Keep polling
	})
}

// waitForArgoCDCRD waits for the ArgoCD Application CRD to be created
// This ensures the Helm chart has fully installed the CRDs before checking for deployments
func (h *HelmManager) waitForArgoCDCRD(ctx context.Context, verbose bool) error {
	if h.crdClient == nil {
		return nil // Skip if CRD client is not available
	}

	timeout := 30 * time.Second      // 30 seconds for CRD to appear
	retryInterval := 1 * time.Second // Check every second

	if verbose {
		pterm.Info.Println("Waiting for ArgoCD Application CRD to appear...")
	}

	return wait.PollUntilContextTimeout(ctx, retryInterval, timeout, false, func(ctx context.Context) (bool, error) {
		_, err := h.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "applications.argoproj.io", metav1.GetOptions{})
		if err == nil {
			if verbose {
				pterm.Success.Println("ArgoCD Application CRD found.")
			}
			return true, nil
		}
		if k8serrors.IsNotFound(err) {
			return false, nil // Keep polling
		}
		// Log transient errors but keep polling
		if verbose {
			pterm.Debug.Printf("Transient error checking CRD: %v\n", err)
		}
		return false, nil
	})
}

// ensureArgoCDNamespace creates the argocd namespace if it doesn't exist and waits for it to be active
// This addresses the race condition where Helm's --create-namespace may not complete before the command returns
// On Windows/WSL, uses kubectl since the native Go client can't reach the cluster running in WSL
func (h *HelmManager) ensureArgoCDNamespace(ctx context.Context, clusterName string, verbose bool) error {
	namespace := "argocd"

	// On Windows, use kubectl since the native Go client can't reach the cluster running in WSL
	if runtime.GOOS == "windows" || h.kubeClient == nil {
		return h.ensureArgoCDNamespaceKubectl(ctx, clusterName, verbose)
	}

	// Use native Go client for non-Windows platforms
	if verbose {
		pterm.Info.Println("Ensuring argocd namespace exists via native Go client...")
	}

	// Check if namespace already exists
	_, err := h.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		if verbose {
			pterm.Debug.Println("Namespace argocd already exists")
		}
		return nil
	}

	if !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to check namespace existence: %w", err)
	}

	// Create the namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = h.kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create argocd namespace: %w", err)
	}

	if verbose {
		pterm.Info.Println("Created argocd namespace, waiting for it to become Active...")
	}

	// Wait for namespace to become Active
	return wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		ns, err := h.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			return false, nil // Keep polling on transient errors
		}
		if ns.Status.Phase == corev1.NamespaceActive {
			if verbose {
				pterm.Success.Println("Namespace argocd is Active")
			}
			return true, nil
		}
		return false, nil
	})
}

// ensureArgoCDNamespaceKubectl creates the argocd namespace using kubectl (for Windows/WSL)
func (h *HelmManager) ensureArgoCDNamespaceKubectl(ctx context.Context, clusterName string, verbose bool) error {
	namespace := "argocd"

	if verbose {
		pterm.Info.Println("Ensuring argocd namespace exists via kubectl...")
	}

	// Build kubectl args with explicit context if cluster name is provided
	baseArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}

	// Check if namespace exists
	checkArgs := append(baseArgs, "get", "namespace", namespace, "-o", "name")
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    checkArgs,
	})

	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		if verbose {
			pterm.Debug.Println("Namespace argocd already exists")
		}
		return nil
	}

	// Create the namespace
	createArgs := append(baseArgs, "create", "namespace", namespace)
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    createArgs,
	})

	if err != nil {
		// Check if it's an "already exists" error (race condition)
		if result != nil && strings.Contains(result.Stderr, "already exists") {
			if verbose {
				pterm.Debug.Println("Namespace argocd already exists (created by concurrent process)")
			}
			return nil
		}
		return fmt.Errorf("failed to create argocd namespace: %w", err)
	}

	if verbose {
		pterm.Info.Println("Created argocd namespace, waiting for it to become Active...")
	}

	// Wait for namespace to become Active
	maxRetries := 60 // 60 * 500ms = 30 seconds
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		statusArgs := append(baseArgs, "get", "namespace", namespace, "-o", "jsonpath={.status.phase}")
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    statusArgs,
		})

		if err == nil && result != nil && strings.TrimSpace(result.Stdout) == "Active" {
			if verbose {
				pterm.Success.Println("Namespace argocd is Active")
			}
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for argocd namespace to become Active")
}

// waitForArgoCDDeploymentsKubectl waits for ArgoCD workloads using kubectl
// This is a fallback for when the native Go client is unavailable (e.g., Windows/WSL)
//
// ArgoCD v3.x (Helm chart 8.x) deploys the application-controller as a StatefulSet,
// while server and repo-server remain as Deployments.
func (h *HelmManager) waitForArgoCDDeploymentsKubectl(ctx context.Context, clusterName string, verbose bool) error {
	// List of expected Deployments (server and repo-server)
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
	}

	// List of expected StatefulSets (application-controller in ArgoCD v3.x)
	expectedStatefulSets := []string{
		"argocd-application-controller",
	}

	// Wait settings - increased for slow CI environments
	maxRetries := 40           // 40 retries * 3 seconds = 120 seconds max (2 minutes)
	retryInterval := 3 * time.Second
	initialDelay := 5 * time.Second // Give Kubernetes time to create resources after Helm completes

	pterm.Info.Println("Waiting for ArgoCD workloads to appear via kubectl...")

	// Initial delay: Helm's --wait returns before Kubernetes controllers fully create resources
	// This delay allows the controllers to process the Helm release and create resources
	if verbose {
		pterm.Debug.Printf("Initial delay of %v to allow Kubernetes controllers to create resources...\n", initialDelay)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(initialDelay):
	}

	// Build kubectl base args with explicit context if cluster name is provided
	contextArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		contextArgs = append(contextArgs, "--context", contextName)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var missingWorkloads []string

		// Check Deployments
		deployArgs := append(contextArgs, "-n", "argocd", "get", "deployments", "-o", "json")
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    deployArgs,
		})

		if err == nil && result != nil && result.Stdout != "" {
			var deploymentList appsv1.DeploymentList
			if jsonErr := json.Unmarshal([]byte(result.Stdout), &deploymentList); jsonErr != nil {
				lastErr = fmt.Errorf("failed to parse deployments JSON: %v", jsonErr)
				if verbose {
					pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): JSON parse error\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Build set of found deployments
			foundDeployments := make(map[string]bool)
			for _, d := range deploymentList.Items {
				foundDeployments[d.Name] = true
			}

			// Check if all expected deployments are present
			for _, expected := range expectedDeployments {
				if !foundDeployments[expected] {
					missingWorkloads = append(missingWorkloads, "deployment/"+expected)
				}
			}
		} else {
			lastErr = fmt.Errorf("kubectl deployments error: %v", err)
			if verbose {
				pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): kubectl deployments error\n", i+1, maxRetries)
			}
			time.Sleep(retryInterval)
			continue
		}

		// Check StatefulSets
		stsArgs := append(contextArgs, "-n", "argocd", "get", "statefulsets", "-o", "json")
		result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    stsArgs,
		})

		if err == nil && result != nil && result.Stdout != "" {
			var statefulSetList appsv1.StatefulSetList
			if jsonErr := json.Unmarshal([]byte(result.Stdout), &statefulSetList); jsonErr != nil {
				lastErr = fmt.Errorf("failed to parse statefulsets JSON: %v", jsonErr)
				if verbose {
					pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): StatefulSet JSON parse error\n", i+1, maxRetries)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Build set of found statefulsets
			foundStatefulSets := make(map[string]bool)
			for _, s := range statefulSetList.Items {
				foundStatefulSets[s.Name] = true
			}

			// Check if all expected statefulsets are present
			for _, expected := range expectedStatefulSets {
				if !foundStatefulSets[expected] {
					missingWorkloads = append(missingWorkloads, "statefulset/"+expected)
				}
			}
		} else {
			lastErr = fmt.Errorf("kubectl statefulsets error: %v", err)
			if verbose {
				pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): kubectl statefulsets error\n", i+1, maxRetries)
			}
			time.Sleep(retryInterval)
			continue
		}

		// Check if all workloads are present
		if len(missingWorkloads) == 0 {
			pterm.Success.Println("All ArgoCD workloads found.")
			return nil
		}

		lastErr = fmt.Errorf("missing workloads: %v", missingWorkloads)
		if verbose {
			pterm.Debug.Printf("Waiting for workloads (attempt %d/%d): missing %v\n", i+1, maxRetries, missingWorkloads)
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("workloads not found after %d retries: %w", maxRetries, lastErr)
}

// waitForAPIPort waits for the Kubernetes API port to be open before making API calls
// This prevents flooding a dead port with requests on Windows/WSL2 where the port
// might not be immediately available after k3d reports success
func (h *HelmManager) waitForAPIPort(ctx context.Context, timeout time.Duration) error {
	if h.kubeConfig == nil {
		return nil // Skip if no kubeConfig available
	}

	// Extract host:port from kubeConfig.Host
	apiAddress := strings.TrimPrefix(strings.TrimPrefix(h.kubeConfig.Host, "https://"), "http://")
	if apiAddress == "" {
		return nil // Skip if we can't determine the address
	}

	dialer := net.Dialer{Timeout: 2 * time.Second}
	pterm.Info.Printf("Waiting for API port %s to open...\n", apiAddress)

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		conn, err := dialer.DialContext(ctx, "tcp", apiAddress)
		if err == nil {
			conn.Close()
			pterm.Success.Printf("API port %s is open\n", apiAddress)
			return true, nil // Port is open!
		}
		return false, nil // Keep polling
	})
}

// verifyHelmRelease checks if a Helm release was actually created by running helm list
// This helps diagnose issues where Helm reports success but doesn't create resources
func (h *HelmManager) verifyHelmRelease(ctx context.Context, releaseName, namespace, clusterName string, verbose bool) error {
	if verbose {
		pterm.Info.Printf("Verifying Helm release '%s' in namespace '%s'...\n", releaseName, namespace)
	}

	// Build helm list args
	args := []string{"list", "-n", namespace, "--filter", releaseName, "-o", "json"}

	// Add explicit kube-context if cluster name is provided
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		args = append(args, "--kube-context", contextName)
	}

	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    args,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("failed to run helm list: %w", err)
	}

	// Log the helm list output
	if verbose {
		pterm.Info.Println("Helm list output:")
		pterm.Println(result.Stdout)
	}

	// Check if the release exists in the output
	// The JSON output will be an empty array "[]" if no releases found
	output := strings.TrimSpace(result.Stdout)
	if output == "" || output == "[]" {
		return fmt.Errorf("Helm release '%s' not found in namespace '%s' - helm list returned empty", releaseName, namespace)
	}

	// Also run helm status for more details
	statusArgs := []string{"status", releaseName, "-n", namespace}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		statusArgs = append(statusArgs, "--kube-context", contextName)
	}

	statusResult, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "helm",
		Args:    statusArgs,
		Env:     h.getHelmEnv(),
	})
	if err != nil {
		return fmt.Errorf("Helm release exists but status check failed: %w", err)
	}

	if verbose {
		pterm.Info.Println("Helm status output:")
		pterm.Println(statusResult.Stdout)
	}

	pterm.Success.Printf("Helm release '%s' verified successfully\n", releaseName)
	return nil
}

// showArgoCDDiagnostics outputs diagnostic information about ArgoCD pods when installation fails
// This helps identify why the Helm install timed out (e.g., image pull issues, crashloops, pending pods)
func (h *HelmManager) showArgoCDDiagnostics(ctx context.Context, clusterName string) {
	pterm.Warning.Println("=== ArgoCD Installation Diagnostics ===")

	// Build kubectl args with explicit context if cluster name is provided
	baseArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}

	// Get pod status
	pterm.Info.Println("Pod Status in argocd namespace:")
	podArgs := append(baseArgs, "get", "pods", "-n", "argocd", "-o", "wide")
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    podArgs,
	})
	if err == nil && result != nil {
		pterm.Println(result.Stdout)
		if result.Stderr != "" {
			pterm.Println(result.Stderr)
		}
	} else {
		pterm.Error.Printf("Failed to get pods: %v\n", err)
	}

	// Get events in the namespace (sorted by timestamp)
	pterm.Info.Println("\nRecent Events in argocd namespace:")
	eventArgs := append(baseArgs, "get", "events", "-n", "argocd", "--sort-by=.lastTimestamp")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    eventArgs,
	})
	if err == nil && result != nil {
		pterm.Println(result.Stdout)
		if result.Stderr != "" {
			pterm.Println(result.Stderr)
		}
	} else {
		pterm.Error.Printf("Failed to get events: %v\n", err)
	}

	// Get logs from all ArgoCD pods (helpful for CrashLoopBackOff debugging)
	// Use -o json to avoid Windows WSL escaping issues with jsonpath
	pterm.Info.Println("\nPod Logs (last 50 lines from each container):")
	podListArgs := append(baseArgs, "get", "pods", "-n", "argocd", "-o", "json")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    podListArgs,
	})
	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		var podList corev1.PodList
		pods := []string{}
		if jsonErr := json.Unmarshal([]byte(result.Stdout), &podList); jsonErr == nil {
			for _, p := range podList.Items {
				pods = append(pods, p.Name)
			}
		}
		for _, pod := range pods {
			if pod == "" {
				continue
			}
			// Get current logs
			pterm.Info.Printf("--- Logs from pod: %s ---\n", pod)
			logArgs := append(baseArgs, "logs", pod, "-n", "argocd", "--all-containers=true", "--tail=50")
			logResult, logErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    logArgs,
			})
			if logErr == nil && logResult != nil && strings.TrimSpace(logResult.Stdout) != "" {
				pterm.Println(logResult.Stdout)
			} else if logErr != nil {
				pterm.Debug.Printf("No current logs available for %s\n", pod)
			}

			// Get previous logs (from crashed containers)
			prevLogArgs := append(baseArgs, "logs", pod, "-n", "argocd", "--all-containers=true", "--tail=50", "--previous")
			prevLogResult, prevLogErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    prevLogArgs,
			})
			if prevLogErr == nil && prevLogResult != nil && strings.TrimSpace(prevLogResult.Stdout) != "" {
				pterm.Info.Printf("--- Previous logs from pod: %s (crashed container) ---\n", pod)
				pterm.Println(prevLogResult.Stdout)
			}
		}
	}

	// Describe pods that are not Running
	pterm.Info.Println("\nDescribing non-running pods:")
	describeArgs := append(baseArgs, "get", "pods", "-n", "argocd", "--field-selector=status.phase!=Running", "-o", "name")
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    describeArgs,
	})
	if err == nil && result != nil && strings.TrimSpace(result.Stdout) != "" {
		pods := strings.Split(strings.TrimSpace(result.Stdout), "\n")
		for _, pod := range pods {
			if pod == "" {
				continue
			}
			pterm.Info.Printf("--- Describing pod: %s ---\n", pod)
			describePodArgs := append(baseArgs, "describe", "pod", pod, "-n", "argocd")
			descResult, descErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
				Command: "kubectl",
				Args:    describePodArgs,
			})
			if descErr == nil && descResult != nil {
				pterm.Println(descResult.Stdout)
			}
		}
	}

	pterm.Warning.Println("=== End of Diagnostics ===")
}

// verifyClusterConnectivity verifies that the cluster is reachable before running helm commands
// This is important after idle periods where WSL networking may have gone stale
// It also logs kubeconfig details to help diagnose connectivity issues
func (h *HelmManager) verifyClusterConnectivity(ctx context.Context, config config.ChartInstallConfig) error {
	pterm.Info.Println("Verifying cluster connectivity before app-of-apps installation...")

	// Build kubectl args with explicit context if cluster name is provided
	kubectlArgs := []string{"cluster-info"}
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		kubectlArgs = []string{"--context", contextName, "cluster-info"}
	}

	// On Windows/WSL, also dump kubeconfig details for debugging
	if runtime.GOOS == "windows" {
		h.debugWSLKubeconfig(ctx, config.Verbose)
	}

	// Retry kubectl cluster-info a few times (cluster may need a moment after idle)
	maxRetries := 5
	retryDelay := 2 * time.Second
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    kubectlArgs,
		})

		if err == nil && result.ExitCode == 0 {
			pterm.Success.Println("Cluster is reachable")
			if config.Verbose {
				pterm.Info.Println("kubectl cluster-info output:")
				pterm.Println(result.Stdout)
			}
			return nil
		}

		lastErr = err
		if config.Verbose {
			pterm.Warning.Printf("Cluster connectivity check attempt %d/%d failed: %v\n", i+1, maxRetries, err)
			if result != nil {
				if result.Stdout != "" {
					pterm.Println("stdout:", result.Stdout)
				}
				if result.Stderr != "" {
					pterm.Println("stderr:", result.Stderr)
				}
			}
		}

		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("cluster not reachable after %d attempts: %w", maxRetries, lastErr)
}

// debugWSLKubeconfig logs kubeconfig details from WSL for debugging connectivity issues
func (h *HelmManager) debugWSLKubeconfig(ctx context.Context, verbose bool) {
	if !verbose {
		return
	}

	pterm.Info.Println("=== WSL Kubeconfig Debug Info ===")

	// Get the WSL user (same logic as executor)
	wslUser := os.Getenv("WSL_USER")
	if wslUser == "" {
		wslUser = "runner"
	}

	// Check if kubeconfig exists
	checkCmd := fmt.Sprintf("ls -la ~/.kube/config 2>&1 || echo 'Kubeconfig not found'")
	result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", checkCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Kubeconfig file status:")
		pterm.Println(result.Stdout)
	}

	// Show the server addresses in kubeconfig (without showing secrets)
	serverCmd := "grep -A2 'cluster:' ~/.kube/config 2>/dev/null | grep 'server:' || echo 'No server entries found'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", serverCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Server addresses in kubeconfig:")
		pterm.Println(result.Stdout)
	}

	// Show current context
	contextCmd := "kubectl config current-context 2>&1 || echo 'No current context'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", contextCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Current kubectl context:")
		pterm.Println(result.Stdout)
	}

	// Check if the API server port is reachable
	portCheckCmd := "nc -zv 127.0.0.1 6550 2>&1 || echo 'Port 6550 not reachable'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", portCheckCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("Port 6550 connectivity check:")
		pterm.Println(result.Stdout)
	}

	// Show Docker containers (k3d nodes)
	dockerCmd := "docker ps --filter 'name=k3d' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>&1 || echo 'Docker not available or no k3d containers'"
	result, err = h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "wsl",
		Args:    []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", dockerCmd},
	})
	if err == nil && result != nil {
		pterm.Info.Println("k3d Docker containers:")
		pterm.Println(result.Stdout)
	}

	pterm.Info.Println("=== End WSL Kubeconfig Debug ===")
}

