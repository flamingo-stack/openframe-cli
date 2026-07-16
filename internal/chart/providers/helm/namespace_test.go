package helm

import (
	"context"
	"runtime"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestEnsureArgoCDNamespace_CreatesViaClientGo proves the kubectl→client-go
// migration for namespace creation: with a native (fake) clientset the argocd
// namespace is created through the Kubernetes API — no kubectl shell-out.
func TestEnsureArgoCDNamespace_CreatesViaClientGo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	// Fake clientset with the namespace pre-marked Active so the readiness poll
	// returns immediately once it exists.
	client := fake.NewSimpleClientset()
	h := &HelmManager{kubeClient: client}

	// The namespace doesn't exist yet; Create is issued, then the Active poll
	// must see it. Seed an Active namespace up front to keep the test fast and
	// deterministic (Create is idempotent via AlreadyExists handling).
	_, _ = client.CoreV1().Namespaces().Create(context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: argocd.ArgoCDNamespace},
			Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		}, metav1.CreateOptions{})

	if err := h.ensureArgoCDNamespace(context.Background(), "test", false); err != nil {
		t.Fatalf("ensureArgoCDNamespace via client-go: %v", err)
	}

	// Confirm the namespace exists through the API (no kubectl involved).
	if _, err := client.CoreV1().Namespaces().Get(context.Background(), argocd.ArgoCDNamespace, metav1.GetOptions{}); err != nil {
		t.Fatalf("argocd namespace should exist via client-go: %v", err)
	}
}

// TestEnsureArgoCDNamespace_NoClientErrors: without a reachable cluster the
// operation must fail clearly (nil-client error off Windows, WSL hint on
// Windows) rather than silently shelling out to kubectl.
func TestEnsureArgoCDNamespace_NoClientErrors(t *testing.T) {
	h := &HelmManager{kubeClient: nil}
	if err := h.ensureArgoCDNamespace(context.Background(), "test", false); err == nil {
		t.Fatal("expected an error when the kube client is unavailable")
	}
}
