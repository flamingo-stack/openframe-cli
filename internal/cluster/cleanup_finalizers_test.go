package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingCleaner records the order in which the cleanup phases call it,
// appending to a shared trace so it can be interleaved with the helm calls.
type recordingCleaner struct {
	trace      *[]string
	deleteErr  error
	clearErr   error
	deleted    int
	cleared    int
	deleteCall int
	clearCall  int
}

func (r *recordingCleaner) DeleteApplications(context.Context) (int, error) {
	r.deleteCall++
	*r.trace = append(*r.trace, "delete-applications")
	return r.deleted, r.deleteErr
}

func (r *recordingCleaner) RemoveApplicationFinalizers(context.Context) (int, error) {
	r.clearCall++
	*r.trace = append(*r.trace, "clear-finalizers")
	return r.cleared, r.clearErr
}

// tracingExecutor records helm uninstalls into the same trace as the cleaner.
type tracingExecutor struct {
	*executor.MockCommandExecutor
	trace *[]string
}

func (t *tracingExecutor) Execute(ctx context.Context, name string, args ...string) (*executor.CommandResult, error) {
	if name == "helm" && len(args) > 0 && args[0] == "uninstall" {
		*t.trace = append(*t.trace, "helm-uninstall")
	}
	return t.MockCommandExecutor.Execute(ctx, name, args...)
}

func newTracingExecutor(trace *[]string) *tracingExecutor {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   `[{"name":"argo-cd","namespace":"argocd"}]`,
		Duration: time.Millisecond,
	})
	return &tracingExecutor{MockCommandExecutor: mock, trace: trace}
}

// TestCleanup_ApplicationPhasesBracketTheHelmUninstall locks the ordering the
// whole fix depends on: ArgoCD Applications are deleted while the controller
// still runs (so it cascades workload cleanup), and their resources-finalizer
// is stripped only AFTER the helm uninstall removed the controller — nothing
// else can clear it, so a Terminating CR would otherwise pin the namespace.
func TestCleanup_ApplicationPhasesBracketTheHelmUninstall(t *testing.T) {
	var trace []string
	exec := newTracingExecutor(&trace)
	cleaner := &recordingCleaner{trace: &trace, deleted: 3, cleared: 2}

	service := NewClusterService(exec).WithApplicationCleaner(cleaner)
	_, _ = service.CleanupCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false, false)

	require.GreaterOrEqual(t, len(trace), 3, "trace: %v", trace)
	assert.Equal(t, "delete-applications", trace[0], "applications must be deleted first: %v", trace)

	helmIdx, clearIdx := -1, -1
	for i, step := range trace {
		switch step {
		case "helm-uninstall":
			if helmIdx == -1 {
				helmIdx = i
			}
		case "clear-finalizers":
			clearIdx = i
		}
	}
	require.NotEqual(t, -1, helmIdx, "helm uninstall must run: %v", trace)
	require.NotEqual(t, -1, clearIdx, "finalizers must be cleared: %v", trace)
	assert.Greater(t, clearIdx, helmIdx,
		"finalizers must be stripped AFTER the ArgoCD controller is uninstalled: %v", trace)

	assert.Equal(t, 1, cleaner.deleteCall)
	assert.Equal(t, 1, cleaner.clearCall)
}

// TestCleanup_WithoutCleanerStillRuns: the cleaner is optional — a cluster
// without OpenFrame (or an unreachable one) must still get the helm/namespace/
// docker phases rather than failing.
func TestCleanup_WithoutCleanerStillRuns(t *testing.T) {
	var trace []string
	exec := newTracingExecutor(&trace)

	service := NewClusterService(exec) // no cleaner injected
	_, err := service.CleanupCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false, false)
	require.NoError(t, err)
	assert.Contains(t, trace, "helm-uninstall", "helm phase must still run: %v", trace)
	assert.NotContains(t, trace, "delete-applications")
	assert.NotContains(t, trace, "clear-finalizers")
}

// TestCleanup_CleanerErrorsAreNonFatal: the platform-cleanup phases are
// best-effort — a cluster where ArgoCD was never installed (or the API errors)
// must not abort the rest of the cleanup.
func TestCleanup_CleanerErrorsAreNonFatal(t *testing.T) {
	var trace []string
	exec := newTracingExecutor(&trace)
	cleaner := &recordingCleaner{
		trace:     &trace,
		deleteErr: fmt.Errorf("no argocd CRD"),
		clearErr:  fmt.Errorf("no argocd CRD"),
	}

	service := NewClusterService(exec).WithApplicationCleaner(cleaner)
	_, err := service.CleanupCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false, false)
	require.NoError(t, err,
		"cleaner failures must not fail the cleanup")
	assert.Contains(t, trace, "helm-uninstall", "helm phase must still run after a cleaner error: %v", trace)
	assert.Equal(t, 1, cleaner.clearCall, "the finalizer phase must run even if the delete phase failed")
}
