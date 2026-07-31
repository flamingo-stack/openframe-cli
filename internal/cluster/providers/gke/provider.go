// Package gke implements the cluster Provider for Google Kubernetes Engine.
// Provisioning runs through the shared terraform engine: the provider
// generates a root module (public terraform-google-modules/kubernetes-engine
// + network modules, pinned) into the cluster's workspace and drives
// init/apply/destroy there. See the package comment in
// internal/cluster/providers/terraform for the workspace layout.
package gke

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// gcpResourcePathRE pulls the GCP resource path (projects/.../<kind>/<name>)
// out of a terraform 409 error so an orphan can be named concretely.
var gcpResourcePathRE = regexp.MustCompile(`projects/[^'"\s]+`)

// orphanFromInterruptedCreate detects the specific failure where terraform
// tries to create a resource that already exists in GCP (HTTP 409 /
// alreadyExists). This is the signature of a create interrupted (SIGINT) after
// the cloud API created a resource but before terraform saved it to state: the
// resource is real but state-invisible, so every resume collides with it (409)
// and 'cluster delete' — which only knows state-tracked resources — cannot
// remove it. Returns human-readable remediation and true when this is that
// case, so the caller can replace the generic "re-run to resume" hint (which
// would loop forever here) with something actionable.
func orphanFromInterruptedCreate(err error, terraformDir string) (string, bool) {
	msg := err.Error()
	low := strings.ToLower(msg)
	if !strings.Contains(low, "alreadyexists") && !strings.Contains(low, "409") {
		return "", false
	}
	resource := "the resource named in the error above"
	if m := gcpResourcePathRE.FindString(msg); m != "" {
		resource = m
	}
	return fmt.Sprintf(
		"a resource already exists in GCP that terraform is not tracking:\n"+
			"  %s\n"+
			"This is the signature of a create that was interrupted after the resource was\n"+
			"created but before its state was saved — so resume keeps colliding with it and\n"+
			"'cluster delete' cannot remove it (delete only knows state-tracked resources).\n"+
			"Resolve it one of two ways, then re-run create:\n"+
			"  • delete the orphan in GCP (Cloud Console or the matching 'gcloud ... delete'), or\n"+
			"  • import it into this cluster's state, e.g.:\n"+
			"      terraform -chdir=%s import <resource.address> %s",
		resource, terraformDir, resource), true
}

// Provider provisions and manages GKE clusters.
type Provider struct {
	engine   *tfengine.Engine
	registry *tfengine.Registry
	executor executor.CommandExecutor
	// confirmApply, when set, is asked before applying the create plan (the
	// interactive `terraform apply` shape). Nil means auto-approve — the
	// non-interactive/programmatic behavior and the test default.
	confirmApply func(tfengine.PlanSummary) bool
}

// New builds the production provider. The registry defaults to
// ~/.openframe/clusters.
func New(exec executor.CommandExecutor, verbose bool) (*Provider, error) {
	registry, err := tfengine.DefaultRegistry()
	if err != nil {
		return nil, err
	}
	return &Provider{
		engine:       tfengine.NewEngine(verbose),
		registry:     registry,
		executor:     exec,
		confirmApply: tfengine.ConfirmApplyInteractive,
	}, nil
}

// NewWithDeps is the test constructor.
func NewWithDeps(engine *tfengine.Engine, registry *tfengine.Registry, exec executor.CommandExecutor) *Provider {
	return &Provider{engine: engine, registry: registry, executor: exec}
}

// preflightCredentials fails fast with an actionable message when the gcloud
// identity or project access is unusable — before any terraform runs.
func (p *Provider) preflightCredentials(ctx context.Context, project string) error {
	if _, err := p.executor.Execute(ctx, "gcloud", "auth", "print-access-token", "--quiet"); err != nil {
		return fmt.Errorf("gcloud is not authenticated (run 'gcloud auth login' and 'gcloud auth application-default login'): %w", err)
	}
	if _, err := p.executor.Execute(ctx, "gcloud", "projects", "describe", project, "--format=value(projectId)"); err != nil {
		return fmt.Errorf("GCP project '%s' is not accessible with the current gcloud identity: %w", project, err)
	}
	return nil
}

