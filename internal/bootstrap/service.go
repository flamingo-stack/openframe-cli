package bootstrap

import (
	"fmt"
	"runtime"
	"strings"

	chartServices "github.com/flamingo-stack/openframe-cli/internal/chart/services"
	chartErrors "github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	utilTypes "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// Service provides bootstrap functionality
type Service struct{}

// NewService creates a new bootstrap service
func NewService() *Service {
	return &Service{}
}

// Execute handles the bootstrap command execution
func (s *Service) Execute(cmd *cobra.Command, args []string) error {
	// Get verbose flag - first check local flag, then root command
	verbose := false
	if localVerbose, err := cmd.Flags().GetBool("verbose"); err == nil {
		verbose = localVerbose
	}
	if !verbose {
		if rootVerbose, err := cmd.Root().PersistentFlags().GetBool("verbose"); err == nil {
			verbose = rootVerbose
		}
	}

	// Get deployment mode flags
	deploymentMode, err := cmd.Flags().GetString("deployment-mode")
	if err != nil {
		deploymentMode = ""
	}

	nonInteractive, err := cmd.Flags().GetBool("non-interactive")
	if err != nil {
		nonInteractive = false
	}

	// Validate deployment mode
	if deploymentMode != "" {
		validModes := []string{"oss-tenant", "saas-tenant", "saas-shared"}
		isValid := false
		for _, mode := range validModes {
			if deploymentMode == mode {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid deployment mode: %s. Valid options: oss-tenant, saas-tenant, saas-shared", deploymentMode)
		}
	}

	// Validate non-interactive requires deployment mode
	if nonInteractive && deploymentMode == "" {
		return fmt.Errorf("--deployment-mode is required when using --non-interactive")
	}

	// Get cluster name from args if provided
	var clusterName string
	if len(args) > 0 {
		clusterName = strings.TrimSpace(args[0])
	}

	err = s.bootstrap(clusterName, deploymentMode, nonInteractive, verbose)
	if err != nil {
		// Use shared error handler for consistent error display (same as chart install)
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

// bootstrap executes cluster create followed by chart install
func (s *Service) bootstrap(clusterName, deploymentMode string, nonInteractive, verbose bool) error {
	// Normalize cluster name (use default if empty)
	config := s.buildClusterConfig(clusterName)
	actualClusterName := config.Name

	// Step 1: Create cluster with suppressed UI and get the rest.Config
	kubeConfig, err := s.createClusterSuppressed(actualClusterName, verbose, nonInteractive)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Add spacing between commands
	fmt.Println()
	fmt.Println()

	// Step 2: Install charts with deployment mode flags on the created cluster
	if err := s.installChartWithMode(actualClusterName, deploymentMode, nonInteractive, verbose, kubeConfig); err != nil {
		return fmt.Errorf("failed to install charts: %w", err)
	}

	return nil
}

// createClusterSuppressed creates a cluster with suppressed UI elements
// Returns the *rest.Config for the created cluster
func (s *Service) createClusterSuppressed(clusterName string, verbose bool, nonInteractive bool) (*rest.Config, error) {
	// Use the wrapper function that includes prerequisite checks
	return cluster.CreateClusterWithPrerequisitesNonInteractive(clusterName, verbose, nonInteractive)
}

// buildClusterConfig builds a cluster configuration from the cluster name
func (s *Service) buildClusterConfig(clusterName string) models.ClusterConfig {
	if clusterName == "" {
		clusterName = "openframe-dev" // default name
	}

	return models.ClusterConfig{
		Name:       clusterName,
		Type:       models.ClusterTypeK3d,
		K8sVersion: "",
		NodeCount:  4,
	}
}

// installChartWithMode installs charts with deployment mode flags
func (s *Service) installChartWithMode(clusterName, deploymentMode string, nonInteractive, verbose bool, kubeConfig *rest.Config) error {
	// Use the chart installation function with deployment mode flags
	err := chartServices.InstallChartsWithConfig(utilTypes.InstallationRequest{
		Args:           []string{clusterName},
		Force:          false,
		DryRun:         false,
		Verbose:        verbose,
		GitHubRepo:     "https://github.com/flamingo-stack/openframe-oss-tenant", // Default repository
		GitHubBranch:   "main",                                                   // Default branch
		CertDir:        "",                                                       // Auto-detected
		DeploymentMode: deploymentMode,
		NonInteractive: nonInteractive,
		KubeConfig:     kubeConfig,
	})

	if err != nil {
		// Handle recoverable errors specially on Windows/non-interactive mode (CI)
		// Registry DNS issues are common in WSL2/GitHub Actions and shouldn't fail the whole bootstrap
		if s.shouldSoftFailOnError(err, nonInteractive) {
			s.printRecoverableErrorWarning(err, verbose)
			return nil // Soft-fail: cluster is created, ArgoCD can be retried
		}
		return err
	}

	return nil
}

// shouldSoftFailOnError determines if an error should be treated as a soft failure
// This is used for CI scenarios on Windows where registry DNS issues are common
func (s *Service) shouldSoftFailOnError(err error, nonInteractive bool) bool {
	// Only soft-fail in non-interactive mode (CI) on Windows
	if !nonInteractive || runtime.GOOS != "windows" {
		return false
	}

	// Check for registry DNS errors (recoverable infrastructure issues)
	if chartErrors.IsRegistryDNSError(err) {
		return true
	}

	// Check for recoverable chart errors
	if chartErrors.IsRecoverable(err) {
		return true
	}

	// Check for registry DNS patterns in the error message
	if chartErrors.IsHelmTimeoutWithRegistryDNS(err) {
		return true
	}

	// On Windows/WSL2, any Helm pre-install timeout is likely caused by registry DNS issues
	// The actual DNS error is in kubectl events, not in the helm error message
	if chartErrors.IsHelmPreInstallTimeout(err) {
		return true
	}

	return false
}

// printRecoverableErrorWarning prints a warning about a recoverable error
func (s *Service) printRecoverableErrorWarning(err error, verbose bool) {
	fmt.Println()
	pterm.Warning.Println("Chart installation encountered a recoverable infrastructure error")
	fmt.Println()

	// Print specific message based on error type
	if regErr, ok := err.(*chartErrors.RegistryDNSError); ok {
		pterm.Info.Printf("Registry DNS issue detected: %s\n", regErr.Registry)
		fmt.Println()
		pterm.Info.Println("This is a known issue with WSL2/Docker networking on Windows CI.")
		pterm.Info.Println("The cluster was created successfully - ArgoCD installation can be retried.")
		fmt.Println()
		pterm.Info.Println("Troubleshooting steps:")
		for _, suggestion := range regErr.GetTroubleshootingSteps() {
			pterm.Printf("  • %s\n", suggestion)
		}
	} else if chartErrors.IsHelmPreInstallTimeout(err) {
		// Helm pre-install timeout on Windows - likely registry DNS issue
		pterm.Info.Println("Helm pre-install timed out waiting for pods to start.")
		fmt.Println()
		pterm.Info.Println("This is typically caused by pods failing to pull container images")
		pterm.Info.Println("due to WSL2/Docker networking issues (registry DNS resolution).")
		fmt.Println()
		pterm.Info.Println("The cluster was created successfully - ArgoCD installation can be retried.")
		fmt.Println()
		pterm.Info.Println("Troubleshooting steps:")
		pterm.Printf("  • Check kubectl events: kubectl get events -n argocd\n")
		pterm.Printf("  • Check WSL2 DNS: wsl cat /etc/resolv.conf\n")
		pterm.Printf("  • Test registry connectivity: wsl curl -I https://registry-1.docker.io/v2/\n")
		pterm.Printf("  • Restart Docker: wsl sudo systemctl restart docker\n")
		pterm.Printf("  • Retry: openframe chart install\n")
	} else {
		pterm.Info.Printf("Error: %v\n", err)
		fmt.Println()
		pterm.Info.Println("The cluster was created successfully.")
		pterm.Info.Println("You can retry chart installation with: openframe chart install")
	}

	fmt.Println()
	pterm.Success.Println("Bootstrap completed with warnings (cluster created, charts partially installed)")
}
