package bootstrap

import (
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/bootstrap"
	clustermodels "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/spf13/cobra"
)

// GetBootstrapCmd returns the bootstrap command
func GetBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap [cluster-name]",
		Short: "Bootstrap complete OpenFrame environment",
		Long: `Bootstrap Complete OpenFrame Environment

This command performs a complete OpenFrame setup by running:
1. openframe cluster create - Creates a Kubernetes cluster
2. openframe chart install - Installs ArgoCD and OpenFrame charts

This is equivalent to running both commands sequentially but provides
a streamlined experience for getting started with OpenFrame.

Examples:
  openframe bootstrap                                    # Interactive mode (default)
  openframe bootstrap my-cluster                        # Bootstrap with custom cluster name
  openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
  openframe bootstrap --deployment-mode=saas-shared --non-interactive  # Full CI/CD mode
  openframe bootstrap --verbose                         # Show detailed logs including ArgoCD sync progress
  openframe bootstrap -v --deployment-mode=oss-tenant  # Verbose mode with pre-selected deployment`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate the cluster name at the boundary (RFC1123) so nothing
			// unsafe reaches the cluster-creation shell-outs downstream.
			if len(args) > 0 {
				if err := clustermodels.ValidateClusterName(strings.TrimSpace(args[0])); err != nil {
					verbose, _ := cmd.Flags().GetBool("verbose")
					return sharedErrors.HandleGlobalError(err, verbose)
				}
			}
			// Logo will be shown by cluster wrapper before prerequisites
			return bootstrap.NewService().Execute(cmd, args)
		},
	}

	// Add deployment mode flags
	cmd.Flags().String("deployment-mode", "", "Deployment mode: oss-tenant, saas-tenant, saas-shared (skips deployment selection)")
	cmd.Flags().Bool("non-interactive", false, "Skip all prompts, use existing helm-values.yaml")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed logging including ArgoCD sync progress")

	return cmd
}