// preflightNameCollision refuses to create a cluster whose name already
// exists in the target project but has no openframe workspace: terraform
// would build the VPC first and then fail mid-apply on the duplicate cluster,
// leaving partial billed infrastructure. External clusters are strictly not
// ours to touch — the user must pick another name.
//
// Existence criterion: the project-wide listing (deliberately NOT scoped to a
// location — a ZONAL cluster with the same name must also count) prints a
// line equal to the cluster name. Anything else — non-zero exit, empty or
// unrelated output — is treated as "does not exist"; a genuinely broken API
// call fails later with a clearer terraform error anyway.
func (p *Provider) preflightNameCollision(ctx context.Context, config models.ClusterConfig) error {
	result, err := p.executor.Execute(ctx, "gcloud", "container", "clusters", "list",
		"--project", config.Cloud.Project, "--filter=name="+config.Name, "--format=value(name)")
	if err != nil || result == nil {
		return nil // indeterminate — proceed
	}
	for _, line := range strings.Split(result.Stdout, "\n") {
		if strings.TrimSpace(line) == config.Name {
			return fmt.Errorf("cluster '%s' already exists in project '%s' but is not managed by openframe — refusing to touch it; pick another cluster name",
				config.Name, config.Cloud.Project)
		}
	}
	return nil
}

// firstZoneInRegion returns the alphabetically-first zone of a region (e.g.
// "us-central1-a"), used as the single location for a zonal cluster. Sorting
// makes the choice deterministic across runs — a resume must not pick a
// different zone and try to move the cluster. gcloud is already a hard GKE
// prerequisite, so this adds no new dependency.
func (p *Provider) firstZoneInRegion(ctx context.Context, project, region string) (string, error) {
	res, err := p.executor.Execute(ctx, "gcloud", "compute", "zones", "list",
		"--project", project,
		"--filter", "name~^"+region+"-",
		"--format=value(name)", "--sort-by=name", "--limit=1")
	if err != nil {
		return "", fmt.Errorf("could not list zones for region %s (enable the Compute Engine API for project %q and check your gcloud access): %w", region, project, err)
	}
	var zone string
	if res != nil {
		zone = strings.TrimSpace(strings.SplitN(strings.TrimSpace(res.Stdout), "\n", 2)[0])
	}
	if zone == "" {
		return "", fmt.Errorf("no zones found for region %q in project %q — is %q a valid GCP region?", region, project, region)
	}
	return zone, nil
}

// ensureZone fills config.Cloud.Zone for a zonal (non-HA) cluster so the module
// has a concrete location and the node count is exact. No-op for HA (regional)
// clusters and when a zone is already set.
func (p *Provider) ensureZone(ctx context.Context, config *models.ClusterConfig) error {
	if config.Cloud == nil || config.Cloud.HA || config.Cloud.Zone != "" {
		return nil
	}
	zone, err := p.firstZoneInRegion(ctx, config.Cloud.Project, config.Cloud.Region)
	if err != nil {
		return err
	}
	config.Cloud.Zone = zone
	return nil
}

// backendTF renders the gcs backend block for a GKE workspace.
func backendTF(cfg tfengine.BackendConfig) []byte {
	return []byte(fmt.Sprintf(
		"terraform {\n  backend \"gcs\" {\n    bucket = %q\n    prefix = %q\n  }\n}\n",
		cfg.Bucket, cfg.Prefix))
}

