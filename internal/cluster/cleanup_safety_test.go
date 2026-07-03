package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFilterProtectedNamespaces_NeverIncludesProtected is the I7 regression
// guard: the cleanup namespace list must never include a protected/system
// namespace, even if one is added to the raw list by mistake. cleanup now
// deletes namespaces via client-go through exactly this filtered list.
func TestFilterProtectedNamespaces_NeverIncludesProtected(t *testing.T) {
	// A raw list deliberately tainted with every protected namespace.
	raw := []string{"argocd", "kube-system", "openframe", "kube-public", "kube-node-lease", "default", "my-app"}

	got := filterProtectedNamespaces(raw)

	for _, protected := range []string{"kube-system", "kube-public", "kube-node-lease", "default"} {
		assert.NotContainsf(t, got, protected, "protected namespace %q must be filtered out", protected)
	}
	// Non-protected namespaces survive.
	for _, ns := range []string{"argocd", "openframe", "my-app"} {
		assert.Containsf(t, got, ns, "non-protected namespace %q must survive", ns)
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
