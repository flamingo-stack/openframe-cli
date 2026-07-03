package app

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/app/target"
	"github.com/flamingo-stack/openframe-cli/internal/chart/services"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// getInstallCmd returns the install subcommand
func getInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [cluster-name]",
		Short: "Install ArgoCD and app-of-apps",
		Long: `Install ArgoCD and app-of-apps on a Kubernetes cluster

This command installs:
1. ArgoCD (version 10.1.0) with custom values
2. App-of-apps from GitHub repository (configurable)

The cluster must exist before running this command.
Certificates are automatically regenerated during installation.

Examples:
  openframe chart install                                    # Interactive mode (default)
  openframe chart install my-cluster                        # Install on specific cluster
  openframe chart install --deployment-mode=oss-tenant     # Skip deployment selection
  openframe chart install --deployment-mode=saas-shared --non-interactive  # Full CI/CD mode
  openframe chart install --github-branch develop          # Use develop branch`,
		RunE:          runInstallCommand,
		SilenceErrors: true, // Errors are handled by our custom error handler
		SilenceUsage:  true, // Don't show usage on errors
	}

	// Add flags directly
	addInstallFlags(cmd)

	return cmd
}

// runInstallCommand handles the install command execution
func runInstallCommand(cmd *cobra.Command, args []string) error {
	// Logo is already shown in PersistentPreRunE

	// Extract flags directly
	flags, err := extractInstallFlags(cmd)
	if err != nil {
		return err
	}

	// Get verbose flag (with fallback)
	verbose := getVerboseFlag(cmd)

	// Use common installation function
	req := types.InstallationRequest{
		Args:           args,
		Force:          flags.Force,
		DryRun:         flags.DryRun,
		Verbose:        verbose,
		GitHubRepo:     flags.GitHubRepo,
		GitHubBranch:   flags.GitHubBranch,
		CertDir:        flags.CertDir,
		DeploymentMode: flags.DeploymentMode,
		NonInteractive: flags.NonInteractive,
		// Inject cluster access from the command layer (composition root) so the
		// app subsystem stays isolated from cluster-creation code (req 18/19).
		ClusterAccess: cluster.NewClusterService(executor.NewRealCommandExecutor(false, verbose)),
	}

	// Explicit --context targets a specific cluster directly (scriptable, skips
	// interactive selection). Its rest.Config is resolved here at the command layer.
	if contextName, _ := cmd.Flags().GetString("context"); contextName != "" {
		cfg, cerr := k8s.RestConfigForContext(k8s.DefaultKubeconfigPath(), contextName)
		if cerr != nil {
			return sharedErrors.HandleGlobalError(fmt.Errorf("could not use context %q: %w", contextName, cerr), verbose)
		}
		req.KubeConfig = cfg
	}

	// Bare interactive install (`openframe app install`, no cluster name): let the
	// user pick a kube-context and validate the cluster is reachable/ready before
	// installing (req 16/27). Every other invocation keeps its existing behavior:
	// a named cluster, --non-interactive, --dry-run, or --context all skip this.
	if req.KubeConfig == nil && !flags.NonInteractive && !flags.DryRun && len(args) == 0 {
		sel := target.NewSelector(target.UIPrompter{}, recommendedRequirements())
		res, serr := sel.Select(cmd.Context())
		if serr != nil {
			return sharedErrors.HandleGlobalError(serr, verbose)
		}
		if !res.ResourcesSufficient {
			pterm.Warning.Printf("Cluster %q is smaller than recommended (~%d cores / %dGB RAM). Continuing anyway.\n",
				res.Context, recommendedCPUCores, recommendedMemGB)
		}
		pterm.Info.Printf("Installing OpenFrame into context %q\n", res.Context)
		req.KubeConfig = res.Config
	}

	err = services.InstallChartsWithConfigContext(cmd.Context(), req)
	if err != nil {
		// Use shared error handler for consistent error display
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

const (
	recommendedCPUCores = 6
	recommendedMemGB    = 24
)

// recommendedRequirements is the advisory minimum cluster capacity for OpenFrame
// (per the README system requirements). Falling short only warns; it never blocks.
func recommendedRequirements() k8s.Requirements {
	return k8s.Requirements{
		CPUMillis: recommendedCPUCores * 1000,
		MemBytes:  int64(recommendedMemGB) << 30,
	}
}

// InstallFlags contains all flags needed for chart installation
type InstallFlags struct {
	Force          bool
	DryRun         bool
	GitHubRepo     string
	GitHubBranch   string
	CertDir        string
	DeploymentMode string
	NonInteractive bool
}

// extractInstallFlags extracts install flags from cobra command
func extractInstallFlags(cmd *cobra.Command) (*InstallFlags, error) {
	flags := &InstallFlags{}
	var err error

	if flags.Force, err = cmd.Flags().GetBool("force"); err != nil {
		return nil, err
	}

	if flags.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return nil, err
	}

	if flags.GitHubRepo, err = cmd.Flags().GetString("github-repo"); err != nil {
		return nil, err
	}

	if flags.GitHubBranch, err = cmd.Flags().GetString("github-branch"); err != nil {
		return nil, err
	}

	if flags.CertDir, err = cmd.Flags().GetString("cert-dir"); err != nil {
		return nil, err
	}

	if flags.DeploymentMode, err = cmd.Flags().GetString("deployment-mode"); err != nil {
		return nil, err
	}

	if flags.NonInteractive, err = cmd.Flags().GetBool("non-interactive"); err != nil {
		return nil, err
	}

	// Validate deployment mode
	if err := types.ValidateDeploymentMode(flags.DeploymentMode); err != nil {
		return nil, err
	}

	// Validate non-interactive requires deployment mode
	if flags.NonInteractive && flags.DeploymentMode == "" {
		return nil, fmt.Errorf("--deployment-mode is required when using --non-interactive")
	}

	return flags, nil
}

// getVerboseFlag extracts verbose flag with fallback
func getVerboseFlag(cmd *cobra.Command) bool {
	// Try root command first
	if cmd.Root() != nil {
		if verbose, err := cmd.Root().PersistentFlags().GetBool("verbose"); err == nil {
			return verbose
		}
	}

	// Try current command
	if verbose, err := cmd.Flags().GetBool("verbose"); err == nil {
		return verbose
	}

	// Default to false
	return false
}

// addInstallFlags adds all install flags to the command
func addInstallFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("force", "f", false, "Force installation even if charts already exist")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without executing")
	cmd.Flags().String("github-repo", "https://github.com/flamingo-stack/openframe-oss-tenant", "GitHub repository URL")
	cmd.Flags().String("github-branch", "main", "GitHub repository branch")
	cmd.Flags().String("cert-dir", "", "Certificate directory (auto-detected if not provided)")
	cmd.Flags().String("deployment-mode", "", "Deployment mode: oss-tenant, saas-tenant, saas-shared (skips deployment selection)")
	cmd.Flags().Bool("non-interactive", false, "Skip all prompts, use existing helm-values.yaml")
	cmd.Flags().String("context", "", "Kube-context to install into (skips interactive selection)")
}