// parseBackend validates the optional --backend-config value for GKE.
func parseBackend(config models.ClusterConfig) (*tfengine.BackendConfig, error) {
	if config.Cloud.BackendConfig == "" {
		return nil, nil
	}
	cfg, err := tfengine.ParseBackendURL(config.Cloud.BackendConfig)
	if err != nil {
		return nil, models.NewInvalidConfigError("backend-config", config.Cloud.BackendConfig, err.Error())
	}
	if cfg.Scheme != "gcs" {
		return nil, models.NewInvalidConfigError("backend-config", config.Cloud.BackendConfig, "GKE remote state must be gcs://bucket/prefix")
	}
	return &cfg, nil
}

// PlanCluster previews what CreateCluster would do — a real terraform plan —
// without registering the cluster or touching any state. A brand-new cluster
// is planned in a throwaway directory; an existing (failed/interrupted)
// workspace is planned in place to show what a resume would change.
func (p *Provider) PlanCluster(ctx context.Context, config models.ClusterConfig) (tfengine.PlanSummary, error) {
	if err := validate(config); err != nil {
		return tfengine.PlanSummary{}, err
	}
	if err := p.preflightCredentials(ctx, config.Cloud.Project); err != nil {
		return tfengine.PlanSummary{}, err
	}
	if err := p.ensureZone(ctx, &config); err != nil {
		return tfengine.PlanSummary{}, err
	}

	dir := p.registry.Workspace(config.Name).TerraformDir()
	if !p.registry.Workspace(config.Name).Exists() {
		vars, err := tfvarsFor(config)
		if err != nil {
			return tfengine.PlanSummary{}, err
		}
		tmp, err := os.MkdirTemp("", "openframe-plan-*")
		if err != nil {
			return tfengine.PlanSummary{}, err
		}
		defer func() { _ = os.RemoveAll(tmp) }()
		if err := tfengine.WriteModule(tmp, mainTF, vars); err != nil {
			return tfengine.PlanSummary{}, err
		}
		dir = tmp
	}

	if err := p.engine.Init(ctx, dir); err != nil {
		return tfengine.PlanSummary{}, err
	}
	return p.engine.Plan(ctx, dir)
}

