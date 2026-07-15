package argocd

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func repoServerPod(name string, started time.Time) *corev1.Pod {
	t := metav1.NewTime(started)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ArgoCDNamespace,
			Labels:    map[string]string{"app.kubernetes.io/name": "argocd-repo-server"},
		},
		Status: corev1.PodStatus{StartTime: &t},
	}
}

func TestRepoServerAge_NoClient(t *testing.T) {
	m := &Manager{}
	if _, ok := m.repoServerAge(context.Background()); ok {
		t.Fatal("expected ok=false without a client")
	}
}

func TestRepoServerAge_NoPods(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset()}
	if _, ok := m.repoServerAge(context.Background()); ok {
		t.Fatal("expected ok=false with no repo-server pods")
	}
}

func TestRepoServerAge_YoungestPodWins(t *testing.T) {
	old := repoServerPod("repo-old", time.Now().Add(-30*time.Minute))
	young := repoServerPod("repo-young", time.Now().Add(-1*time.Minute))
	m := &Manager{kubeClient: fake.NewSimpleClientset(old, young)}

	age, ok := m.repoServerAge(context.Background())
	if !ok {
		t.Fatal("expected ok=true")
	}
	if age < 30*time.Second || age > 2*time.Minute {
		t.Fatalf("expected age of the youngest pod (~1m), got %s", age)
	}
	if age >= repoServerColdStartGrace {
		t.Fatalf("a 1-minute-old repo-server must be within the cold-start grace, got %s", age)
	}
}

func TestRepoServerAge_NotYetStarted(t *testing.T) {
	pod := repoServerPod("repo-pending", time.Now())
	pod.Status.StartTime = nil
	m := &Manager{kubeClient: fake.NewSimpleClientset(pod)}

	age, ok := m.repoServerAge(context.Background())
	if !ok || age != 0 {
		t.Fatalf("a pod without StartTime should report age 0, ok=true; got %s, %v", age, ok)
	}
}
