// Package discovery finds cloud Kubernetes clusters that exist OUTSIDE the
// openframe workspace registry — clusters created by Terraform repos, other
// tools, or other people. Discovered clusters are strictly read-only for the
// CLI: they appear in `cluster list --all` and `cluster status`, and every
// mutating command refuses them (no workspace = not ours).
//
// Discovery never stores credentials: it delegates to the provider CLIs
// (gcloud) through the shared CommandExecutor, so it is fully mockable and
// respects whatever the user has authenticated.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"k8s.io/client-go/tools/clientcmd"
)

// AuthStatus reports whether a provider's CLI is usable without treating
// "not logged in" as an error — list must degrade gracefully.
type AuthStatus int

const (
	Authenticated AuthStatus = iota
	NotAuthenticated
	CLIMissing
)

// GKEDiscoverer lists GKE clusters across the projects configured in the
// user's named gcloud configurations (e.g. dev-*, stage-*, prod-*).
type GKEDiscoverer struct {
	exec executor.CommandExecutor
}

func NewGKEDiscoverer(exec executor.CommandExecutor) *GKEDiscoverer {
	return &GKEDiscoverer{exec: exec}
}

// AuthStatus probes gcloud without failing: a missing binary and a logged-out
// state are states to report, not errors.
func (d *GKEDiscoverer) AuthStatus(ctx context.Context) AuthStatus {
	result, err := d.exec.Execute(ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
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

// gcloudConfiguration is the subset of `gcloud config configurations list
// --format=json` the discoverer reads.
type gcloudConfiguration struct {
	Name       string `json:"name"`
	Properties struct {
		Core struct {
			Project string `json:"project"`
		} `json:"core"`
	} `json:"properties"`
}

// Projects returns the unique GCP projects named by the user's gcloud
// configurations, sorted. Configurations without a project (or with the
// literal "none") are skipped.
func (d *GKEDiscoverer) Projects(ctx context.Context) ([]string, error) {
	result, err := d.exec.Execute(ctx, "gcloud", "config", "configurations", "list", "--format=json")
	if err != nil {
		return nil, fmt.Errorf("listing gcloud configurations: %w", err)
	}
	var configs []gcloudConfiguration
	if err := json.Unmarshal([]byte(result.Stdout), &configs); err != nil {
		return nil, fmt.Errorf("parsing gcloud configurations: %w", err)
	}
	seen := map[string]bool{}
	var projects []string
	for _, c := range configs {
		p := strings.TrimSpace(c.Properties.Core.Project)
		if p == "" || p == "none" || seen[p] {
			continue
		}
		seen[p] = true
		projects = append(projects, p)
	}
	sort.Strings(projects)
	return projects, nil
}

// gkeCluster is the subset of `gcloud container clusters list --format=json`
// the discoverer reads.
type gkeCluster struct {
	Name                 string `json:"name"`
	Location             string `json:"location"`
	Status               string `json:"status"`
	CurrentNodeCount     int    `json:"currentNodeCount"`
	CurrentMasterVersion string `json:"currentMasterVersion"`
}

// Result is one discovery pass: the clusters found plus per-project warnings
// (permission denied on some projects must not hide the rest).
type Result struct {
	Clusters []models.ClusterInfo
	Warnings []string
}

// Discover lists clusters in every configured project, best-effort per
// project. Discovered entries carry Source=external and, when resolvable, the
// kubeconfig context that reaches them.
func (d *GKEDiscoverer) Discover(ctx context.Context) (Result, error) {
	projects, err := d.Projects(ctx)
	if err != nil {
		return Result{}, err
	}
	contexts := kubeconfigContexts()

	var res Result
	for _, project := range projects {
		result, err := d.exec.Execute(ctx, "gcloud", "container", "clusters", "list", "--project", project, "--format=json")
		if err != nil {
			// Typical: PERMISSION_DENIED on projects outside the user's role, or
			// the container API disabled. Report and keep going.
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: %v", project, err))
			continue
		}
		var clusters []gkeCluster
		if err := json.Unmarshal([]byte(result.Stdout), &clusters); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: unparseable clusters list", project))
			continue
		}
		for _, c := range clusters {
			res.Clusters = append(res.Clusters, models.ClusterInfo{
				Name:       c.Name,
				Type:       models.ClusterTypeGKE,
				Source:     models.SourceExternal,
				Status:     titleCase(c.Status),
				NodeCount:  c.CurrentNodeCount,
				K8sVersion: c.CurrentMasterVersion,
				Project:    project,
				Region:     c.Location,
				Context:    matchContext(contexts, project, c.Location, c.Name),
			})
		}
	}
	return res, nil
}

// kubeconfigContexts returns the context names of the user's kubeconfig
// (honoring $KUBECONFIG); best-effort — a missing kubeconfig is fine.
func kubeconfigContexts() []string {
	cfg, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// matchContext maps a discovered cluster onto a kubeconfig context by the
// three conventional naming shapes; a renamed context cannot be matched and
// yields "".
func matchContext(contexts []string, project, location, name string) string {
	candidates := []string{
		name, // plain (openframe-style or hand-renamed to the cluster name)
		fmt.Sprintf("gke_%s_%s_%s", project, location, name),            // gcloud get-credentials
		fmt.Sprintf("connectgateway_%s_%s_%s", project, location, name), // fleet connect gateway
	}
	for _, want := range candidates {
		for _, have := range contexts {
			if have == want {
				return have
			}
		}
	}
	return ""
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	s = strings.ToLower(s)
	return strings.ToUpper(s[:1]) + s[1:]
}
