package bootstrap

import (
	"context"
	"fmt"
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

// bootstrap executes cluster create followed by chart install.
//
// There is no Windows-specific WSL bootstrapping here: on Windows the root
// command forwards the whole CLI into WSL before any command runs (see
// wsllauncher), so this code only ever executes as a Linux process. The old
// initializeWSL PowerShell step was an unreachable, conflicting second WSL
// strategy — it hardcoded the "Ubuntu" distro (the launcher is distro-agnostic
// via OPENFRAME_WSL_DISTRO) and created a `runner:runner` account with
// NOPASSWD sudo, a CI artifact that had no business in a released binary.
func (s *Service) bootstrap(ctx context.Context, clusterName string, nonInteractive, verbose bool) error {
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
