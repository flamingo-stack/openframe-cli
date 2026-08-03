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
by openframe, and external GKE/EKS clusters discovered in your gcloud projects
and AWS profiles. For an external cluster without a kubeconfig entry,
credentials are fetched via 'gcloud container clusters get-credentials' /
'aws eks update-kubeconfig' first.

Only local configuration changes: the cluster itself is never touched.

Examples:
  openframe cluster use openframe-dev     # local k3d
  openframe cluster use my-gke            # openframe-managed GKE
  openframe cluster use tenant-cluster-1  # external GKE/EKS (discovered)
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

	// 2) External cloud clusters, addressed by name via discovery: GKE first
	// (the historical behavior), then EKS.
	return useExternalCluster(ctx, exec, kubeconfig, name)
}

// useExternalCluster looks for the named cluster among the external clusters
// of each cloud the operator's CLIs can reach. A cloud whose CLI is missing
// (or whose credentials are unusable, for AWS) is skipped and named in the
// final error; a cloud that IS searchable but errors mid-flow fails hard —
// silently skipping it would misreport an existing cluster as absent.
func useExternalCluster(ctx context.Context, exec executor.CommandExecutor, kubeconfig, name string) error {
	var searched []string

	handled, gkeSearched, err := useExternalGKE(ctx, exec, kubeconfig, name)
	if handled {
		return err
	}
	searched = append(searched, gkeSearched...)

	handled, eksSearched, err := useExternalEKS(ctx, exec, kubeconfig, name)
	if handled {
		return err
	}
	searched = append(searched, eksSearched...)

	if len(searched) == 0 {
		return fmt.Errorf("cluster '%s' is not known locally, and neither gcloud nor the AWS CLI is available to look for it in the cloud", name)
	}
	return fmt.Errorf("cluster '%s' not found locally or in %s", name, strings.Join(searched, ", or "))
}

// useExternalGKE finds an external GKE cluster by name and points kubectl at
// it, fetching credentials when the kubeconfig has no entry yet. handled=false
// means GKE could not be searched (gcloud missing) or the cluster is not
// there; searched names what was looked through for the caller's final error.
func useExternalGKE(ctx context.Context, exec executor.CommandExecutor, kubeconfig, name string) (handled bool, searched []string, err error) {
	d := discovery.NewGKEDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return false, nil, nil
	case discovery.NotAuthenticated:
		// One unambiguous flow: offer the login right here (interactive only).
		pterm.Info.Printf("Cluster '%s' is not known locally — looking for it in your GCP projects requires a Google Cloud login\n", name)
		if err := discovery.NewAuthFlow(exec).Ensure(ctx, false); err != nil {
			return true, nil, err
		}
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return true, nil, err
	}
	searched = []string{"the GCP projects of your gcloud configurations"}
	var found *models.ClusterInfo
	for i := range result.Clusters {
		if result.Clusters[i].Name == name {
			found = &result.Clusters[i]
			break
		}
	}
	if found == nil {
		return false, searched, nil
	}

	alignGcloudConfiguration(ctx, exec, found.Project)

	contextName := found.Context
	if contextName == "" {
		// No kubeconfig entry yet — fetch credentials (adds the gke_* context).
		pterm.Info.Printf("Fetching credentials for '%s' (project %s, %s)...\n", name, found.Project, found.Region)
		// --location (not --region): discovery's Region field carries the GKE
		// location, which is a ZONE for zonal clusters.
		if _, err := exec.Execute(ctx, "gcloud", "container", "clusters", "get-credentials", name,
			"--project", found.Project, "--location", found.Region); err != nil {
			return true, searched, fmt.Errorf("could not fetch credentials for '%s' (for private clusters try 'gcloud container fleet memberships get-credentials %s'): %w", name, name, err)
		}
		contextName = fmt.Sprintf("gke_%s_%s_%s", found.Project, found.Region, name)
	}
	return true, searched, switchTo(kubeconfig, contextName, name)
}

// useExternalEKS is the AWS twin of useExternalGKE. There is no interactive
// AWS login to offer (credentials come from profiles/SSO/env, not a browser
// flow the CLI can drive), so unusable credentials make AWS unsearchable
// rather than prompting.
func useExternalEKS(ctx context.Context, exec executor.CommandExecutor, kubeconfig, name string) (handled bool, searched []string, err error) {
	d := discovery.NewEKSDiscoverer(exec)
	switch d.AuthStatus(ctx) {
	case discovery.CLIMissing:
		return false, nil, nil
	case discovery.NotAuthenticated:
		return false, []string{"your AWS profiles (skipped: credentials not usable — run 'aws configure' or 'aws sso login')"}, nil
	}

	result, err := d.Discover(ctx)
	if err != nil {
		return true, nil, err
	}
	searched = []string{"the default regions of your AWS profiles"}
	var found *models.ClusterInfo
	for i := range result.Clusters {
		if result.Clusters[i].Name == name {
			found = &result.Clusters[i]
			break
		}
	}
	if found == nil {
		return false, searched, nil
	}

	contextName := found.Context
	if contextName == "" {
		// No kubeconfig entry yet — fetch credentials. --alias names the new
		// context after the cluster (instead of its ARN) so the rest of the CLI
		// resolves it by exact match.
		pterm.Info.Printf("Fetching credentials for '%s' (%s)...\n", name, found.Region)
		args := []string{"eks", "update-kubeconfig", "--name", name, "--region", found.Region, "--alias", name}
		if found.Profile != "" {
			args = append(args, "--profile", found.Profile)
		}
		if _, err := exec.Execute(ctx, "aws", args...); err != nil {
			return true, searched, fmt.Errorf("could not fetch credentials for '%s': %w", name, err)
		}
		contextName = name
	}
	return true, searched, switchTo(kubeconfig, contextName, name)
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
