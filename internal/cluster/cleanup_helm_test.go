package cluster

import (
	"context"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helmArgvsOf returns the argv of every helm invocation recorded by the mock.
func helmArgvsOf(mock *executor.MockCommandExecutor) [][]string {
	var out [][]string
	for _, rc := range mock.Commands() {
		if rc.Name == "helm" {
			out = append(out, rc.Args)
		}
	}
	return out
}

// hasFlagValue reports whether argv contains the flag immediately followed by value.
func hasFlagValue(argv []string, flag, value string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == flag && argv[i+1] == value {
			return true
		}
	}
	return false
}

// TestCleanupHelmReleases_PinsKubeContext is the T0-1 regression guard: every
// helm call issued by cleanup must carry --kube-context for the cluster being
// cleaned. Without the pin, helm acts on the kubeconfig's CURRENT context —
// switching context to a production cluster and running `cluster cleanup`
// would uninstall every release there.
func TestCleanupHelmReleases_PinsKubeContext(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	// Real `helm list --output json` emits a single-line JSON array.
	mock.SetResponse("helm list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   `[{"name":"argo-cd","namespace":"argocd","status":"deployed"},{"name":"openframe","namespace":"openframe","status":"deployed"}]`,
		Duration: time.Millisecond,
	})
	service := NewClusterService(mock)

	err := service.cleanupHelmReleases(context.Background(), "k3d-test-cluster", false, false)
	require.NoError(t, err)

	argvs := helmArgvsOf(mock)
	require.NotEmpty(t, argvs, "cleanup must invoke helm")
	for _, argv := range argvs {
		assert.Truef(t, hasFlagValue(argv, "--kube-context", "k3d-test-cluster"),
			"every helm call must pin --kube-context k3d-test-cluster, got: %v", argv)
	}

	// Both releases are uninstalled, each in its own namespace.
	var uninstalls [][]string
	for _, argv := range argvs {
		if len(argv) > 0 && argv[0] == "uninstall" {
			uninstalls = append(uninstalls, argv)
		}
	}
	require.Len(t, uninstalls, 2, "one uninstall per listed release")
	assert.Equal(t, "argo-cd", uninstalls[0][1])
	assert.True(t, hasFlagValue(uninstalls[0], "--namespace", "argocd"))
	assert.Equal(t, "openframe", uninstalls[1][1])
	assert.True(t, hasFlagValue(uninstalls[1], "--namespace", "openframe"))
}

// TestCleanupHelmReleases_ForceAddsIgnoreNotFound locks the force-mode flag.
func TestCleanupHelmReleases_ForceAddsIgnoreNotFound(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   `[{"name":"argo-cd","namespace":"argocd"}]`,
		Duration: time.Millisecond,
	})
	service := NewClusterService(mock)

	require.NoError(t, service.cleanupHelmReleases(context.Background(), "k3d-x", false, true))

	found := false
	for _, argv := range helmArgvsOf(mock) {
		if len(argv) > 0 && argv[0] == "uninstall" {
			found = true
			assert.Contains(t, argv, "--ignore-not-found")
		}
	}
	assert.True(t, found, "expected an uninstall call")
}

// TestCleanupHelmReleases_EmptyList: nothing to uninstall on "[]" or empty output.
func TestCleanupHelmReleases_EmptyList(t *testing.T) {
	for name, stdout := range map[string]string{"empty-array": "[]", "blank": ""} {
		t.Run(name, func(t *testing.T) {
			mock := executor.NewMockCommandExecutor()
			mock.SetResponse("helm list", &executor.CommandResult{ExitCode: 0, Stdout: stdout, Duration: time.Millisecond})
			service := NewClusterService(mock)

			require.NoError(t, service.cleanupHelmReleases(context.Background(), "k3d-x", false, false))
			for _, argv := range helmArgvsOf(mock) {
				assert.NotEqual(t, "uninstall", argv[0], "no uninstall may run for an empty release list")
			}
		})
	}
}

// TestCleanupHelmReleases_RefusesWithoutContext: a missing kube-context must be
// a hard error, never a fall-through to the current context.
func TestCleanupHelmReleases_RefusesWithoutContext(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	service := NewClusterService(mock)

	err := service.cleanupHelmReleases(context.Background(), "", false, false)
	require.Error(t, err)
	assert.Zero(t, mock.GetCommandCount(), "no command may run without an explicit kube-context")
}

// TestCleanupHelmReleases_GarbageOutputErrors: unparseable helm output must
// surface as an error instead of being half-parsed (the old code split the
// JSON on ":" and produced garbage namespaces).
func TestCleanupHelmReleases_GarbageOutputErrors(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{ExitCode: 0, Stdout: "not json", Duration: time.Millisecond})
	service := NewClusterService(mock)

	err := service.cleanupHelmReleases(context.Background(), "k3d-x", false, false)
	require.Error(t, err)
	for _, argv := range helmArgvsOf(mock) {
		assert.NotEqual(t, "uninstall", argv[0], "no uninstall may run on unparseable output")
	}
}

// TestCleanupCluster_HelmPhasePinsKubeContext exercises the full CleanupCluster
// entry point: whatever context resolution yields, the helm phase must never
// issue a helm call without --kube-context.
func TestCleanupCluster_HelmPhasePinsKubeContext(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("helm list", &executor.CommandResult{ExitCode: 0, Stdout: "[]", Duration: time.Millisecond})
	service := NewClusterService(mock)

	// K8s/Docker phases run against the mock too and are allowed to no-op/fail.
	_ = service.CleanupCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false, false)

	argvs := helmArgvsOf(mock)
	require.NotEmpty(t, argvs, "cleanup must reach the helm phase")
	for _, argv := range argvs {
		assert.Containsf(t, argv, "--kube-context", "helm call without --kube-context: %v", argv)
	}
}
