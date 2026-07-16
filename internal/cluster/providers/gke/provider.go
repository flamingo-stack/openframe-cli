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
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Provider provisions and manages GKE clusters.
type Provider struct {
	engine   *tfengine.Engine
	registry *tfengine.Registry
	executor executor.CommandExecutor
}

// New builds the production provider. The registry defaults to
// ~/.openframe/clusters.
func New(exec executor.CommandExecutor, verbose bool) (*Provider, error) {
	registry, err := tfengine.DefaultRegistry()
	if err != nil {
		return nil, err
	}
	return &Provider{
		engine:   tfengine.NewEngine(verbose),
		registry: registry,
		executor: exec,
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

	ws := p.registry.Workspace(config.Name)
	if !ws.Exists() {
		vars, err := tfvarsFor(config)
		if err != nil {
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
	}
	// An existing workspace means a previous create failed or was interrupted;
	// keep its tfvars (the state may reference them) and simply resume.

	if err := p.engine.Init(ctx, ws.TerraformDir()); err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
	if err := p.engine.Apply(ctx, ws.TerraformDir()); err != nil {
		_ = ws.SetStatus(tfengine.StatusFailed)
		return nil, models.NewClusterOperationError("create", config.Name,
			fmt.Errorf("%w\nThe terraform state is kept in %s; re-run create to resume or 'openframe cluster delete %s' to tear down", err, ws.Dir(), config.Name))
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
	if err := p.engine.Destroy(ctx, ws.TerraformDir()); err != nil {
		return models.NewClusterOperationError("delete", name,
			fmt.Errorf("%w\nThe terraform state is kept in %s; re-run delete to retry", err, ws.Dir()))
	}
	_ = removeFromDefaultKubeconfig(name)
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
		Status:     strings.ToTitle(string(rec.Status[0:1])) + string(rec.Status[1:]),
		NodeCount:  rec.NodeCount,
		K8sVersion: rec.K8sVersion,
		CreatedAt:  rec.CreatedAt,
	}
}
