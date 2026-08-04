package argocd

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func labeledCrashPod(name, ns, instance string) *corev1.Pod {
	p := waitingPod(name, "main", "CrashLoopBackOff", "img", 5)
	p.Namespace = ns
	if instance != "" {
		p.Labels = map[string]string{"app.kubernetes.io/instance": instance}
	}
	return &p
}

func labeledHealthyPod(name, ns, instance string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: name, Namespace: ns,
		Labels: map[string]string{"app.kubernetes.io/instance": instance},
	}}
}

// A leftover CrashLooping pod of a NEIGHBOURING app in the shared namespace
// must not mark THIS app as terminally stuck — shared destination namespaces
// are the norm for an app-of-apps platform, and this false positive aborted a
// whole install with "retrying or waiting cannot fix it".
func TestDiagnoseFailingApps_NeighbourPodDoesNotAbort(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset(
		labeledHealthyPod("tenant-api", "tenant", "tenant"),
		labeledCrashPod("old-migration", "tenant", "other-app"),
	)}
	diag, stuck := m.diagnoseFailingApps(context.Background(),
		[]Application{{Name: "tenant", Namespace: "tenant"}})
	if len(stuck) != 0 {
		t.Fatalf("a neighbour's terminal pod must not mark the app stuck, got %v", stuck)
	}
	_ = diag // the diagnostic text itself may mention whatever it saw
}

// An app whose OWN (instance-labeled) pod is terminally stuck is reported.
func TestDiagnoseFailingApps_OwnTerminalPodAborts(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset(
		labeledCrashPod("tenant-stream", "tenant", "tenant"),
	)}
	diag, stuck := m.diagnoseFailingApps(context.Background(),
		[]Application{{Name: "tenant", Namespace: "tenant"}})
	if len(stuck) != 1 || stuck[0].Name != "tenant" {
		t.Fatalf("the app's own terminal pod must mark it stuck, got %v", stuck)
	}
	if !strings.Contains(diag, "CrashLoopBackOff") {
		t.Fatalf("diagnostic must name the failure, got:\n%s", diag)
	}
}

// A CrashLooping pod caught Running between restarts (the real-world timeout
// diagnostic: 8 restarts, healthy-looking current state) must still get a pod
// line in the diagnostic and mark its own app stuck — previously it was
// skipped entirely and only namespace events were printed.
func TestDiagnoseFailingApps_RunningBetweenRestartsIsReported(t *testing.T) {
	p := runningCrashLoopPod("openframe-management-0", "management", "img", 8, 1)
	p.Namespace = "tenant"
	p.Labels = map[string]string{"app.kubernetes.io/instance": "tenant"}
	m := &Manager{kubeClient: fake.NewSimpleClientset(&p)}
	diag, stuck := m.diagnoseFailingApps(context.Background(),
		[]Application{{Name: "tenant", Namespace: "tenant"}})
	if len(stuck) != 1 || stuck[0].Name != "tenant" {
		t.Fatalf("a running-between-restarts crash loop must mark the app stuck, got %v", stuck)
	}
	if !strings.Contains(diag, "openframe-management-0") || !strings.Contains(diag, "CrashLoop (running between restarts)") {
		t.Fatalf("diagnostic must name the pod and the crash loop, got:\n%s", diag)
	}
}

// Unlabeled workloads (no tracking labels at all) fall back to the namespace
// listing for the human diagnostic, but never drive a fail-fast: the failure
// cannot be attributed to the candidate app.
func TestDiagnoseFailingApps_UnattributableIsDiagnosedButNotStuck(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset(
		labeledCrashPod("mystery-pod", "tenant", ""),
	)}
	diag, stuck := m.diagnoseFailingApps(context.Background(),
		[]Application{{Name: "tenant", Namespace: "tenant"}})
	if len(stuck) != 0 {
		t.Fatalf("an unattributable pod must not mark the app stuck, got %v", stuck)
	}
	if !strings.Contains(diag, "mystery-pod") {
		t.Fatalf("the namespace fallback must still describe the failure for humans, got:\n%s", diag)
	}
}
