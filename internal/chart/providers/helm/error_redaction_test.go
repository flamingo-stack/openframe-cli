package helm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// TestInstallArgoCD_ErrorDoesNotLeakStdout is the regression guard for the
// stdout fallback in the failure path: when helm fails with EMPTY stderr (the
// Windows/WSL case, where stderr is folded into stdout via 2>&1), the error
// embeds result.Stdout. Helm's rendered output can echo values — including the
// docker registry password — so it must be redacted before it reaches the
// user-facing error. Stderr is already redacted by the executor at population;
// stdout is not, because callers parse it.
func TestInstallArgoCD_ErrorDoesNotLeakStdout(t *testing.T) {
	const secret = "dckr_pat_leakedRegistryPassword"
	redact.RegisterSecret(secret)
	t.Cleanup(redact.ClearSecrets)

	mock := executor.NewMockCommandExecutor()
	// helm install fails, reporting only on stdout (stderr empty) — the exact
	// shape the stdout fallback exists for.
	mock.SetResponse("upgrade", &executor.CommandResult{
		ExitCode: 1,
		Stdout:   "Error: values rendered:\n  registry.docker.password: " + secret + "\n",
		Stderr:   "",
		Duration: time.Millisecond,
	})
	m, err := NewHelmManager(mock, nil, false)
	require.NoError(t, err)
	// Pre-create an Active argocd namespace so the pre-install ensure/reachability
	// steps pass immediately and the test reaches the helm failure path.
	m.kubeClient = k8sfake.NewSimpleClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "argocd"},
		Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	})

	err = m.InstallArgoCDWithProgress(context.Background(), config.ChartInstallConfig{
		NonInteractive: true,
		KubeContext:    "k3d-test",
	})
	require.Error(t, err, "the install must fail")

	assert.NotContainsf(t, err.Error(), secret,
		"the error leaks a registered secret from helm stdout:\n%s", err.Error())
	assert.Truef(t, strings.Contains(err.Error(), "***"),
		"expected the redaction marker in the error, got:\n%s", err.Error())
}
