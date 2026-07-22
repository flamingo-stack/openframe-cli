package terraform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRunner records calls and returns canned results.
type fakeRunner struct {
	calls       []string
	initErr     error
	apply       error
	outputs     map[string]tfexec.OutputMeta
	planChanges bool
	plan        *tfjson.Plan
	applyJSON   string // lines streamed into the writer by ApplyJSON
}

func (f *fakeRunner) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	f.calls = append(f.calls, "init")
	return f.initErr
}

func (f *fakeRunner) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	f.calls = append(f.calls, "apply")
	return f.apply
}

func (f *fakeRunner) ApplyJSON(ctx context.Context, w io.Writer, opts ...tfexec.ApplyOption) error {
	f.calls = append(f.calls, "apply")
	if f.applyJSON != "" {
		_, _ = w.Write([]byte(f.applyJSON))
	}
	return f.apply
}

func (f *fakeRunner) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	f.calls = append(f.calls, "destroy")
	return nil
}

func (f *fakeRunner) DestroyJSON(ctx context.Context, w io.Writer, opts ...tfexec.DestroyOption) error {
	f.calls = append(f.calls, "destroy")
	return nil
}

func (f *fakeRunner) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	f.calls = append(f.calls, "plan")
	return f.planChanges, nil
}

func (f *fakeRunner) ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error) {
	f.calls = append(f.calls, "show")
	return f.plan, nil
}

func (f *fakeRunner) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	f.calls = append(f.calls, "output")
	return f.outputs, nil
}

func engineWith(f *fakeRunner) *Engine {
	return NewEngineWithRunner(func(workdir string) (Runner, error) { return f, nil })
}

func TestEngine_LifecycleCalls(t *testing.T) {
	f := &fakeRunner{outputs: map[string]tfexec.OutputMeta{
		"cluster_endpoint": {Value: json.RawMessage(`"https://example.eks"`)},
	}}
	e := engineWith(f)
	ctx := context.Background()

	require.NoError(t, e.Init(ctx, "dir"))
	require.NoError(t, e.Apply(ctx, "dir"))
	changes, err := e.Plan(ctx, t.TempDir())
	require.NoError(t, err)
	assert.False(t, changes.HasChanges(), "planChanges=false must summarize to no changes")
	require.NoError(t, e.Destroy(ctx, "dir"))

	outputs, err := e.Outputs(ctx, "dir")
	require.NoError(t, err)
	endpoint, err := StringOutput(outputs, "cluster_endpoint")
	require.NoError(t, err)
	assert.Equal(t, "https://example.eks", endpoint)

	assert.Equal(t, []string{"init", "apply", "plan", "destroy", "output"}, f.calls)
}

// action builds a tfjson resource change with the given address and actions.
func action(address string, actions ...tfjson.Action) *tfjson.ResourceChange {
	return &tfjson.ResourceChange{Address: address, Change: &tfjson.Change{Actions: actions}}
}

func TestEngine_PlanSummaryCountsAndListsActions(t *testing.T) {
	f := &fakeRunner{
		planChanges: true,
		plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
			action("module.network.google_compute_network.vpc", tfjson.ActionCreate),
			action("module.gke.google_container_cluster.primary", tfjson.ActionCreate),
			action("module.gke.google_container_node_pool.pools", tfjson.ActionUpdate),
			action("google_project_service.required", tfjson.ActionDelete),
			action("module.gke.random_string.suffix", tfjson.ActionDelete, tfjson.ActionCreate), // replace
		}},
	}
	summary, err := engineWith(f).Plan(context.Background(), t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, 3, summary.Add)
	assert.Equal(t, 1, summary.Change)
	assert.Equal(t, 2, summary.Destroy)
	assert.True(t, summary.HasChanges())
	// The per-resource listing preserves plan order and diff notation.
	assert.Equal(t, []PlanChange{
		{Action: "+", Address: "module.network.google_compute_network.vpc"},
		{Action: "+", Address: "module.gke.google_container_cluster.primary"},
		{Action: "~", Address: "module.gke.google_container_node_pool.pools"},
		{Action: "-", Address: "google_project_service.required"},
		{Action: "-/+", Address: "module.gke.random_string.suffix"},
	}, summary.Changes)
	assert.Contains(t, f.calls, "show")
}

func TestProgressLine(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		verbose bool
		want    string
		ok      bool
	}{
		{"planned change shown", `{"@message":"module.gke.google_container_cluster.primary: Plan to create","type":"planned_change"}`, false, "module.gke.google_container_cluster.primary: Plan to create", true},
		{"apply start shown", `{"@message":"aws_eks_cluster.this: Creating...","type":"apply_start"}`, false, "aws_eks_cluster.this: Creating...", true},
		{"apply complete shown", `{"@message":"aws_eks_cluster.this: Creation complete after 9m2s","type":"apply_complete"}`, false, "aws_eks_cluster.this: Creation complete after 9m2s", true},
		{"change summary shown", `{"@message":"Apply complete! Resources: 47 added.","type":"change_summary"}`, false, "Apply complete! Resources: 47 added.", true},
		{"progress ticks dropped", `{"@message":"still creating...","type":"apply_progress"}`, false, "", false},
		{"refresh noise dropped", `{"@message":"Refreshing state...","type":"refresh_start"}`, false, "", false},
		{"error diagnostics shown", `{"@level":"error","@message":"Error: quota exceeded","type":"diagnostic"}`, false, "Error: quota exceeded", true},
		{"warning diagnostics dropped", `{"@level":"warn","@message":"deprecation","type":"diagnostic"}`, false, "", false},
		{"verbose forwards everything", `{"@message":"still creating...","type":"apply_progress"}`, true, "still creating...", true},
		{"garbage dropped", `not json`, false, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := progressLine([]byte(tc.line), tc.verbose)
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestProgressWriter_HandlesPartialLines(t *testing.T) {
	w := newProgressWriter(false).(*progressWriter)
	line := `{"@message":"done","type":"apply_complete"}` + "\n"
	half := len(line) / 2

	n, err := w.Write([]byte(line[:half]))
	require.NoError(t, err)
	assert.Equal(t, half, n)
	n, err = w.Write([]byte(line[half:]))
	require.NoError(t, err)
	assert.Equal(t, len(line)-half, n)
	assert.Zero(t, w.buf.Len(), "a completed line must be fully consumed")
}

func TestEngine_WrapsErrors(t *testing.T) {
	f := &fakeRunner{initErr: errors.New("boom")}
	err := engineWith(f).Init(context.Background(), "dir")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "terraform init failed")
}

func TestStringOutput_Missing(t *testing.T) {
	_, err := StringOutput(map[string]json.RawMessage{}, "nope")
	assert.ErrorContains(t, err, `terraform output "nope" missing`)
}

func TestStringOutput_WrongType(t *testing.T) {
	_, err := StringOutput(map[string]json.RawMessage{"n": json.RawMessage(`42`)}, "n")
	assert.ErrorContains(t, err, "is not a string")
}
