package cluster

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
)

// deletesNamespace reports whether any recorded command is a
// `kubectl delete namespace <ns>`.
func deletesNamespace(cmds []executor.RecordedCommand, ns string) bool {
	for _, c := range cmds {
		var sawDelete, sawNamespaceKw, sawNS bool
		for _, a := range c.Args {
			switch a {
			case "delete":
				sawDelete = true
			case "namespace":
				sawNamespaceKw = true
			case ns:
				sawNS = true
			}
		}
		if sawDelete && sawNamespaceKw && sawNS {
			return true
		}
	}
	return false
}

// TestCleanup_NeverDeletesProtectedNamespace is the I7 regression guard: cleanup
// must never issue a `kubectl delete namespace <protected>` — even with force.
func TestCleanup_NeverDeletesProtectedNamespace(t *testing.T) {
	for _, force := range []bool{false, true} {
		mock := executor.NewMockCommandExecutor()
		svc := NewClusterServiceSuppressed(mock)

		err := svc.cleanupKubernetesResources(context.Background(), false, force)
		assert.NoError(t, err)

		cmds := mock.Commands()
		for _, protected := range []string{"kube-system", "kube-public", "kube-node-lease", "default"} {
			assert.Falsef(t, deletesNamespace(cmds, protected),
				"cleanup deleted protected namespace %q (force=%v)", protected, force)
		}
	}
}

func TestIsProtectedNamespace(t *testing.T) {
	for _, ns := range []string{"kube-system", "kube-public", "kube-node-lease", "default"} {
		assert.Truef(t, isProtectedNamespace(ns), "%s must be protected", ns)
	}
	for _, ns := range []string{"argocd", "openframe", "my-app"} {
		assert.Falsef(t, isProtectedNamespace(ns), "%s must not be protected", ns)
	}
}
