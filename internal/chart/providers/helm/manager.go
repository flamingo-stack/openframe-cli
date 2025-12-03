package helm

import (
	"context"
	"fmt"
	"io"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

// HelmManager handles Helm operations
type HelmManager struct {
	executor      executor.CommandExecutor
	kubeConfig    *rest.Config      // Stores the cluster connection config
	dynamicClient dynamic.Interface // Dynamic client for programmatic resource management
	verbose       bool              // Enable verbose logging
}

// NewHelmManager creates a new Helm manager
func NewHelmManager(exec executor.CommandExecutor) *HelmManager {
	return &HelmManager{
		executor: exec,
	}
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
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--version=8.2.7",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
		"-f", valuesFilePath,
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
	if !config.SkipCRDs {
		if config.Verbose {
			pterm.Info.Println("Installing ArgoCD CRDs using native Go client...")
		}

		// Build kube context for the dynamic client
		kubeContext := ""
		if config.ClusterName != "" {
			kubeContext = fmt.Sprintf("k3d-%s", config.ClusterName)
		}

		// Initialize the dynamic client for programmatic CRD installation
		// This reduces reliance on external kubectl binary
		if err := h.initDynamicClient(kubeContext); err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			return fmt.Errorf("failed to initialize Kubernetes client for CRD installation: %w", err)
		}

		// Set verbose mode for debug logging
		h.verbose = config.Verbose

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

	// Install ArgoCD with upgrade --install
	args := []string{
		"upgrade", "--install", "argo-cd", "argo/argo-cd",
		"--version=8.2.7",
		"--namespace", "argocd",
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
		"-f", valuesFilePath,
	}

	// Add explicit kube-context if cluster name is provided (important for Windows/WSL)
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		args = append(args, "--kube-context", contextName)
	}

	if config.DryRun {
		args = append(args, "--dry-run")
		if config.Verbose {
			pterm.Info.Println("🔍 Running in dry-run mode...")
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

	// Verify that ArgoCD deployments exist after Helm install
	// This catches cases where Helm --wait returns too quickly
	if config.Verbose {
		pterm.Debug.Println("Verifying ArgoCD deployments were created...")
	}

	// Build kubectl args with explicit context if cluster name is provided
	verifyArgs := []string{"-n", "argocd", "get", "deployments", "-l", "app.kubernetes.io/part-of=argocd", "-o", "jsonpath={.items[*].metadata.name}"}
	if config.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", config.ClusterName)
		verifyArgs = append([]string{"--context", contextName}, verifyArgs...)
	}

	verifyResult, verifyErr := h.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    verifyArgs,
	})

	if verifyErr != nil || verifyResult == nil || strings.TrimSpace(verifyResult.Stdout) == "" {
		if spinner != nil {
			spinner.Stop()
		}
		// Helm reported success but no ArgoCD deployments exist - this is an error condition
		pterm.Warning.Println("Helm install reported success but no ArgoCD deployments found")
		pterm.Info.Println("This may indicate a Helm caching issue or WSL connectivity problem")

		return fmt.Errorf("ArgoCD Helm install completed but no deployments were created - this may indicate a Helm or cluster connectivity issue")
	}

	if config.Verbose {
		pterm.Debug.Printf("Found ArgoCD deployments: %s\n", verifyResult.Stdout)
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

// initDynamicClient initializes the Kubernetes dynamic client for the given context
// This reduces reliance on external kubectl binary for resource management
func (h *HelmManager) initDynamicClient(kubeContext string) error {
	// Build config from kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	h.kubeConfig = config

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	h.dynamicClient = dynamicClient
	return nil
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
