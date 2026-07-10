package helm

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// TestInstallArgoCD_DryRunSkipsVerification is the regression guard for the
// dry-run failure the e2e caught the moment `--context ... --dry-run` became
// reachable (N2 fix): helm runs with --dry-run=client and the dry-run executor
// suppresses real calls, so post-install release verification ("helm list
// returned empty") and deployment waits are guaranteed to fail — they must be
// skipped entirely in dry-run mode.
func TestInstallArgoCD_DryRunSkipsVerification(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	m, err := NewHelmManager(mock, nil, false)
	require.NoError(t, err)
	// The pre-install reachability check stays active in dry-run (validating
	// the target is part of dry-run's value); satisfy it with a fake clientset
	// (a NotFound answer still proves the API server responded).
	m.kubeClient = k8sfake.NewSimpleClientset()

	cfg := config.ChartInstallConfig{
		DryRun:         true,
		NonInteractive: true,
		KubeContext:    "k3d-test",
	}
	require.NoError(t, m.InstallArgoCDWithProgress(context.Background(), cfg),
		"dry-run install must succeed without a live cluster")

	// The repo add/update + install commands run (through the dry-run-aware
	// executor), but nothing may try to LIST the release afterwards — that is
	// the verification step dry-run can never satisfy.
	for _, rc := range mock.Commands() {
		if rc.Name == "helm" && len(rc.Args) > 0 && rc.Args[0] == "list" {
			t.Fatalf("dry-run must not verify the release via helm list: %v", rc.Args)
		}
	}
	var sawInstall bool
	for _, rc := range mock.Commands() {
		if rc.Name == "helm" && len(rc.Args) > 0 && rc.Args[0] == "upgrade" {
			sawInstall = true
			assert.Contains(t, strings.Join(rc.Args, " "), "--dry-run=client")
		}
	}
	assert.True(t, sawInstall, "the helm install itself must still be issued")
}
