package bootstrap

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	chartServices "github.com/flamingo-stack/openframe-cli/internal/chart/services"
	utilTypes "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
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

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		force = false
	}

	// Get repo/branch overrides
	githubRepo, err := cmd.Flags().GetString("repo")
	if err != nil {
		githubRepo = ""
	}

	githubBranch, err := cmd.Flags().GetString("branch")
	if err != nil {
		githubBranch = ""
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

	err = s.bootstrap(clusterName, deploymentMode, nonInteractive, force, verbose, githubRepo, githubBranch)
	if err != nil {
		// Use shared error handler for consistent error display (same as chart install)
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

// bootstrap executes cluster create followed by chart install
func (s *Service) bootstrap(clusterName, deploymentMode string, nonInteractive, force, verbose bool, githubRepo, githubBranch string) error {
	// Show logo first
	ui.ShowLogo()

	// On Windows, initialize WSL2 first before anything else
	if runtime.GOOS == "windows" {
		if err := s.initializeWSL(verbose); err != nil {
			return fmt.Errorf("failed to initialize WSL: %w", err)
		}
	}

	// Unified pre-flight check: verify ALL prerequisites (cluster + chart) before
	// creating a cluster. This prevents the user from waiting 5-10 minutes for
	// cluster creation only to fail on chart prerequisites like memory or mkcert.
	preflight := NewPreflightChecker(nonInteractive, force, verbose)
	if err := preflight.CheckAll(); err != nil {
		return err
	}

	// Normalize cluster name (use default if empty)
	config := s.buildClusterConfig(clusterName)
	actualClusterName := config.Name

	// Step 1: Create cluster (skip prerequisite checks — already done above)
	kubeConfig, err := s.createClusterSkipPrereqs(actualClusterName, verbose, nonInteractive)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Add spacing between commands
	fmt.Println()
	fmt.Println()

	// Step 2: Install charts with deployment mode flags on the created cluster
	// Skip chart prerequisites — already checked by preflight
	if err := s.installChartWithMode(actualClusterName, deploymentMode, nonInteractive, verbose, kubeConfig, githubRepo, githubBranch); err != nil {
		return fmt.Errorf("failed to install charts: %w", err)
	}

	return nil
}

// createClusterSkipPrereqs creates a cluster without re-running prerequisite checks.
// Prerequisites have already been verified by the unified preflight checker.
func (s *Service) createClusterSkipPrereqs(clusterName string, verbose bool, nonInteractive bool) (*rest.Config, error) {
	return cluster.CreateClusterSkipPrerequisites(clusterName, verbose, nonInteractive)
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
func (s *Service) installChartWithMode(clusterName, deploymentMode string, nonInteractive, verbose bool, kubeConfig *rest.Config, githubRepo, githubBranch string) error {
	// githubRepo is left empty when user didn't pass --repo;
	// buildConfiguration() will derive it from deployment mode.
	if githubBranch == "" {
		githubBranch = "main"
	}

	return chartServices.InstallChartsWithConfig(utilTypes.InstallationRequest{
		Args:              []string{clusterName},
		Force:             false,
		DryRun:            false,
		Verbose:           verbose,
		GitHubRepo:        githubRepo,
		GitHubBranch:      githubBranch,
		CertDir:           "", // Auto-detected
		DeploymentMode:    deploymentMode,
		NonInteractive:    nonInteractive,
		KubeConfig:        kubeConfig,
		SkipPrerequisites: true,                // Already checked by preflight
		SkipConfigPrompts: deploymentMode != "", // Skip config mode selection when deployment mode is pre-set
	})
}

// initializeWSL initializes WSL2 with Ubuntu and configures the runner user
// This must run before any tools installation or cluster creation on Windows
func (s *Service) initializeWSL(verbose bool) error {
	fmt.Println("Initializing WSL2 with Ubuntu...")

	// PowerShell script to initialize WSL2
	script := `
$ErrorActionPreference = 'Continue'

# Install WSL2 with Ubuntu
echo Y | wsl --install -d Ubuntu --no-launch
Start-Sleep -Seconds 20
wsl --set-default-version 2
wsl --list --verbose
if ($LASTEXITCODE -ne 0) { exit 1 }

# Initialize Ubuntu with retries
$maxRetries = 5
$retryDelay = 10
for ($i = 1; $i -le $maxRetries; $i++) {
    wsl -d Ubuntu -u root bash -c "echo 'init'" 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) { break }
    Start-Sleep -Seconds $retryDelay
    $retryDelay += 5
}
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to initialize Ubuntu"
    exit 1
}
Start-Sleep -Seconds 10

# Create runner user with sudo access
wsl -d Ubuntu -u root bash -c "id runner 2>/dev/null || (useradd -m -s /bin/bash runner && echo 'runner:runner' | chpasswd && usermod -aG sudo runner)"
if ($LASTEXITCODE -ne 0) { exit 1 }

wsl -d Ubuntu -u root bash -c "grep -q '%sudo ALL=(ALL) NOPASSWD:ALL' /etc/sudoers || echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers"
wsl -d Ubuntu -u runner bash -c "whoami" | Out-Null
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "WSL2 configured successfully"
`

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if verbose {
		cmd.Stdout = nil // Will use default (os.Stdout)
		cmd.Stderr = nil // Will use default (os.Stderr)
	}

	output, err := cmd.CombinedOutput()
	if verbose {
		fmt.Println(string(output))
	}

	if err != nil {
		if verbose {
			fmt.Printf("WSL initialization output: %s\n", string(output))
		}
		return fmt.Errorf("WSL initialization failed: %w", err)
	}

	fmt.Println("✓ WSL2 initialized successfully")
	return nil
}
