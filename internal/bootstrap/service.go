package bootstrap

import (
	"context"
	"fmt"
	"strings"

	chartmodels "github.com/flamingo-stack/openframe-cli/internal/chart/models"
	chartServices "github.com/flamingo-stack/openframe-cli/internal/chart/services"
	utilTypes "github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/steps"
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

	// The buildkit-style stage checklist: every stage announces itself, and
	// closes with ✔/✖ + duration; the collected timings feed the summary card.
	tracker := steps.NewTracker("Validate helm values", "Create cluster", "Install platform")

	// Step 0: Pre-flight the helm values file BEFORE creating the cluster. A
	// malformed `argocd:` override (or unparseable YAML) otherwise costs a full
	// cluster create before the chart install rejects the same file.
	tracker.Begin(0, "")
	if err := chartServices.ValidateHelmValuesFile(); err != nil {
		tracker.Fail(0)
		return err
	}
	tracker.Done(0)

	// Step 1: Create cluster with suppressed UI and get the rest.Config
	tracker.Begin(1, actualClusterName)
	kubeConfig, err := s.createClusterSuppressed(ctx, actualClusterName, verbose, nonInteractive)
	if err != nil {
		tracker.Fail(1)
		ui.Notify("OpenFrame bootstrap failed")
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	tracker.Done(1)

	// Add spacing between commands. DefaultBasicText, not raw fmt: --silent
	// redirects it — these two raw Printlns were the "three blank lines" the
	// 0.4.7 verification report found in an otherwise silent bootstrap log.
	pterm.DefaultBasicText.Println()
	pterm.DefaultBasicText.Println()

	// Step 2: Install charts on the created cluster
	tracker.Begin(2, "ArgoCD + app-of-apps")
	if err := s.installChart(ctx, actualClusterName, nonInteractive, verbose, kubeConfig); err != nil {
		tracker.Fail(2)
		ui.Notify("OpenFrame bootstrap failed")
		return fmt.Errorf("failed to install charts: %w", err)
	}
	tracker.Done(2)

	printBootstrapSummary(actualClusterName, tracker)
	// Desktop toast + bell for the user who switched away during the install;
	// no-op off-TTY and under --silent.
	ui.Notify(fmt.Sprintf("OpenFrame ready %s %s", ui.Glyphs().Bullet, steps.FormatDuration(tracker.Total())))
	return nil
}

// printBootstrapSummary is the closing card: what was built, how long each
// stage took, and where to go next.
func printBootstrapSummary(clusterName string, tracker *steps.Tracker) {
	g := ui.Glyphs()
	pterm.DefaultBasicText.Println()
	pterm.Success.Printf("OpenFrame ready %s %s\n", g.Bullet, steps.FormatDuration(tracker.Total()))
	ui.SummaryRow("cluster", clusterName+" (context "+k8s.ResolveContextForCluster(k8s.DefaultKubeconfigPath(), clusterName)+")")
	ui.SummaryRow("stages", steps.TimingsLine(tracker.Timings()))
	ui.SummaryRow("status", "openframe app status")
	ui.SummaryRow("access", "openframe app access   (ArgoCD UI credentials)")
}

// createClusterSuppressed creates a cluster with suppressed UI elements
// Returns the *rest.Config for the created cluster
func (s *Service) createClusterSuppressed(ctx context.Context, clusterName string, verbose bool, nonInteractive bool) (*rest.Config, error) {
	// Use the wrapper function that includes prerequisite checks
	return cluster.CreateClusterWithPrerequisitesNonInteractive(ctx, clusterName, verbose, nonInteractive)
}

// installChart installs charts on the created cluster
func (s *Service) installChart(ctx context.Context, clusterName string, nonInteractive, verbose bool, kubeConfig *rest.Config) error {
	return chartServices.InstallChartsWithConfigContext(ctx, bootstrapInstallRequest(clusterName, nonInteractive, verbose, kubeConfig))
}

// bootstrapInstallRequest builds the chart-install request for the cluster the
// bootstrap just created. KubeContext must be set alongside KubeConfig: the
// workflow ignores Args entirely once a rest.Config is provided, so without it
// the target name came through empty — the interactive confirmation asked
// "install OpenFrame chart on ”?" and every helm call ran WITHOUT
// --kube-context, silently targeting the kubeconfig's current context instead
// of the cluster the native client was pointed at.
func bootstrapInstallRequest(clusterName string, nonInteractive, verbose bool, kubeConfig *rest.Config) utilTypes.InstallationRequest {
	return utilTypes.InstallationRequest{
		Args:           []string{clusterName},
		Force:          false,
		DryRun:         false,
		Verbose:        verbose,
		GitHubRepo:     chartmodels.RepoOSSTenant,    // Default repository
		GitHubBranch:   chartmodels.DefaultGitBranch, // Default branch
		CertDir:        "",                           // Auto-detected
		NonInteractive: nonInteractive,
		KubeConfig:     kubeConfig,
		KubeContext:    k8s.ResolveContextForCluster(k8s.DefaultKubeconfigPath(), clusterName),
		// Inject cluster access from the orchestrator (composition root) so the
		// app subsystem stays isolated from cluster-creation code (req 18/19).
		ClusterAccess: cluster.NewClusterService(executor.NewRealCommandExecutor(false, verbose)),
	}
}
