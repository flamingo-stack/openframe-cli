package terraform

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRunner records calls and returns canned results.
type fakeRunner struct {
	calls   []string
	initErr error
	apply   error
	outputs map[string]tfexec.OutputMeta
}

func (f *fakeRunner) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	f.calls = append(f.calls, "init")
	return f.initErr
}

func (f *fakeRunner) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	f.calls = append(f.calls, "apply")
	return f.apply
}

func (f *fakeRunner) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	f.calls = append(f.calls, "destroy")
	return nil
}

func (f *fakeRunner) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	f.calls = append(f.calls, "plan")
	return true, nil
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
	changes, err := e.Plan(ctx, "dir")
	require.NoError(t, err)
	assert.True(t, changes)
	require.NoError(t, e.Destroy(ctx, "dir"))

	outputs, err := e.Outputs(ctx, "dir")
	require.NoError(t, err)
	endpoint, err := StringOutput(outputs, "cluster_endpoint")
	require.NoError(t, err)
	assert.Equal(t, "https://example.eks", endpoint)

	assert.Equal(t, []string{"init", "apply", "plan", "destroy", "output"}, f.calls)
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
