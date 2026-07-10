package bootstrap

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	chartmodels "github.com/flamingo-stack/openframe-cli/internal/chart/models"
	chartServices "github.com/flamingo-stack/openframe-cli/internal/chart/services"
	utilTypes "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// defaultClusterName is used when the user doesn't name the cluster.
const defaultClusterName = "openframe-dev"

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

	nonInteractive, err := cmd.Flags().GetBool("non-interactive")
	if err != nil {
		nonInteractive = false
	}

	// Get cluster name from args if provided
	var clusterName string
	if len(args) > 0 {
		clusterName = strings.TrimSpace(args[0])
	}

	err = s.bootstrap(cmd.Context(), clusterName, nonInteractive, verbose)
	if err != nil {
		// Use shared error handler for consistent error display (same as chart install)
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

// bootstrap executes cluster create followed by chart install
func (s *Service) bootstrap(ctx context.Context, clusterName string, nonInteractive, verbose bool) error {
	// On Windows, initialize WSL2 first before anything else
	if runtime.GOOS == "windows" {
		if err := s.initializeWSL(verbose); err != nil {
			return fmt.Errorf("failed to initialize WSL: %w", err)
		}
	}

	// Normalize cluster name (use default if empty)
	actualClusterName := clusterName
	if actualClusterName == "" {
		actualClusterName = defaultClusterName
	}

	// Step 1: Create cluster with suppressed UI and get the rest.Config
	kubeConfig, err := s.createClusterSuppressed(ctx, actualClusterName, verbose, nonInteractive)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	// Add spacing between commands. DefaultBasicText, not raw fmt: --silent
	// redirects it — these two raw Printlns were the "three blank lines" the
	// 0.4.7 verification report found in an otherwise silent bootstrap log.
	pterm.DefaultBasicText.Println()
	pterm.DefaultBasicText.Println()

	// Step 2: Install charts on the created cluster
	if err := s.installChart(ctx, actualClusterName, nonInteractive, verbose, kubeConfig); err != nil {
		return fmt.Errorf("failed to install charts: %w", err)
	}

	return nil
}

// createClusterSuppressed creates a cluster with suppressed UI elements
// Returns the *rest.Config for the created cluster
func (s *Service) createClusterSuppressed(ctx context.Context, clusterName string, verbose bool, nonInteractive bool) (*rest.Config, error) {
	// Use the wrapper function that includes prerequisite checks
	return cluster.CreateClusterWithPrerequisitesNonInteractive(ctx, clusterName, verbose, nonInteractive)
}

// installChart installs charts on the created cluster
func (s *Service) installChart(ctx context.Context, clusterName string, nonInteractive, verbose bool, kubeConfig *rest.Config) error {
	return chartServices.InstallChartsWithConfigContext(ctx, utilTypes.InstallationRequest{
		Args:           []string{clusterName},
		Force:          false,
		DryRun:         false,
		Verbose:        verbose,
		GitHubRepo:     chartmodels.RepoOSSTenant,    // Default repository
		GitHubBranch:   chartmodels.DefaultGitBranch, // Default branch
		CertDir:        "",                           // Auto-detected
		NonInteractive: nonInteractive,
		KubeConfig:     kubeConfig,
		// Inject cluster access from the orchestrator (composition root) so the
		// app subsystem stays isolated from cluster-creation code (req 18/19).
		ClusterAccess: cluster.NewClusterService(executor.NewRealCommandExecutor(false, verbose)),
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
