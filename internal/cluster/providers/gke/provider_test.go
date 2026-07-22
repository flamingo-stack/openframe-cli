package gke

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCA = base64.StdEncoding.EncodeToString([]byte("fake-ca-pem"))

// fakeRunner is a canned tfexec stand-in.
type fakeRunner struct {
	calls    *[]string
	applyErr error
}

func (f *fakeRunner) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	*f.calls = append(*f.calls, "init")
	return nil
}

func (f *fakeRunner) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	*f.calls = append(*f.calls, "apply")
	return f.applyErr
}

func (f *fakeRunner) ApplyJSON(ctx context.Context, w io.Writer, opts ...tfexec.ApplyOption) error {
	return f.Apply(ctx)
}

func (f *fakeRunner) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	*f.calls = append(*f.calls, "destroy")
	return nil
}

func (f *fakeRunner) DestroyJSON(ctx context.Context, w io.Writer, opts ...tfexec.DestroyOption) error {
	return f.Destroy(ctx)
}

func (f *fakeRunner) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	*f.calls = append(*f.calls, "plan")
	return true, nil
}

func (f *fakeRunner) ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error) {
	*f.calls = append(*f.calls, "show")
	return &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
		{Change: &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionCreate}}},
	}}, nil
}

func (f *fakeRunner) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	*f.calls = append(*f.calls, "output")
	return map[string]tfexec.OutputMeta{
		"cluster_name": {Value: json.RawMessage(`"demo"`)},
		// The GKE module emits a bare host, no scheme — the provider must add it.
		"cluster_endpoint":                   {Value: json.RawMessage(`"34.10.20.30"`)},
		"cluster_certificate_authority_data": {Value: json.RawMessage(`"` + testCA + `"`)},
		"region":                             {Value: json.RawMessage(`"us-central1"`)},
	}, nil
}

// newTestProvider wires the provider onto a temp registry, a fake runner, and
// an isolated kubeconfig.
func newTestProvider(t *testing.T, applyErr error) (*Provider, *[]string, *tfengine.Registry) {
	t.Helper()
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "kubeconfig"))

	calls := &[]string{}
	engine := tfengine.NewEngineWithRunner(func(workdir string) (tfengine.Runner, error) {
		return &fakeRunner{calls: calls, applyErr: applyErr}, nil
	})
	registry := tfengine.NewRegistry(t.TempDir())
	mock := executor.NewMockCommandExecutor() // gcloud preflight succeeds by default
	return NewWithDeps(engine, registry, mock), calls, registry
}

func gkeConfig(name string) models.ClusterConfig {
	return models.ClusterConfig{
		Name:      name,
		Type:      models.ClusterTypeGKE,
		NodeCount: 3,
		Cloud: &models.CloudConfig{
			Region:      "us-central1",
			Project:     "my-project",
			MachineType: "e2-standard-4",
		},
	}
}

func TestCreateCluster_HappyPath(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)

	restConfig, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)
	assert.Equal(t, []string{"init", "plan", "show", "apply", "output"}, *calls,
		"create follows the terraform-apply shape: plan+show before the (auto-approved) apply")
	assert.Equal(t, "https://34.10.20.30", restConfig.Host, "bare module endpoint must be prefixed")
	assert.Equal(t, []byte("fake-ca-pem"), restConfig.CAData)
	require.NotNil(t, restConfig.ExecProvider)
	assert.Equal(t, "gke-gcloud-auth-plugin", restConfig.ExecProvider.Command)
	assert.True(t, restConfig.ExecProvider.ProvideClusterInfo)

	rec, err := registry.Get("demo")
	require.NoError(t, err)
	assert.Equal(t, tfengine.StatusReady, rec.Status)
	assert.Equal(t, "my-project", rec.Project)

	kubeconfig, err := os.ReadFile(os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	assert.Contains(t, string(kubeconfig), "current-context: demo")
}

func TestCreateCluster_FailedApplyKeepsWorkspace(t *testing.T) {
	p, _, registry := newTestProvider(t, errors.New("quota exceeded"))

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "re-run create to resume")

	rec, err := registry.Get("demo")
	require.NoError(t, err)
	assert.Equal(t, tfengine.StatusFailed, rec.Status)
}

func TestCreateCluster_RequiresProjectAndRegion(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)

	noProject := gkeConfig("demo")
	noProject.Cloud.Project = ""
	_, err := p.CreateCluster(context.Background(), noProject)
	var invalid models.ErrInvalidClusterConfig
	require.ErrorAs(t, err, &invalid)

	noCloud := gkeConfig("demo")
	noCloud.Cloud = nil
	_, err = p.CreateCluster(context.Background(), noCloud)
	require.ErrorAs(t, err, &invalid)

	assert.Empty(t, *calls, "terraform must not run without project/region")
}

func TestCreateCluster_CredentialPreflightFailsFast(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "not logged in")
	p.executor = mock

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
	assert.Empty(t, *calls, "terraform must not run with broken credentials")
}

