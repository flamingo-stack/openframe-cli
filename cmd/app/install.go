package app

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/app/target"
	chartmodels "github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
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
		Long: fmt.Sprintf(`Install ArgoCD and app-of-apps on a Kubernetes cluster

This command installs:
1. ArgoCD (version %s) with custom values
2. App-of-apps from GitHub repository (configurable)

The cluster must exist before running this command.
Certificates are automatically regenerated during installation.

Examples:
  openframe chart install                                    # Interactive mode (default)
  openframe chart install my-cluster                        # Install on specific cluster
  openframe chart install --non-interactive                 # Use existing helm-values.yaml (CI/CD)
  openframe chart install --github-branch develop          # Use develop branch
  openframe chart install --ref v1.2.3                     # Deploy a release tag`, argocd.ArgoCDChartVersion),
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

	req, err := buildInstallRequest(cmd, args, flags, verbose, "Installing")
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	if err := services.InstallChartsWithConfigContext(cmd.Context(), req); err != nil {
		// Use shared error handler for consistent error display
		return sharedErrors.HandleGlobalError(err, verbose)
	}
	return nil
}

// buildInstallRequest assembles the InstallationRequest and resolves the target
// cluster's rest.Config: an explicit --context, or — for a bare interactive run
// (no cluster name, not --non-interactive/--dry-run) — a prompt-selected
// context. A named cluster / --non-interactive / --dry-run leave KubeConfig nil
// so the service layer selects the cluster. `action` labels the interactive
// message (e.g. "Installing"/"Upgrading"). Shared by install and upgrade Mode 1.
func buildInstallRequest(cmd *cobra.Command, args []string, flags *InstallFlags, verbose bool, action string) (types.InstallationRequest, error) {
	req := types.InstallationRequest{
		Args:         args,
		Force:        flags.Force,
		DryRun:       flags.DryRun,
		Verbose:      verbose,
		GitHubRepo:   flags.GitHubRepo,
		GitHubBranch: flags.resolvedRef(),
		// An explicitly set ref must win over the branch baked into helm-values.yaml.
		GitHubRefExplicit: cmd.Flags().Changed("ref") || cmd.Flags().Changed("github-branch"),
		CertDir:           flags.CertDir,
		NonInteractive:    flags.NonInteractive,
		// Inject cluster access from the command layer (composition root) so the
		// app subsystem stays isolated from cluster-creation code (req 18/19).
		ClusterAccess: cluster.NewClusterService(executor.NewRealCommandExecutor(false, verbose)),
	}

	// Explicit --context targets a specific cluster directly (scriptable, skips
	// interactive selection). Its rest.Config is resolved here at the command layer.
	if contextName, _ := cmd.Flags().GetString("context"); contextName != "" {
		cfg, cerr := k8s.RestConfigForContext(k8s.DefaultKubeconfigPath(), contextName)
		if cerr != nil {
			return req, fmt.Errorf("could not use context %q: %w", contextName, cerr)
		}
		req.KubeConfig = cfg
	}

	// Bare interactive run (no cluster name): let the user pick a kube-context and
	// validate the cluster is reachable/ready before proceeding (req 16/27). Every
	// other invocation keeps its existing behavior: a named cluster,
	// --non-interactive, --dry-run, or --context all skip this.
	if req.KubeConfig == nil && !flags.NonInteractive && !flags.DryRun && len(args) == 0 {
		sel := target.NewSelector(target.UIPrompter{}, recommendedRequirements())
		res, serr := sel.Select(cmd.Context())
		if serr != nil {
			return req, serr
		}
		if !res.ResourcesSufficient {
			pterm.Warning.Printf("Cluster %q is smaller than recommended (~%d cores / %dGB RAM). Continuing anyway.\n",
				res.Context, recommendedCPUCores, recommendedMemGB)
		}
		pterm.Info.Printf("%s OpenFrame into context %q\n", action, res.Context)
		req.KubeConfig = res.Config
	}

	return req, nil
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
	Ref            string
	CertDir        string
	NonInteractive bool
}

// resolvedRef returns the git ref to deploy: the general --ref wins over the
// legacy --github-branch when both are set, otherwise --github-branch (whose
// default is "main") is used.
func (f *InstallFlags) resolvedRef() string {
	if f.Ref != "" {
		return f.Ref
	}
	return f.GitHubBranch
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

	if flags.Ref, err = cmd.Flags().GetString("ref"); err != nil {
		return nil, err
	}

	if flags.CertDir, err = cmd.Flags().GetString("cert-dir"); err != nil {
		return nil, err
	}

	if flags.NonInteractive, err = cmd.Flags().GetBool("non-interactive"); err != nil {
		return nil, err
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
	cmd.Flags().String("github-repo", chartmodels.RepoOSSTenant, "GitHub repository URL")
	cmd.Flags().String("github-branch", chartmodels.DefaultGitBranch, "Git ref (branch or tag) to deploy")
	cmd.Flags().StringP("ref", "r", "", "Git ref (branch or release tag, e.g. v1.2.3) to deploy; supersedes --github-branch")
	cmd.Flags().String("cert-dir", "", "Certificate directory (auto-detected if not provided)")
	cmd.Flags().Bool("non-interactive", false, "Skip all prompts, use existing helm-values.yaml")
	cmd.Flags().StringP("context", "c", "", "Kube-context to install into (skips interactive selection)")
}