// CreateCluster provisions the cluster and returns a rest.Config for it.
// Re-running after a failed apply resumes the same workspace.
func (p *Provider) CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error) {
	if err := validate(config); err != nil {
		return nil, err
	}
	backend, err := parseBackend(config)
	if err != nil {
		return nil, err
	}
	if err := p.preflightCredentials(ctx, config.Cloud.Project); err != nil {
		return nil, err
	}
	if err := p.ensureZone(ctx, &config); err != nil {
		return nil, err
	}

	ws := p.registry.Workspace(config.Name)
	freshWorkspace := !ws.Exists()
	vars, err := tfvarsFor(config)
	if err != nil {
		return nil, err
	}
	// The collision check only guards NEW clusters: an existing workspace means
	// the cloud cluster (partial or complete) is ours and create resumes it.
	if freshWorkspace {
		if err := p.preflightNameCollision(ctx, config); err != nil {
			return nil, err
		}
		record := tfengine.Record{
			Name:       config.Name,
			Type:       models.ClusterTypeGKE,
			Status:     tfengine.StatusCreating,
			Region:     config.Cloud.Region,
			Project:    config.Cloud.Project,
			K8sVersion: vars.KubernetesVersion,
			NodeCount:  config.NodeCount,
			CreatedAt:  time.Now().UTC(),
		}
		if err := ws.Scaffold(record, mainTF, vars); err != nil {
			return nil, err
		}
		if backend != nil {
			if err := ws.WriteBackend(backendTF(*backend)); err != nil {
				return nil, err
			}
		}
	} else if backend != nil {
		// Repointing an existing workspace's state backend needs a terraform
		// state migration this CLI does not drive. Dropping the flag silently
		// would leave the user believing their state moved — say it out loud.
		pterm.Warning.Println("--backend-config is ignored when resuming an existing workspace; the backend chosen at first create stays in effect")
	}
	// An existing workspace means a previous create failed or was interrupted.
	// Refresh the generated module from the CURRENT template before resuming:
	// the retry must pick up template bugfixes (e.g. the private-nodes fix for
	// org-policy environments), not replay the broken files. The state is
	// untouched — terraform reconciles it against the refreshed config.
	if err := tfengine.WriteModule(ws.TerraformDir(), mainTF, vars); err != nil {
		return nil, err
	}

	if err := p.engine.Init(ctx, ws.TerraformDir()); err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}

	// The `terraform apply` shape: plan, show, confirm, then apply the SAVED
	// plan — what the user approved is exactly what runs.
	summary, planFile, err := p.engine.PlanForApply(ctx, ws.TerraformDir())
	if planFile != "" {
		defer func() { _ = os.Remove(planFile) }()
	}
	if err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	if p.confirmApply != nil && !p.confirmApply(summary) {
		if freshWorkspace {
			// Nothing was applied — a declined brand-new create leaves no trace.
			_ = ws.Remove()
		}
		return nil, fmt.Errorf("cluster creation cancelled — no changes were applied")
	}
	if err := p.engine.ApplyPlan(ctx, ws.TerraformDir(), planFile); err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		if hint, ok := orphanFromInterruptedCreate(err, ws.TerraformDir()); ok {
			return nil, models.NewClusterOperationError("create", config.Name,
				fmt.Errorf("%w\n\n%s", err, hint))
		}
		// Carry the resume hint structurally as well as in the text: on Ctrl+C the
		// interruption handler prints only "Operation cancelled by user." and drops
		// the message, so a text-only hint would never reach the user even though
		// the workspace state is preserved and the create IS resumable.
		resumeHint := fmt.Sprintf("The terraform state is kept in %s; re-run create to resume or 'openframe cluster delete %s' to tear down", ws.Dir(), config.Name)
		return nil, models.NewClusterOperationError("create", config.Name,
			withResumeHint(fmt.Errorf("%w\n%s", err, resumeHint), resumeHint))
	}

	outputs, err := p.engine.Outputs(ctx, ws.TerraformDir())
	if err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	record, err := ws.ReadRecord()
	if err != nil {
		return nil, err
	}
	// The module applied above was regenerated from the CURRENT config; make
	// the record agree with it, or a resumed create with changed flags (e.g.
	// --nodes) reports the first attempt's values from list/status forever.
	record.Region = config.Cloud.Region
	record.NodeCount = config.NodeCount
	record.K8sVersion = vars.KubernetesVersion
	endpoint, err := tfengine.StringOutput(outputs, "cluster_endpoint")
	if err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	// The GKE module emits a bare host; kubeconfig/rest need a URL.
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	record.Endpoint = endpoint
	if record.CACert, err = tfengine.StringOutput(outputs, "cluster_certificate_authority_data"); err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	record.Status = tfengine.StatusReady
	if err := ws.WriteRecord(record); err != nil {
		return nil, err
	}

	if err := mergeIntoDefaultKubeconfig(record); err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	return restConfigFor(record)
}

// DeleteCluster destroys the cluster's cloud resources, then removes the
// workspace and the kubeconfig context. The workspace survives a failed
// destroy — its state is the only pointer to still-billed resources.
func (p *Provider) DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error {
	if clusterType != models.ClusterTypeGKE {
		return models.NewProviderNotFoundError(clusterType)
	}
	ws := p.registry.Workspace(name)
	if !ws.Exists() {
		return models.NewClusterNotFoundError(name)
	}
	// Read the record BEFORE destroy: the endpoint in it is what proves the
	// kubeconfig entry is ours to remove afterwards, and the project is needed
	// for the post-destroy orphan-disk sweep.
	rec, recErr := ws.ReadRecord()
	if recErr == nil {
		// Tear down app workloads first so the CSI driver deletes PVC-backed
		// Persistent Disks before the node pool dies — otherwise they orphan as
		// billable leftovers. Best-effort; never blocks the destroy.
		releaseWorkloadDisks(ctx, rec)
	}
	if err := p.engine.Destroy(ctx, ws.TerraformDir()); err != nil {
		return models.NewClusterOperationError("delete", name,
			fmt.Errorf("%w\nThe terraform state is kept in %s; re-run delete to retry", err, ws.Dir()))
	}
	if recErr == nil {
		_ = removeFromDefaultKubeconfig(rec)
		// Sweep any PVC-provisioned disks that outlived the destroy: delete them
		// when the caller consented to an unattended teardown (--force), else
		// report them — a "cleaned up" delete must never silently leave orphans.
		p.sweepOrphanedDisks(ctx, rec.Project, name, rec.Region, force)
	}
	return ws.Remove()
}

