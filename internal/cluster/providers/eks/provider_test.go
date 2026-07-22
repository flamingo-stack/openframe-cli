package eks

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
		"cluster_name":                       {Value: json.RawMessage(`"demo"`)},
		"cluster_endpoint":                   {Value: json.RawMessage(`"https://demo.eks.example"`)},
		"cluster_certificate_authority_data": {Value: json.RawMessage(`"` + testCA + `"`)},
		"region":                             {Value: json.RawMessage(`"us-east-1"`)},
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
	mock := executor.NewMockCommandExecutor() // aws sts get-caller-identity succeeds by default
	return NewWithDeps(engine, registry, mock), calls, registry
}

func eksConfig(name string) models.ClusterConfig {
	return models.ClusterConfig{
		Name:      name,
		Type:      models.ClusterTypeEKS,
		NodeCount: 3,
		Cloud:     &models.CloudConfig{Region: "us-east-1", MachineType: "m6i.large"},
	}
}

func TestCreateCluster_HappyPath(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)

	restConfig, err := p.CreateCluster(context.Background(), eksConfig("demo"))
	require.NoError(t, err)
	assert.Equal(t, []string{"init", "plan", "show", "apply", "output"}, *calls,
		"create follows the terraform-apply shape: plan+show before the (auto-approved) apply")
	assert.Equal(t, "https://demo.eks.example", restConfig.Host)
	assert.Equal(t, []byte("fake-ca-pem"), restConfig.CAData)
	require.NotNil(t, restConfig.ExecProvider)
	assert.Equal(t, "aws", restConfig.ExecProvider.Command)
	assert.Contains(t, restConfig.ExecProvider.Args, "get-token")

	rec, err := registry.Get("demo")
	require.NoError(t, err)
	assert.Equal(t, tfengine.StatusReady, rec.Status)
	assert.Equal(t, "https://demo.eks.example", rec.Endpoint)

	// The kubeconfig context is the plain cluster name.
	kubeconfig, err := os.ReadFile(os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	assert.Contains(t, string(kubeconfig), "current-context: demo")
}

func TestCreateCluster_FailedApplyKeepsWorkspace(t *testing.T) {
	p, _, registry := newTestProvider(t, errors.New("quota exceeded"))

	_, err := p.CreateCluster(context.Background(), eksConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "re-run create to resume")

	rec, err := registry.Get("demo")
	require.NoError(t, err)
	assert.Equal(t, tfengine.StatusFailed, rec.Status)
}

func TestCreateCluster_RequiresRegion(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)

	config := eksConfig("demo")
	config.Cloud = nil
	_, err := p.CreateCluster(context.Background(), config)

	var invalid models.ErrInvalidClusterConfig
	require.ErrorAs(t, err, &invalid)
	assert.Empty(t, *calls, "terraform must not run without a region")
}

func TestCreateCluster_CredentialPreflightFailsFast(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "InvalidClientTokenId")
	p.executor = mock

	_, err := p.CreateCluster(context.Background(), eksConfig("demo"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot authenticate")
	assert.Empty(t, *calls, "terraform must not run with broken credentials")
}

func TestDeleteCluster_DestroysAndRemovesWorkspace(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), eksConfig("demo"))
	require.NoError(t, err)

	require.NoError(t, p.DeleteCluster(context.Background(), "demo", models.ClusterTypeEKS, false))
	assert.Contains(t, *calls, "destroy")

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestDeleteCluster_MissingIsNotFound(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	err := p.DeleteCluster(context.Background(), "ghost", models.ClusterTypeEKS, false)
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestStartCluster_Unsupported(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	err := p.StartCluster(context.Background(), "demo", models.ClusterTypeEKS)
	assert.ErrorContains(t, err, "not supported")
}

func TestListAndDetect(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	_, err := p.CreateCluster(context.Background(), eksConfig("demo"))
	require.NoError(t, err)

	clusters, err := p.ListClusters(context.Background())
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	assert.Equal(t, models.ClusterTypeEKS, clusters[0].Type)
	assert.Equal(t, "Ready", clusters[0].Status)

	clusterType, err := p.DetectClusterType(context.Background(), "demo")
	require.NoError(t, err)
	assert.Equal(t, models.ClusterTypeEKS, clusterType)

	_, err = p.DetectClusterType(context.Background(), "ghost")
	assert.Error(t, err)
}

func TestGetKubeconfig_RendersExecAuth(t *testing.T) {
	p, _, _ := newTestProvider(t, nil)
	config := eksConfig("demo")
	config.Cloud.Profile = "staging"
	_, err := p.CreateCluster(context.Background(), config)
	require.NoError(t, err)

	kubeconfig, err := p.GetKubeconfig(context.Background(), "demo", models.ClusterTypeEKS)
	require.NoError(t, err)
	assert.Contains(t, kubeconfig, "eks")
	assert.Contains(t, kubeconfig, "get-token")
	assert.Contains(t, kubeconfig, "--profile")
	assert.Contains(t, kubeconfig, "staging")
}

func TestTfvarsFor_VersionMapping(t *testing.T) {
	base := eksConfig("demo")

	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"latest", "", false},
		{"1.33", "1.33", false},
		{"v1.33", "1.33", false},
		{"v1.31.5-k3s1", "", true}, // k3s-style versions are not EKS versions
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
	assert.Contains(t, tf, `source  = "terraform-aws-modules/eks/aws"`)
	assert.Contains(t, tf, `version = "~> 21.0"`)
	assert.Contains(t, tf, `source  = "terraform-aws-modules/vpc/aws"`)
	assert.Contains(t, tf, `version = "~> 6.0"`)
	assert.Contains(t, tf, "enable_cluster_creator_admin_permissions = true")
}

func TestPlanCluster_NewClusterDoesNotRegister(t *testing.T) {
	p, calls, registry := newTestProvider(t, nil)

	summary, err := p.PlanCluster(context.Background(), eksConfig("demo"))
	require.NoError(t, err)
	assert.True(t, summary.HasChanges())
	assert.Equal(t, []string{"init", "plan", "show"}, *calls)

	_, err = registry.Get("demo")
	var notFound models.ErrClusterNotFound
	assert.ErrorAs(t, err, &notFound, "a plan preview must not register the cluster")
}

func TestCreateCluster_WritesS3Backend(t *testing.T) {
	p, _, registry := newTestProvider(t, nil)
	config := eksConfig("demo")
	config.Cloud.BackendConfig = "s3://my-bucket/clusters/demo"

	_, err := p.CreateCluster(context.Background(), config)
	require.NoError(t, err)

	backend, err := os.ReadFile(filepath.Join(registry.Workspace("demo").TerraformDir(), "backend.tf"))
	require.NoError(t, err)
	assert.Contains(t, string(backend), `backend "s3"`)
	assert.Contains(t, string(backend), `bucket = "my-bucket"`)
	assert.Contains(t, string(backend), `key    = "clusters/demo/terraform.tfstate"`)
	assert.Contains(t, string(backend), `region = "us-east-1"`)
}

func TestCreateCluster_RejectsGCSBackend(t *testing.T) {
	p, calls, _ := newTestProvider(t, nil)
	config := eksConfig("demo")
	config.Cloud.BackendConfig = "gcs://my-bucket/prefix"

	_, err := p.CreateCluster(context.Background(), config)
	var invalid models.ErrInvalidClusterConfig
	require.ErrorAs(t, err, &invalid)
	assert.Contains(t, err.Error(), "must be s3://")
	assert.Empty(t, *calls)
}
