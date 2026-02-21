package bootstrap

import (
	"github.com/flamingo-stack/openframe-cli/internal/bootstrap"
	"github.com/spf13/cobra"
)

// GetBootstrapCmd returns the bootstrap command
func GetBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap [cluster-name]",
		Short: "Bootstrap complete OpenFrame environment",
		Long: `Bootstrap Complete OpenFrame Environment

This command performs a complete OpenFrame setup by running:
1. Pre-flight check - Validates ALL prerequisites (tools, memory, certificates) upfront
2. openframe cluster create - Creates a Kubernetes cluster
3. openframe chart install - Installs ArgoCD and OpenFrame charts

This is equivalent to running both commands sequentially but provides
a streamlined experience for getting started with OpenFrame.

Examples:
  openframe bootstrap                                    # Interactive mode (default)
  openframe bootstrap my-cluster                        # Bootstrap with custom cluster name
  openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
  openframe bootstrap --deployment-mode=saas-shared --non-interactive  # Full CI/CD mode
  openframe bootstrap --verbose                         # Show detailed logs including ArgoCD sync progress
  openframe bootstrap -v --deployment-mode=oss-tenant  # Verbose mode with pre-selected deployment
  openframe bootstrap --repo=https://github.com/myorg/myrepo --branch=dev  # Custom repository`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return bootstrap.NewService().Execute(cmd, args)
		},
	}

	// Deployment configuration
	cmd.Flags().String("deployment-mode", "", "Deployment mode: oss-tenant, saas-tenant, saas-shared (skips deployment selection)")
	cmd.Flags().Bool("non-interactive", false, "Skip all prompts, use existing helm-values.yaml")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed logging including ArgoCD sync progress")
	cmd.Flags().Bool("force", false, "Continue even with insufficient memory or other warnings")

	// Repository overrides (useful for contributors working on forks)
	cmd.Flags().String("repo", "", "Override the default GitHub repository URL")
	cmd.Flags().String("branch", "", "Override the default Git branch (default: main)")

	return cmd
}
