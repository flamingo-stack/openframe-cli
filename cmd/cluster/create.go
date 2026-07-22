package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/discovery"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/provider"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func getCreateCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	createCmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a new Kubernetes cluster",
		Long: `Create a new Kubernetes cluster with quick defaults or interactive configuration.

By default, shows a selection menu where you can choose:
1. Quick start with defaults (press Enter) - creates cluster with default settings
2. Interactive configuration wizard - step-by-step cluster customization

Creates a local k3d cluster or a cloud EKS cluster for OpenFrame. If a cluster
with the same name already exists it is left untouched and reused — delete it
first to start from scratch. Use the bootstrap command to install OpenFrame
components after creation.

EKS clusters are provisioned with Terraform (installed automatically) and
create AWS resources that incur costs; the workspace and state live under
~/.openframe/clusters/<name>. A failed create can be re-run to resume, or
torn down with 'openframe cluster delete'.

Examples:
  openframe cluster create                    # Show creation mode selection
  openframe cluster create my-cluster        # Show selection with custom name
  openframe cluster create --skip-wizard     # Direct creation with defaults
  openframe cluster create --nodes 3 --type k3d --skip-wizard
  openframe cluster create my-eks --type eks --region us-east-1 --skip-wizard`,
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			if err := utils.ValidateGlobalFlags(); err != nil {
				return err
			}
			globalFlags := utils.GetGlobalFlags()
			if globalFlags != nil && globalFlags.Create != nil {
				return models.ValidateCreateFlags(globalFlags.Create)
			}
			return nil
		},
		RunE: utils.WrapCommandWithCommonSetup(runCreateCluster),
	}

	// Add create-specific flags
	globalFlags := utils.GetGlobalFlags()
	if globalFlags != nil && globalFlags.Create != nil {
		models.AddCreateFlags(createCmd, globalFlags.Create)
	}

	return createCmd
}

// planPreviewFn is the cloud dry-run implementation. A package variable so
// cmd-layer tests can stub it out: the real one shells out to terraform,
// which unit tests must never do.
var planPreviewFn = cloudPlanPreview

// cloudPlanPreview runs a real terraform plan for a cloud config and prints
// the resource footprint. Terraform not being installed is a soft skip — the
// prerequisite gate only runs on a real create, and a dry-run must not
// install anything.
func cloudPlanPreview(ctx context.Context, config models.ClusterConfig) error {
	if _, err := terraform.FindTerraform(); err != nil {
		pterm.Info.Println("terraform is not installed — skipping the plan preview (it installs automatically on a real create)")
		return nil
	}

	exec := executor.NewRealCommandExecutor(false, utils.GetGlobalFlags().Global.Verbose)
	p, err := provider.New(config.Type, exec)
	if err != nil {
		return err
	}
	planner, ok := p.(provider.Planner)
	if !ok {
		return nil
	}

	if config.Type == models.ClusterTypeGKE {
		if err := discovery.NewAuthFlow(exec).Ensure(ctx, true); err != nil {
			return err
		}
	}

	pterm.Info.Printf("Computing terraform plan for %s cluster '%s'...\n", config.Type, config.Name)
	summary, err := planner.PlanCluster(ctx, config)
	if err != nil {
		return err
	}
	if !summary.HasChanges() {
		pterm.Success.Println("Plan: no changes — the cluster already matches this configuration")
		return nil
	}
	for _, change := range summary.Changes {
		pterm.DefaultBasicText.Printf("  %-3s %s\n", change.Action, change.Address)
	}
	pterm.Success.Printf("Plan: %d to add, %d to change, %d to destroy\n", summary.Add, summary.Change, summary.Destroy)
	return nil
}

// showEKSComingSoonBanner is the temporary stub for AWS EKS creation.
func showEKSComingSoonBanner() {
	pterm.DefaultBox.
		WithTitle(" 🚧 AWS EKS — coming soon ").
		WithTitleTopCenter().
		Println("Creating AWS EKS clusters will be available shortly.\n" +
			"GKE is fully supported today:\n" +
			"  openframe cluster create my-gke --type gke --project <project> --region <region>")
}

