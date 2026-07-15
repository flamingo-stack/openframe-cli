package gke

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/hashicorp/terraform-exec/tfexec"
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

func (f *fakeRunner) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	*f.calls = append(*f.calls, "destroy")
	return nil
}

func (f *fakeRunner) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	*f.calls = append(*f.calls, "plan")
	return true, nil
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
	assert.Equal(t, []string{"init", "apply", "output"}, *calls)
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
	assert.Contains(t, tf, `source  = "terraform-google-modules/kubernetes-engine/google"`)
	assert.Contains(t, tf, `version = "~> 44.0"`)
	assert.Contains(t, tf, `source  = "terraform-google-modules/network/google"`)
	assert.Contains(t, tf, `version = "~> 18.0"`)
	assert.Contains(t, tf, "deletion_protection = false")
}
