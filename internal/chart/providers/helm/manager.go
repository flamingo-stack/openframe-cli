package helm

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
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
	// The API server's certificate is issued to the cluster name or specific hostnames,
	// which may not match when connecting via 127.0.0.1 from Windows/WSL2.
	// This is safe for local development clusters and solves handshake failures.
	// Applied here as defense-in-depth in case the caller's config doesn't have it set.
	config.Insecure = true
	config.TLSClientConfig.CAData = nil
	config.TLSClientConfig.CAFile = ""

	if verbose {
		pterm.Debug.Println("TLS verification bypassed for local k3d cluster (HelmManager)")
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
	if _, err := tmpFile.WriteString(argocd.GetArgoCDValues()); err != nil {
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
		"--version=8.2.7",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
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

		// Include stderr output for better debugging
		if result != nil && result.Stderr != "" {
			return fmt.Errorf("failed to install ArgoCD: %w\nHelm output: %s", err, result.Stderr)
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
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.8/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.8/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.8/manifests/crds/appproject-crd.yaml",
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
	if _, err := tmpFile.WriteString(argocd.GetArgoCDValues()); err != nil {
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
		pterm.Info.Printf("   Version: 8.2.7\n")
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
		"--version=8.2.7",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
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
		// Include stderr output for better debugging
		if result != nil && result.Stderr != "" {
			return fmt.Errorf("failed to install ArgoCD: %w\nHelm output: %s", err, result.Stderr)
		}
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Wait for ArgoCD deployments to be created after Helm install
	// This addresses the race condition where Helm --wait returns before Kubernetes
	// has actually created the Deployment objects (common in k3d/CI environments)
	//
	// Use native Go client for all platforms (including Windows) for fast, reliable polling
	// The kubeClient uses the same kubeconfig that was used to create the cluster
	if h.kubeClient != nil {
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
		// Fallback to kubectl-based verification when native Go client is unavailable
		if config.Verbose {
			pterm.Warning.Println("Native Go client unavailable, using kubectl for deployment verification")
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

	// Convert Windows paths to WSL paths if needed (for Helm running in WSL2)
	valuesFilePath := appConfig.ValuesFile
	certFilePath := certFile
	keyFilePath := keyFile

	if runtime.GOOS == "windows" {
		var err error

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
		"upgrade", "--install", "app-of-apps", appConfig.ChartPath,
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
func (h *HelmManager) convertWindowsPathToWSL(windowsPath string) (string, error) {
	if windowsPath == "" {
		return "", fmt.Errorf("empty path provided")
	}

	// Replace backslashes with forward slashes
	path := strings.ReplaceAll(windowsPath, "\\", "/")

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

// waitForArgoCDDeployments waits for ArgoCD deployments to be created in the cluster
// This addresses the race condition where Helm's --wait returns before Kubernetes
// has actually created the Deployment objects (common in k3d/CI environments)
//
// NOTE: CRDs are now installed and verified BEFORE Helm runs (see InstallArgoCDWithProgress),
// so this function focuses only on verifying the deployments exist.
func (h *HelmManager) waitForArgoCDDeployments(ctx context.Context, verbose bool) error {
	if h.kubeClient == nil {
		return fmt.Errorf("Kubernetes core client not initialized")
	}

	// Wait for API port to be available before making API calls
	// This prevents flooding a dead port with requests on Windows/WSL2
	if err := h.waitForAPIPort(ctx, 45*time.Second); err != nil {
		return fmt.Errorf("API port never opened: %w", err)
	}

	// List of expected deployments
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
		"argocd-application-controller",
	}

	// CRITICAL: Use extended timeout since cluster operations can be slow
	// Native API calls are fast (~ms), so we use frequent polling with longer total timeout
	timeout := 90 * time.Second      // 90 seconds for slow CI/Windows environments
	retryInterval := 1 * time.Second // Fast polling interval (native API is ~ms per call)

	pterm.Info.Println("Waiting for ArgoCD deployments via NATIVE API...")

	// Use wait.PollUntilContextTimeout for resilient polling
	return wait.PollUntilContextTimeout(ctx, retryInterval, timeout, false, func(ctx context.Context) (bool, error) {

		missingDeployments := []string{}

		for _, name := range expectedDeployments {
			// Native API call to check for deployment existence
			_, err := h.kubeClient.AppsV1().Deployments("argocd").Get(ctx, name, metav1.GetOptions{})

			if k8serrors.IsNotFound(err) {
				missingDeployments = append(missingDeployments, name)
			} else if err != nil {
				// If it's a transient API error (not 'Not Found'), log and retry
				pterm.Warning.Printf("Transient API error checking deployment %s: %v\n", name, err)
				return false, nil
			}
		}

		if len(missingDeployments) == 0 {
			pterm.Success.Println("All ArgoCD deployments found.")
			return true, nil // Success: All deployments exist.
		}

		if verbose {
			pterm.Debug.Printf("Still missing deployments: %v\n", missingDeployments)
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

// waitForArgoCDDeploymentsKubectl waits for ArgoCD deployments using kubectl
// This is a fallback for when the native Go client is unavailable (e.g., Windows/WSL)
func (h *HelmManager) waitForArgoCDDeploymentsKubectl(ctx context.Context, clusterName string, verbose bool) error {
	// List of expected deployments
	expectedDeployments := []string{
		"argocd-server",
		"argocd-repo-server",
		"argocd-application-controller",
	}

	// Wait settings - increased for slow CI environments
	maxRetries := 40           // 40 retries * 3 seconds = 120 seconds max (2 minutes)
	retryInterval := 3 * time.Second
	initialDelay := 5 * time.Second // Give Kubernetes time to create resources after Helm completes

	pterm.Info.Println("Waiting for ArgoCD deployments to appear via kubectl...")

	// Initial delay: Helm's --wait returns before Kubernetes controllers fully create Deployments
	// This delay allows the deployment controller to process the Helm release and create resources
	if verbose {
		pterm.Debug.Printf("Initial delay of %v to allow Kubernetes controllers to create resources...\n", initialDelay)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(initialDelay):
	}

	// Build kubectl args with explicit context if cluster name is provided
	// Use a single kubectl call to list all deployments in the namespace
	baseArgs := []string{}
	if clusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", clusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}
	baseArgs = append(baseArgs, "-n", "argocd", "get", "deployments", "-o", "jsonpath={.items[*].metadata.name}")

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Single kubectl call to get all deployment names in the namespace
		result, err := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
			Command: "kubectl",
			Args:    baseArgs,
		})

		if err == nil && result != nil {
			foundDeployments := strings.Fields(strings.TrimSpace(result.Stdout))
			foundSet := make(map[string]bool)
			for _, d := range foundDeployments {
				foundSet[d] = true
			}

			// Check if all expected deployments are present
			allFound := true
			var missingDeployments []string
			for _, expected := range expectedDeployments {
				if !foundSet[expected] {
					allFound = false
					missingDeployments = append(missingDeployments, expected)
				}
			}

			if allFound {
				pterm.Success.Println("All ArgoCD deployments found.")
				return nil
			}

			lastErr = fmt.Errorf("missing deployments: %v", missingDeployments)
			if verbose {
				pterm.Debug.Printf("Waiting for deployments (attempt %d/%d): missing %v\n", i+1, maxRetries, missingDeployments)
			}
		} else {
			lastErr = fmt.Errorf("kubectl error: %v", err)
			if verbose {
				pterm.Debug.Printf("Waiting for deployments (attempt %d/%d): kubectl error\n", i+1, maxRetries)
			}
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("deployments not found after %d retries: %w", maxRetries, lastErr)
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