func TestDeleteCluster_DestroysAndRemovesWorkspace(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	require.NoError(t, p.DeleteCluster(context.Background(), "demo", models.ClusterTypeGKE, false))
	assert.Contains(t, *calls, "destroy")

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestStartCluster_Unsupported(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	err := p.StartCluster(context.Background(), "demo", models.ClusterTypeGKE)
	assert.ErrorContains(t, err, "not supported")
}

func TestListAndDetect_FiltersByType(t *testing.T) {
	p, _, registry := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	// An EKS record in the same registry must not leak into the GKE provider.
	eksRecord := tfengine.Record{Name: "other", Type: models.ClusterTypeEKS, Status: tfengine.StatusReady}
	require.NoError(t, registry.Workspace("other").Scaffold(eksRecord, nil, nil))

	clusters, err := p.ListClusters(context.Background())
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	assert.Equal(t, models.ClusterTypeGKE, clusters[0].Type)

	_, err = p.DetectClusterType(context.Background(), "other")
	assert.Error(t, err, "an eks cluster must not detect as gke")
}

func TestGetKubeconfig_RendersExecAuth(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	kubeconfig, err := p.GetKubeconfig(context.Background(), "demo", models.ClusterTypeGKE)
	require.NoError(t, err)
	assert.Contains(t, kubeconfig, "gke-gcloud-auth-plugin")
	assert.Contains(t, kubeconfig, "provideClusterInfo: true")
}

func TestTfvarsFor_VersionMapping(t *testing.T) {
	base := gkeConfig("demo")

	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"latest", "", false},
		{"1.33", "1.33", false},
		{"v1.33", "1.33", false},
		{"v1.31.5-k3s1", "", true},
	}
	for _, tc := range cases {
		config := base
		config.K8sVersion = tc.in
		vars, err := tfvarsFor(config)
		if tc.wantErr {
			assert.Error(t, err, "input %q", tc.in)
			continue
		}
		require.NoError(t, err, "input %q", tc.in)
		assert.Equal(t, tc.want, vars.KubernetesVersion, "input %q", tc.in)
	}
}

func TestTemplateEmbedsModulePins(t *testing.T) {
	tf := string(mainTF)
	assert.Contains(t, tf, `source  = "terraform-google-modules/kubernetes-engine/google//modules/private-cluster"`)
	assert.Contains(t, tf, `version = "~> 44.0"`)
	assert.Contains(t, tf, `source  = "terraform-google-modules/network/google"`)
	assert.Contains(t, tf, `version = "~> 18.0"`)
	assert.Contains(t, tf, "deletion_protection = false")
	// Org-policy compatibility (restrict_vm_external_ips): private nodes with
	// NAT egress and a public control-plane endpoint.
	assert.Contains(t, tf, "enable_private_nodes    = true")
	assert.Contains(t, tf, "enable_private_endpoint = false")
	assert.Contains(t, tf, `resource "google_compute_router_nat" "nat"`)
}

// TestCreateCluster_ResumeRefreshesTemplate: a retry after a failed create
// must regenerate main.tf from the CURRENT embedded template so template
// bugfixes reach existing workspaces.
func TestCreateCluster_ResumeRefreshesTemplate(t *testing.T) {
	p, _, registry := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	// Simulate a stale workspace from an older CLI version.
	stalePath := filepath.Join(registry.Workspace("demo").TerraformDir(), "main.tf")
	require.NoError(t, os.WriteFile(stalePath, []byte("# stale broken template"), 0o600))

	_, err = p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	refreshed, err := os.ReadFile(stalePath)
	require.NoError(t, err)
	assert.Contains(t, string(refreshed), "enable_private_nodes", "resume must rewrite main.tf from the current template")
}

func TestPlanCluster_NewClusterDoesNotRegister(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)

	summary, err := p.PlanCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)
	assert.True(t, summary.HasChanges())
	assert.Equal(t, []string{"init", "plan", "show"}, *calls)

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound, "a plan preview must not register the cluster")
}

func TestCreateCluster_WritesGCSBackend(t *testing.T) {
	p, _, registry := newTestProvider(t, nil)
	config := gkeConfig("demo")
	config.Cloud.BackendConfig = "gcs://my-bucket/clusters/demo"

	_, err := p.CreateCluster(context.Background(), config)
	require.NoError(t, err)

	backend, err := os.ReadFile(filepath.Join(registry.Workspace("demo").TerraformDir(), "backend.tf"))
	require.NoError(t, err)
	assert.Contains(t, string(backend), `backend "gcs"`)
	assert.Contains(t, string(backend), `bucket = "my-bucket"`)
	assert.Contains(t, string(backend), `prefix = "clusters/demo"`)
}

