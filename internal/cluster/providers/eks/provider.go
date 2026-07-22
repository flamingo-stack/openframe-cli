// Package eks implements the cluster Provider for AWS EKS. Provisioning runs
// through the shared terraform engine: the provider generates a root module
// (public terraform-aws-modules/eks + vpc modules, pinned) into the cluster's
// workspace and drives init/apply/destroy there. See the package comment in
// internal/cluster/providers/terraform for the workspace layout.
package eks

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

// Provider provisions and manages EKS clusters.
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

// preflightCredentials fails fast with an actionable message when the AWS
// identity is unusable — before any terraform runs.
func (p *Provider) preflightCredentials(ctx context.Context, profile string) error {
	args := []string{"sts", "get-caller-identity", "--output", "json"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	if _, err := p.executor.Execute(ctx, "aws", args...); err != nil {
		which := "default credentials"
		if profile != "" {
			which = fmt.Sprintf("profile '%s'", profile)
		}
		return fmt.Errorf("AWS %s cannot authenticate (aws sts get-caller-identity failed): %w", which, err)
	}
	return nil
}

// backendTF renders the s3 backend block for an EKS workspace.
func backendTF(cfg tfengine.BackendConfig, region string) []byte {
	key := "terraform.tfstate"
	if cfg.Prefix != "" {
		key = cfg.Prefix + "/terraform.tfstate"
	}
	return []byte(fmt.Sprintf(
		"terraform {\n  backend \"s3\" {\n    bucket = %q\n    key    = %q\n    region = %q\n  }\n}\n",
		cfg.Bucket, key, region))
}

// parseBackend validates the optional --backend-config value for EKS.
func parseBackend(config models.ClusterConfig) (*tfengine.BackendConfig, error) {
	if config.Cloud.BackendConfig == "" {
		return nil, nil
	}
	cfg, err := tfengine.ParseBackendURL(config.Cloud.BackendConfig)
	if err != nil {
		return nil, models.NewInvalidConfigError("backend-config", config.Cloud.BackendConfig, err.Error())
	}
	if cfg.Scheme != "s3" {
		return nil, models.NewInvalidConfigError("backend-config", config.Cloud.BackendConfig, "EKS remote state must be s3://bucket/prefix")
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
	if err := p.preflightCredentials(ctx, config.Cloud.Profile); err != nil {
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

// preflightNameCollision refuses to create a cluster whose name already
// exists in the target account/region but has no openframe workspace —
// terraform would build the VPC first and fail mid-apply on the duplicate
// cluster, leaving partial billed infrastructure (the GKE twin's rationale).
// Existence criterion: describe exits 0 AND prints exactly the name.
func (p *Provider) preflightNameCollision(ctx context.Context, config models.ClusterConfig) error {
	args := []string{"eks", "describe-cluster", "--name", config.Name,
		"--region", config.Cloud.Region, "--query", "cluster.name", "--output", "text"}
	if config.Cloud.Profile != "" {
		args = append(args, "--profile", config.Cloud.Profile)
	}
	result, err := p.executor.Execute(ctx, "aws", args...)
	if err != nil || result == nil || strings.TrimSpace(result.Stdout) != config.Name {
		return nil // not found (or indeterminate) — proceed
	}
	return fmt.Errorf("cluster '%s' already exists in region '%s' but is not managed by openframe — refusing to touch it; pick another cluster name",
		config.Name, config.Cloud.Region)
}

// CreateCluster provisions the cluster and returns a rest.Config for it.
// Re-running after a failed apply resumes the same workspace: terraform apply
// is idempotent over the recorded state.
func (p *Provider) CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error) {
	if err := validate(config); err != nil {
		return nil, err
	}
	backend, err := parseBackend(config)
	if err != nil {
		return nil, err
	}
	if err := p.preflightCredentials(ctx, config.Cloud.Profile); err != nil {
		return nil, err
	}

	ws := p.registry.Workspace(config.Name)
	freshWorkspace := !ws.Exists()
	if freshWorkspace {
		if err := p.preflightNameCollision(ctx, config); err != nil {
			return nil, err
		}
		vars, err := tfvarsFor(config)
		if err != nil {
			return nil, err
		}
		record := tfengine.Record{
			Name:       config.Name,
			Type:       models.ClusterTypeEKS,
			Status:     tfengine.StatusCreating,
			Region:     config.Cloud.Region,
			Profile:    config.Cloud.Profile,
			K8sVersion: vars.KubernetesVersion,
			NodeCount:  config.NodeCount,
			CreatedAt:  time.Now().UTC(),
		}
		if err := ws.Scaffold(record, mainTF, vars); err != nil {
			return nil, err
		}
		if backend != nil {
			if err := ws.WriteBackend(backendTF(*backend, config.Cloud.Region)); err != nil {
				return nil, err
			}
		}
	}
	// An existing workspace means a previous create failed or was interrupted.
	// Refresh the generated module from the CURRENT template before resuming:
	// the retry must pick up template bugfixes (e.g. the private-nodes fix for
	// org-policy environments), not replay the broken files. The state is
	// untouched — terraform reconciles it against the refreshed config.
	if ws.Exists() {
		vars, err := tfvarsFor(config)
		if err != nil {
			return nil, err
		}
		if err := tfengine.WriteModule(ws.TerraformDir(), mainTF, vars); err != nil {
			return nil, err
		}
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
	if record.Endpoint, err = tfengine.StringOutput(outputs, "cluster_endpoint"); err != nil {
		return nil, models.NewClusterOperationError("create", config.Name, err)
	}
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
	if clusterType != models.ClusterTypeEKS {
		return models.NewProviderNotFoundError(clusterType)
	}
	ws := p.registry.Workspace(name)
	if !ws.Exists() {
		return models.NewClusterNotFoundError(name)
	}
	// Read the record BEFORE destroy: the endpoint in it is what proves the
	// kubeconfig entry is ours to remove afterwards.
	rec, recErr := ws.ReadRecord()
	if err := p.engine.Destroy(ctx, ws.TerraformDir()); err != nil {
		return models.NewClusterOperationError("delete", name,
			fmt.Errorf("%w\nThe terraform state is kept in %s; re-run delete to retry", err, ws.Dir()))
	}
	if recErr == nil {
		_ = removeFromDefaultKubeconfig(rec)
	}
	return ws.Remove()
}

// StartCluster is meaningless for a managed control plane.
func (p *Provider) StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error {
	return fmt.Errorf("starting is not supported for EKS clusters: the managed control plane is always running")
}

// ListClusters returns the EKS clusters recorded in the local registry.
func (p *Provider) ListClusters(ctx context.Context) ([]models.ClusterInfo, error) {
	records, err := p.registry.List()
	if err != nil {
		return nil, err
	}
	infos := make([]models.ClusterInfo, 0, len(records))
	for _, rec := range records {
		if rec.Type != models.ClusterTypeEKS {
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
	if err != nil || rec.Type != models.ClusterTypeEKS {
		return models.ClusterInfo{}, models.NewClusterNotFoundError(name)
	}
	return infoFor(rec), nil
}

// DetectClusterType reports eks for registry-recorded clusters.
func (p *Provider) DetectClusterType(ctx context.Context, name string) (models.ClusterType, error) {
	rec, err := p.registry.Get(name)
	if err != nil || rec.Type != models.ClusterTypeEKS {
		return "", models.NewClusterNotFoundError(name)
	}
	return models.ClusterTypeEKS, nil
}

// GetRestConfig builds a rest.Config from the recorded endpoint/CA — no
// terraform run needed.
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
	if clusterType != models.ClusterTypeEKS {
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
		Type:       models.ClusterTypeEKS,
		Source:     models.SourceOpenframe,
		Context:    rec.Name,
		Region:     rec.Region,
		Status:     strings.ToTitle(string(rec.Status[0:1])) + string(rec.Status[1:]),
		NodeCount:  rec.NodeCount,
		K8sVersion: rec.K8sVersion,
		CreatedAt:  rec.CreatedAt,
	}
}