// StartCluster is meaningless for a managed control plane.
func (p *Provider) StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error {
	return fmt.Errorf("starting is not supported for GKE clusters: the managed control plane is always running")
}

// ListClusters returns the GKE clusters recorded in the local registry.
func (p *Provider) ListClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	records, err := p.registry.List()
	if err != nil {
		return nil, err
	}
	infos := make([]models.ClusterInfo, 0, len(records))
	for _, rec := range records {
		if rec.Type != models.ClusterTypeGKE {
			continue
		}
		infos = append(infos, infoFor(rec))
	}
	return infos, nil
}

// ListAllClusters is the same as ListClusters: the registry is this
// provider's full visibility.
func (p *Provider) ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	return p.ListClusters(ctx)
}

// GetClusterStatus returns the recorded status for a single cluster.
func (p *Provider) GetClusterStatus(ctx context.Context, name string) (models.ClusterInfo, error) {
	rec, err := p.registry.Get(name)
	if err != nil || rec.Type != models.ClusterTypeGKE {
		return models.ClusterInfo{}, models.NewClusterNotFoundError(name)
	}
	return infoFor(rec), nil
}

// DetectClusterType reports gke for registry-recorded clusters.
func (p *Provider) DetectClusterType(ctx context.Context, name string) (models.ClusterType, error) {
	rec, err := p.registry.Get(name)
	if err != nil || rec.Type != models.ClusterTypeGKE {
		return "", models.NewClusterNotFoundError(name)
	}
	return models.ClusterTypeGKE, nil
}

// GetRestConfig builds a rest.Config from the recorded endpoint/CA.
func (p *Provider) GetRestConfig(ctx context.Context, name string) (*rest.Config, error) {
	rec, err := p.registry.Get(name)
	if err != nil {
		return nil, err
	}
	if rec.Status != tfengine.StatusReady {
		return nil, fmt.Errorf("cluster '%s' is not ready (status: %s)", name, rec.Status)
	}
	return restConfigFor(rec)
}

// GetKubeconfig renders the cluster's kubeconfig as YAML.
func (p *Provider) GetKubeconfig(ctx context.Context, name string, clusterType models.ClusterType) (string, error) {
	if clusterType != models.ClusterTypeGKE {
		return "", models.NewProviderNotFoundError(clusterType)
	}
	rec, err := p.registry.Get(name)
	if err != nil {
		return "", err
	}
	cfg, err := kubeconfigFor(rec)
	if err != nil {
		return "", err
	}
	data, err := clientcmd.Write(*cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// infoFor maps a registry record onto the shared ClusterInfo shape.
func infoFor(rec tfengine.Record) models.ClusterInfo {
	return models.ClusterInfo{
		Name:       rec.Name,
		Type:       models.ClusterTypeGKE,
		Source:     models.SourceOpenframe,
		Context:    rec.Name,
		Project:    rec.Project,
		Region:     rec.Region,
		Status:     rec.Status.Title(),
		NodeCount:  rec.NodeCount,
		K8sVersion: rec.K8sVersion,
		CreatedAt:  rec.CreatedAt,
	}
}
