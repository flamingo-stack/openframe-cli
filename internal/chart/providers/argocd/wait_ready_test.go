package argocd

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

// argoPartOfPod builds a pod carrying the app.kubernetes.io/part-of=argocd
// label waitForArgoCDReady selects on. ready toggles the PodReady condition.
func argoPartOfPod(name string, ready bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ArgoCDNamespace,
			Labels:    map[string]string{"app.kubernetes.io/part-of": "argocd"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	return pod
}

// applicationsCRD is the CRD the readiness gate polls for before looking at
// pods at all.
func applicationsCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "applications.argoproj.io"},
	}
}

// readyWaitManager wires fake clients and tiny budgets so the full
// waitForArgoCDReady gate (CRD, pod existence, pod readiness) runs in
// milliseconds. clientsInitialized keeps initKubernetesClients from reaching
// for a real kubeconfig.
func readyWaitManager(crds []runtime.Object, pods ...runtime.Object) *Manager {
	return &Manager{
		kubeClient:         fake.NewSimpleClientset(pods...),
		apiextClient:       apiextfake.NewSimpleClientset(crds...),
		clientsInitialized: true,
		crdWaitRetries:     3,
		podWaitTimeout:     200 * time.Millisecond,
		podWaitInterval:    10 * time.Millisecond,
		podReadyTimeout:    200 * time.Millisecond,
	}
}

func TestWaitForArgoCDReady_ReadyPodsPassTheFullGate(t *testing.T) {
	m := readyWaitManager(
		[]runtime.Object{applicationsCRD()},
		argoPartOfPod("argocd-server", true),
		argoPartOfPod("argocd-repo-server", true),
	)

	start := time.Now()
	if err := m.waitForArgoCDReady(context.Background(), false, false); err != nil {
		t.Fatalf("expected success with CRD present and pods Ready, got: %v", err)
	}
	// Success needs no polling, so it must return well inside the tiny budgets.
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Fatalf("ready gate took %v, expected an immediate return", elapsed)
	}
}

func TestWaitForArgoCDReady_NoPodsTimesOutWithCreationError(t *testing.T) {
	// skipCRDs exercises the CRDs-managed-by-Helm branch; no pods ever appear.
	m := readyWaitManager(nil)

	err := m.waitForArgoCDReady(context.Background(), false, true)
	if err == nil {
		t.Fatal("expected a timeout error when no ArgoCD pods exist")
	}
	if !strings.Contains(err.Error(), "timeout waiting for ArgoCD pods to be created") {
		t.Fatalf("expected the pod-creation timeout error, got: %v", err)
	}
}

func TestWaitForArgoCDReady_PodsNeverReadyTimesOutWithReadyError(t *testing.T) {
	m := readyWaitManager(
		[]runtime.Object{applicationsCRD()},
		argoPartOfPod("argocd-server", false),
	)

	err := m.waitForArgoCDReady(context.Background(), false, false)
	if err == nil {
		t.Fatal("expected a timeout error when pods never become Ready")
	}
	if !strings.Contains(err.Error(), "timeout waiting for ArgoCD pods to be ready") {
		t.Fatalf("expected the pods-not-ready timeout error, got: %v", err)
	}
}

func TestWaitForArgoCDReady_MissingCRDTimesOutWithReleaseHint(t *testing.T) {
	// No CRD registered: the gate must fail before ever looking at pods, and
	// the error must point at the Helm release that installs the CRD.
	m := readyWaitManager(nil, argoPartOfPod("argocd-server", true))

	err := m.waitForArgoCDReady(context.Background(), false, false)
	if err == nil {
		t.Fatal("expected a timeout error when the ArgoCD CRD never appears")
	}
	if !strings.Contains(err.Error(), "timeout waiting for the ArgoCD CRD applications.argoproj.io") {
		t.Fatalf("expected the CRD timeout error, got: %v", err)
	}
}

func TestWaitForArgoCDReady_ContextCancelledMidWaitReturnsPromptly(t *testing.T) {
	// Generous budgets so only cancellation can end the wait early.
	m := readyWaitManager(nil)
	m.podWaitTimeout = 30 * time.Second
	m.podWaitInterval = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := m.waitForArgoCDReady(ctx, false, true)
	if err == nil {
		t.Fatal("expected a cancellation error")
	}
	if !strings.Contains(err.Error(), "operation cancelled") {
		t.Fatalf("expected the cancellation error, got: %v", err)
	}
	// Promptness is the point: the ctx-aware sleeps must not ride out the
	// 30s pod-existence budget.
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("cancellation took %v, expected a prompt return", elapsed)
	}
}
