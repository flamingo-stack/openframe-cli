package cluster

import (
	"context"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCleanupHelmReleases_CountsAndReportsPartialFailure (M2.1): a release that
// fails to uninstall must NOT be counted as removed, and the phase must report
// the failure. Cleanup used to return nil unconditionally, so the summary
// printed "Freed up disk space" whether or not anything was freed.
func TestCleanupHelmReleases_CountsAndReportsPartialFailure(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   `[{"name":"argo-cd","namespace":"argocd"},{"name":"openframe","namespace":"openframe"}]`,
		Duration: time.Millisecond,
	})
	// One of the two uninstalls fails.
	mock.SetResponse("helm uninstall openframe", &executor.CommandResult{
		ExitCode: 1,
		Stderr:   "release: not found",
		Duration: time.Millisecond,
	})
	service := NewClusterService(mock)

	removed, err := service.cleanupHelmReleases(context.Background(), "k3d-test", false, false)

	assert.Equal(t, 1, removed, "only the release that actually uninstalled may be counted")
	require.Error(t, err, "a failed uninstall must be reported, not swallowed")
	assert.Contains(t, err.Error(), "openframe", "the failure must name the release that survived")
}

// TestCleanupHelmReleases_CountsCleanRun is the control: nothing failed, so the
// count matches and no error is reported.
func TestCleanupHelmReleases_CountsCleanRun(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   `[{"name":"argo-cd","namespace":"argocd"},{"name":"openframe","namespace":"openframe"}]`,
		Duration: time.Millisecond,
	})
	service := NewClusterService(mock)

	removed, err := service.cleanupHelmReleases(context.Background(), "k3d-test", false, false)
	require.NoError(t, err)
	assert.Equal(t, 2, removed)
}

// TestCleanupHelmReleases_EmptyClusterRemovesNothing: an empty cluster must
// report zero removals rather than an implied success.
func TestCleanupHelmReleases_EmptyClusterRemovesNothing(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{ExitCode: 0, Stdout: `[]`})
	service := NewClusterService(mock)

	removed, err := service.cleanupHelmReleases(context.Background(), "k3d-test", false, false)
	require.NoError(t, err)
	assert.Zero(t, removed)
}
