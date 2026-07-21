package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/discovery"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/ui"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func getUseCmd() *cobra.Command {
	// Ensure global flags are initialized
	utils.InitGlobalFlags()

	useCmd := &cobra.Command{
		Use:   "use [NAME]",
		Short: "Switch the kubectl context to a cluster",
		Long: `Switch the current kubectl context (and, for GKE, the active gcloud
configuration) to the named cluster.

Works for every cluster the CLI can see: local k3d clusters, clusters created
by openframe, and external GKE clusters discovered in your gcloud projects.
For an external cluster without a kubeconfig entry, credentials are fetched
via 'gcloud container clusters get-credentials' first.

Only local configuration changes: the cluster itself is never touched.

Examples:
  openframe cluster use openframe-dev     # local k3d
  openframe cluster use my-gke            # openframe-managed GKE
  openframe cluster use tenant-cluster-1  # external GKE (discovered)
  openframe cluster use                   # interactive selection`,
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SyncGlobalFlags()
			return utils.ValidateGlobalFlags()
		},
		RunE: utils.WrapCommandWithCommonSetup(runUseCluster),
	}

	return useCmd
}

func runUseCluster(cmd *cobra.Command, args []string) error {
	service := utils.GetCommandService()
	exec := utils.CommandExecutor()
	ctx := cmd.Context()

	// Resolve the cluster name: explicit arg, or interactive selection from
	// the clusters the CLI already knows (local + managed — fast, no cloud
	// calls; external clusters are addressed by explicit name).
	name := ""
	if len(args) > 0 {
		name = strings.TrimSpace(args[0])
	} else {
		clusters, err := service.ListClusters()
		if err != nil {
			return fmt.Errorf("failed to list clusters: %w", err)
		}
		name, err = ui.SelectClusterByName(clusters, "Select a cluster to use")
		if err != nil {
			return err
		}
		if name == "" {
			return nil
		}
	}

	kubeconfig := k8s.DefaultKubeconfigPath()

	// 1) Clusters the CLI knows: managed cloud (registry) or local k3d.
	if clusterType, err := service.DetectClusterType(name); err == nil {
		contextName := name
		if clusterType == models.ClusterTypeK3d {
			contextName = k8s.ResolveContextForCluster(kubeconfig, name)
		}
		if clusterType == models.ClusterTypeGKE {
			if info, err := service.GetClusterStatus(name); err == nil {
				alignGcloudConfiguration(ctx, exec, info.Project)
			}
		}
		return switchTo(kubeconfig, contextName, name)
	}

	// 2) External GKE clusters, addressed by name via discovery.
	return useExternalGKE(ctx, exec, kubeconfig, name)
}

// useExternalGKE finds an external cluster by name and points kubectl at it,
// fetching credentials when the kubeconfig has no entry yet.
func useExternalGKE(ctx context.Context, exec executor.CommandExecutor, kubeconfig, name string) error {
	d := discovery.NewGKEDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return fmt.Errorf("cluster '%s' is not known locally, and gcloud is not installed to look for it in GCP", name)
	case discovery.NotAuthenticated:
		// One unambiguous flow: offer the login right here (interactive only).
		pterm.Info.Printf("Cluster '%s' is not known locally — looking for it in your GCP projects requires a Google Cloud login\n", name)
		if err := discovery.NewAuthFlow(exec).Ensure(ctx, false); err != nil {
			return err
		}
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return err
	}
	var found *models.ClusterInfo
	for i := range result.Clusters {
		if result.Clusters[i].Name == name {
			found = &result.Clusters[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("cluster '%s' not found locally or in the %s", name, "GCP projects of your gcloud configurations")
	}

	alignGcloudConfiguration(ctx, exec, found.Project)

	contextName := found.Context
	if contextName == "" {
		// No kubeconfig entry yet — fetch credentials (adds the gke_* context).
		pterm.Info.Printf("Fetching credentials for '%s' (project %s, %s)...\n", name, found.Project, found.Region)
		if _, err := exec.Execute(ctx, "gcloud", "container", "clusters", "get-credentials", name,
			"--project", found.Project, "--region", found.Region); err != nil {
			return fmt.Errorf("could not fetch credentials for '%s' (for private clusters try 'gcloud container fleet memberships get-credentials %s'): %w", name, name, err)
		}
		contextName = fmt.Sprintf("gke_%s_%s_%s", found.Project, found.Region, name)
	}
	return switchTo(kubeconfig, contextName, name)
}

// switchTo flips current-context and reports the result.
func switchTo(kubeconfig, contextName, clusterName string) error {
	if !k8s.HasContext(kubeconfig, contextName) {
		return fmt.Errorf("cluster '%s' has no kubeconfig context '%s' — fetch credentials for it first", clusterName, contextName)
	}
	if err := k8s.SwitchContext(kubeconfig, contextName); err != nil {
		return err
	}
	pterm.Success.Printf("Switched kubectl context to '%s' (cluster '%s')\n", contextName, clusterName)
	return nil
}

// alignGcloudConfiguration activates the gcloud configuration whose project
// matches the cluster, so gcloud commands line up with kubectl. Best-effort:
// no matching configuration is fine, and failures never block the switch.
func alignGcloudConfiguration(ctx context.Context, exec executor.CommandExecutor, project string) {
	if project == "" {
		return
	}
	d := discovery.NewGKEDiscoverer(exec)
	configName, err := d.ConfigurationForProject(ctx, project)
	if err != nil || configName == "" {
		return
	}
	if _, err := exec.Execute(ctx, "gcloud", "config", "configurations", "activate", configName); err == nil {
		pterm.Info.Printf("Activated gcloud configuration '%s' (project %s)\n", configName, project)
	}
}
