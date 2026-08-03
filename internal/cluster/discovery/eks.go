package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// EKSDiscoverer lists EKS clusters across the user's configured AWS profiles,
// each in that profile's default region. Like the GKE twin it never stores
// credentials: everything is delegated to the aws CLI through the shared
// CommandExecutor, so it is fully mockable and respects whatever the user has
// authenticated.
type EKSDiscoverer struct {
	exec executor.CommandExecutor
}

func NewEKSDiscoverer(exec executor.CommandExecutor) *EKSDiscoverer {
	return &EKSDiscoverer{exec: exec}
}

// AuthStatus probes the aws CLI without failing: a missing binary and unusable
// credentials are states to report, not errors.
func (d *EKSDiscoverer) AuthStatus(ctx context.Context) AuthStatus {
	result, err := d.exec.Execute(ctx, "aws", "sts", "get-caller-identity", "--output", "json")
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return CLIMissing
		}
		return NotAuthenticated
	}
	if result == nil || strings.TrimSpace(result.Stdout) == "" {
		return NotAuthenticated
	}
	return Authenticated
}

// Profiles returns the named profiles of the user's AWS config (aws configure
// list-profiles), sorted. An empty list simply means only the default
// credential chain (env vars / instance role) is configured.
func (d *EKSDiscoverer) Profiles(ctx context.Context) ([]string, error) {
	result, err := d.exec.Execute(ctx, "aws", "configure", "list-profiles")
	if err != nil {
		return nil, fmt.Errorf("listing AWS profiles: %w", err)
	}
	if result == nil {
		return nil, nil
	}
	seen := map[string]bool{}
	var profiles []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		p := strings.TrimSpace(line)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)
	return profiles, nil
}

// Regions returns the AWS regions enabled for the account (ec2
// describe-regions), sorted. Used to populate the create wizard's region
// picker so an EKS user selects a valid region instead of typing one. An empty
// profile means the default credential chain.
func (d *EKSDiscoverer) Regions(ctx context.Context, profile string) ([]string, error) {
	args := []string{"ec2", "describe-regions", "--query", "Regions[].RegionName", "--output", "text"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	result, err := d.exec.Execute(ctx, "aws", args...)
	if err != nil {
		return nil, fmt.Errorf("listing AWS regions: %w", err)
	}
	if result == nil {
		return nil, nil
	}
	seen := map[string]bool{}
	var regions []string
	for _, field := range strings.Fields(result.Stdout) {
		if field == "None" || seen[field] {
			continue
		}
		seen[field] = true
		regions = append(regions, field)
	}
	sort.Strings(regions)
	return regions, nil
}

// profileRegion returns a profile's configured default region ("" when none is
// set — such a profile cannot be discovered against and is reported, not
// guessed at).
func (d *EKSDiscoverer) profileRegion(ctx context.Context, profile string) string {
	args := []string{"configure", "get", "region"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	result, err := d.exec.Execute(ctx, "aws", args...)
	if err != nil || result == nil {
		return ""
	}
	return strings.TrimSpace(result.Stdout)
}

// eksListClusters is the shape of `aws eks list-clusters --output json`.
type eksListClusters struct {
	Clusters []string `json:"clusters"`
}

// eksCluster is the subset of `aws eks describe-cluster --output json` the
// discoverer reads.
type eksCluster struct {
	Cluster struct {
		Name    string `json:"name"`
		Arn     string `json:"arn"`
		Status  string `json:"status"`
		Version string `json:"version"`
	} `json:"cluster"`
}

// Discover lists EKS clusters in each configured profile's default region,
// best-effort per profile (missing region, expired credentials, or a denied
// API call must not hide the other profiles). Discovered entries carry
// Source=external and, when resolvable, the kubeconfig context that reaches
// them. Profiles pointing at the same account/region yield the same clusters;
// they are de-duplicated by ARN.
func (d *EKSDiscoverer) Discover(ctx context.Context) (Result, error) {
	profiles, err := d.Profiles(ctx)
	if err != nil {
		return Result{}, err
	}
	if len(profiles) == 0 {
		// No named profiles: discover through the default credential chain.
		profiles = []string{""}
	}
	contexts := kubeconfigContexts()

	var res Result
	seen := map[string]bool{}
	for _, profile := range profiles {
		label := profile
		if label == "" {
			label = "default credentials"
		}
		region := d.profileRegion(ctx, profile)
		if region == "" {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: no default region configured", label))
			continue
		}
		args := []string{"eks", "list-clusters", "--region", region, "--output", "json"}
		if profile != "" {
			args = append(args, "--profile", profile)
		}
		result, err := d.exec.Execute(ctx, "aws", args...)
		if err != nil {
			// Typical: expired SSO session or a role without eks:ListClusters.
			// Report and keep going.
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: %v", label, err))
			continue
		}
		var list eksListClusters
		if err := json.Unmarshal([]byte(result.Stdout), &list); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: unparseable clusters list", label))
			continue
		}
		for _, name := range list.Clusters {
			info, arn, err := d.describe(ctx, profile, region, name)
			if err != nil {
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s: %v", label, err))
				continue
			}
			if arn != "" && seen[arn] {
				continue // same cluster through another profile
			}
			seen[arn] = true
			info.Context = matchEKSContext(contexts, arn, name)
			res.Clusters = append(res.Clusters, info)
		}
	}
	return res, nil
}

// describe fetches one cluster's details and maps them onto the shared
// ClusterInfo shape. EKS does not report a node count from the control plane —
// nodes live in separate node groups — so NodeCount stays 0 for external
// clusters.
func (d *EKSDiscoverer) describe(ctx context.Context, profile, region, name string) (models.ClusterInfo, string, error) {
	args := []string{"eks", "describe-cluster", "--name", name, "--region", region, "--output", "json"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	result, err := d.exec.Execute(ctx, "aws", args...)
	if err != nil {
		return models.ClusterInfo{}, "", fmt.Errorf("describing cluster %s: %w", name, err)
	}
	var c eksCluster
	if err := json.Unmarshal([]byte(result.Stdout), &c); err != nil {
		return models.ClusterInfo{}, "", fmt.Errorf("unparseable describe-cluster for %s", name)
	}
	return models.ClusterInfo{
		Name:       c.Cluster.Name,
		Type:       models.ClusterTypeEKS,
		Source:     models.SourceExternal,
		Status:     titleCase(c.Cluster.Status),
		K8sVersion: c.Cluster.Version,
		Profile:    profile,
		Region:     region,
		Context:    "",
	}, c.Cluster.Arn, nil
}

// matchEKSContext maps a discovered cluster onto a kubeconfig context by the
// two conventional naming shapes; a renamed context cannot be matched and
// yields "".
func matchEKSContext(contexts []string, arn, name string) string {
	candidates := []string{
		name, // plain (openframe-style or hand-renamed to the cluster name)
		arn,  // aws eks update-kubeconfig default
	}
	for _, want := range candidates {
		if want == "" {
			continue
		}
		for _, have := range contexts {
			if have == want {
				return have
			}
		}
	}
	return ""
}