func runCreateCluster(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	globalFlags := utils.GetGlobalFlags()

	var config models.ClusterConfig

	// Check if we should use interactive mode
	if !globalFlags.Create.SkipWizard {
		// Use UI layer to handle cluster configuration
		configHandler := ui.NewConfigurationHandler()

		// Get cluster name from args if provided
		var clusterName string
		if len(args) > 0 {
			clusterName = strings.TrimSpace(args[0])
			if err := models.ValidateClusterName(clusterName); err != nil {
				return err
			}
		}

		// Let UI handle the entire configuration flow
		var err error
		config, err = configHandler.GetClusterConfig(clusterName)
		if err != nil {
			return err
		}
	} else {
		// Non-interactive mode - build config from flags and args
		clusterName := ""
		if len(args) > 0 {
			clusterName = strings.TrimSpace(args[0])
			// Validate the cluster name
			if err := models.ValidateClusterName(clusterName); err != nil {
				return err
			}
		} else {
			clusterName = "openframe-dev" // default name
		}

		// Handle node count validation - error if user explicitly set 0 or negative
		nodeCount := globalFlags.Create.NodeCount
		if cmd.Flags().Changed("nodes") && nodeCount <= 0 {
			return fmt.Errorf("node count must be at least 1: %d", nodeCount)
		}
		// Auto-correct to default if not explicitly set and invalid
		if nodeCount <= 0 {
			nodeCount = 3
		}

		config = models.ClusterConfig{
			Name:       clusterName,
			Type:       models.ClusterType(globalFlags.Create.ClusterType),
			K8sVersion: globalFlags.Create.K8sVersion,
			NodeCount:  nodeCount,
		}

		// Set defaults if needed
		if config.Type == "" {
			config.Type = models.ClusterTypeK3d
		}

		// Cloud settings only exist for cloud types; the k3d backend rejects a
		// non-nil Cloud by design.
		if config.Type == models.ClusterTypeEKS || config.Type == models.ClusterTypeGKE {
			cf := globalFlags.Create
			config.Cloud = &models.CloudConfig{
				Region:        cf.Region,
				Profile:       cf.Profile,
				Project:       cf.Project,
				MachineType:   cf.MachineType,
				MinNodes:      cf.MinNodes,
				MaxNodes:      cf.MaxNodes,
				Spot:          cf.Spot,
				BackendConfig: cf.BackendConfig,
			}
		}
	}

	// AWS EKS creation is temporarily gated behind a coming-soon banner while
	// the GKE flow is being finished end-to-end. The EKS provider stays fully
	// functional for existing clusters (status/delete/resume) — only NEW
	// creates are gated.
	if config.Type == models.ClusterTypeEKS {
		showEKSComingSoonBanner()
		return nil
	}

	// Show configuration summary for dry-run or skip-wizard modes
	if globalFlags.Create.DryRun || globalFlags.Create.SkipWizard || globalFlags.Global.Verbose {
		operationsUI := ui.NewOperationsUI()
		operationsUI.ShowConfigurationSummary(config, globalFlags.Create.DryRun, globalFlags.Create.SkipWizard)

		// If dry-run, don't actually create the cluster. For cloud types the
		// dry-run is a real terraform plan of what create would provision.
		if globalFlags.Create.DryRun {
			if config.Type == models.ClusterTypeEKS || config.Type == models.ClusterTypeGKE {
				return planPreviewFn(cmd.Context(), config)
			}
			return nil
		}
	}

	// Type-aware prerequisite gate: runs after the type is known (wizard or
	// flags), so only the tools the chosen backend needs are demanded. It sits
	// after the dry-run return on purpose — the gate may INSTALL tools, and
	// dry-run must not mutate the system.
	if err := prerequisites.CheckForClusterType(config.Type); err != nil {
		return err
	}

	// Single auth flow: for GKE, offer `gcloud auth login` (+ ADC for
	// terraform) right here instead of failing later in the provider
	// preflight with a "run this command" error.
	if config.Type == models.ClusterTypeGKE {
		if err := discovery.NewAuthFlow(utils.CommandExecutor()).Ensure(cmd.Context(), true); err != nil {
			return err
		}
	}

	// Execute cluster creation through service layer
	// We ignore the returned rest.Config as it's not needed for standalone cluster creation
	_, err := service.CreateCluster(cmd.Context(), config)
	return err
}
