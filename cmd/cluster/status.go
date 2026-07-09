package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func getStatusCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	statusCmd := &cobra.Command{
		Use:   "status [NAME]",
		Short: "Show detailed cluster status and information",
		Long: `Show detailed status information for a Kubernetes cluster.

Displays cluster health, node status, installed applications,
resource usage, and connectivity information.

Examples:
  openframe cluster status my-cluster
  openframe cluster status  # interactive selection
  openframe cluster status my-cluster --detailed
  openframe cluster status my-cluster -o json`,
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			if err := utils.ValidateGlobalFlags(); err != nil {
				return err
			}
			return models.ValidateStatusFlags(utils.GetGlobalFlags().Status)
		},
		RunE: utils.WrapCommandWithCommonSetup(runClusterStatus),
	}

	// Add status-specific flags
	models.AddStatusFlags(statusCmd, utils.GetGlobalFlags().Status)
	statusCmd.Flags().StringP("output", "o", "text", "Output format: text, json, or yaml")

	return statusCmd
}

func runClusterStatus(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	operationsUI := ui.NewOperationsUI()

	// Get all available clusters
	clusters, err := service.ListClusters()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	machine := output == "json" || output == "yaml"

	// Resolve the cluster name. Machine output must never drop into an
	// interactive picker, so it requires an explicit name.
	var clusterName string
	if machine {
		if len(args) == 0 {
			return fmt.Errorf("--output %s requires a cluster name", output)
		}
		clusterName = args[0]
	} else {
		clusterName, err = operationsUI.SelectClusterForOperation(clusters, args, "check status")
		if err != nil {
			return err
		}
		// If no cluster selected (e.g., empty list), exit gracefully
		if clusterName == "" {
			return nil
		}
	}

	switch output {
	case "json", "yaml":
		info, serr := service.GetClusterStatus(clusterName)
		if serr != nil {
			return fmt.Errorf("failed to get cluster status: %w", serr)
		}
		return printClusterStatus(info, output)
	case "", "text":
		globalFlags := utils.GetGlobalFlags()
		return service.ShowClusterStatus(clusterName, globalFlags.Status.Detailed, globalFlags.Status.NoApps, globalFlags.Global.Verbose)
	default:
		return fmt.Errorf("invalid --output %q (want \"text\", \"json\", or \"yaml\")", output)
	}
}

// printClusterStatus writes a single cluster's status as JSON or YAML. Both reuse
// the ClusterInfo `json:` tags, so field names match across formats.
func printClusterStatus(info models.ClusterInfo, format string) error {
	var (
		b   []byte
		err error
	)
	if format == "yaml" {
		b, err = yaml.Marshal(info)
	} else {
		b, err = json.MarshalIndent(info, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("encoding %s: %w", format, err)
	}
	fmt.Println(string(b))
	return nil
}