func TestCreateCluster_RejectsS3Backend(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	config := gkeConfig("demo")
	config.Cloud.BackendConfig = "s3://my-bucket/prefix"

	_, err := p.CreateCluster(context.Background(), config)
	var invalid models.ErrInvalidClusterConfig
	require.ErrorAs(t, err, &invalid)
	assert.Contains(t, err.Error(), "must be gcs://")
	assert.Empty(t, *calls)
}

// TestCreateCluster_RefusesExternalNameCollision locks the ownership boundary:
// a cluster that already exists in the project WITHOUT an openframe workspace
// is somebody else's — create must refuse before any terraform runs (terraform
// would build the VPC first and fail mid-apply, leaving billed debris).
func TestCreateCluster_RefusesExternalNameCollision(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("clusters list --project my-project", &executor.CommandResult{ExitCode: 0, Stdout: "demo\n"})
	p.executor = mock

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not managed by openframe")
	assert.Empty(t, *calls, "terraform must not run on a name collision")

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound, "no workspace must be scaffolded")
}

// TestCreateCluster_ResumeSkipsCollisionCheck: an existing workspace means the
// cloud cluster is OURS (possibly partially created) — resume must proceed
// even though describe would find it.
func TestCreateCluster_ResumeSkipsCollisionCheck(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("clusters list --project my-project", &executor.CommandResult{ExitCode: 0, Stdout: "demo\n"})
	p.executor = mock

	*calls = nil
	_, err = p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err, "resume of an owned cluster must not be blocked by the collision check")
	assert.Contains(t, *calls, "apply")
}

// TestCreateCluster_RefusesKubeconfigContextClobber: a same-named kubeconfig
// context pointing at a DIFFERENT server belongs to something else and must
// not be overwritten.
func TestCreateCluster_RefusesKubeconfigContextClobber(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	kubeconfig := `apiVersion: v1
kind: Config
clusters:
- name: other
  cluster:
    server: https://somebody-elses.example:6443
contexts:
- name: demo
  context:
    cluster: other
    user: other
users:
- name: other
`
	require.NoError(t, os.WriteFile(os.Getenv("KUBECONFIG"), []byte(kubeconfig), 0o600))

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to overwrite")
}

// TestDeleteCluster_LeavesRepointedContextAlone (П4): if the user repointed a
// same-named kubeconfig context at another server after the create, delete
// must not remove it — the mirror of the create-side no-clobber guard.
func TestDeleteCluster_LeavesRepointedContextAlone(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.NoError(t, err)

	// Repoint the "demo" context at somebody else's server.
	repointed := `apiVersion: v1
kind: Config
current-context: demo
clusters:
- name: demo
  cluster:
    server: https://somebody-elses.example:6443
contexts:
- name: demo
  context:
    cluster: demo
    user: demo
users:
- name: demo
`
	require.NoError(t, os.WriteFile(os.Getenv("KUBECONFIG"), []byte(repointed), 0o600))

	require.NoError(t, p.DeleteCluster(context.Background(), "demo", models.ClusterTypeGKE, false))

	kubeconfig, err := os.ReadFile(os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	assert.Contains(t, string(kubeconfig), "somebody-elses.example", "a repointed context must survive delete")
	assert.Contains(t, string(kubeconfig), "current-context: demo")
}

// TestPreflightNameCollision_FindsZonalClusters (П3): the collision check uses
// a project-wide listing, so a zonal cluster with the same name (which a
// region-scoped describe would miss) also counts.
func TestPreflightNameCollision_FindsZonalClusters(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	mock := executor.NewMockCommandExecutor()
	// clusters list output for a zonal cluster: same name, any location.
	mock.SetResponse("clusters list --project my-project", &executor.CommandResult{ExitCode: 0, Stdout: "demo\n"})
	p.executor = mock

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not managed by openframe")
	assert.Empty(t, *calls)
}

// TestCreateCluster_DeclinedPlanAppliesNothing: the interactive plan gate —
// a declined plan must apply nothing, and a declined BRAND-NEW create must
// leave no workspace behind (nothing exists to resume or bill).
func TestCreateCluster_DeclinedPlanAppliesNothing(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)
	p.confirmApply = func(summary tfengine.PlanSummary) bool { return false }

	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.NotContains(t, *calls, "apply", "a declined plan must not apply")

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound, "declined brand-new create must remove the fresh workspace")
}

// TestCreateCluster_DeclinedResumeKeepsWorkspace: declining a RESUME keeps
// the workspace — its state still points at real (billed) resources.
func TestCreateCluster_DeclinedResumeKeepsWorkspace(t *testing.T) {
	p, _, registry := newTestProvider(t, errors.New("first apply fails"))
	_, err := p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err) // failed create leaves a resumable workspace

	p.confirmApply = func(summary tfengine.PlanSummary) bool { return false }
	_, err = p.CreateCluster(context.Background(), gkeConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")

	_, err = registry.Get("demo")
	require.NoError(t, err, "declined resume must keep the workspace and its state")
}
