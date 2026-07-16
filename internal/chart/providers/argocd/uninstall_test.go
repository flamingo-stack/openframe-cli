package argocd

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func newArgoApp(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata":   map[string]interface{}{"name": name, "namespace": "argocd"},
	}}
}

func TestManager_DeleteApplications(t *testing.T) {
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"},
		newArgoApp("app-of-apps"), newArgoApp("child-1"), newArgoApp("child-2"),
	)
	m := &Manager{dynamicClient: dc}

	n, err := m.DeleteApplications(context.Background())
	if err != nil {
		t.Fatalf("DeleteApplications: %v", err)
	}
	if n != 3 {
		t.Fatalf("deleted = %d, want 3", n)
	}

	list, err := dc.Resource(applicationGVR).Namespace("argocd").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("expected all applications deleted, %d remain", len(list.Items))
	}
}

func TestManager_DeleteApplications_Empty(t *testing.T) {
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"},
	)
	m := &Manager{dynamicClient: dc}

	n, err := m.DeleteApplications(context.Background())
	if err != nil {
		t.Fatalf("DeleteApplications on empty: %v", err)
	}
	if n != 0 {
		t.Fatalf("deleted = %d, want 0", n)
	}
}

func TestManager_RemoveApplicationFinalizers(t *testing.T) {
	stuck := newArgoApp("app-of-apps")
	stuck.SetFinalizers([]string{"resources-finalizer.argocd.argoproj.io"})
	clean := newArgoApp("child-no-finalizer")

	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"},
		stuck, clean,
	)
	m := &Manager{dynamicClient: dc}

	n, err := m.RemoveApplicationFinalizers(context.Background())
	if err != nil {
		t.Fatalf("RemoveApplicationFinalizers: %v", err)
	}
	// Only the one that actually had finalizers is patched/counted.
	if n != 1 {
		t.Fatalf("cleared = %d, want 1", n)
	}

	got, err := dc.Resource(applicationGVR).Namespace("argocd").Get(context.Background(), "app-of-apps", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get after patch: %v", err)
	}
	if len(got.GetFinalizers()) != 0 {
		t.Fatalf("finalizers should be cleared, got %v", got.GetFinalizers())
	}
}

func TestManager_RemoveApplicationFinalizers_Empty(t *testing.T) {
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"},
	)
	m := &Manager{dynamicClient: dc}

	n, err := m.RemoveApplicationFinalizers(context.Background())
	if err != nil {
		t.Fatalf("on empty: %v", err)
	}
	if n != 0 {
		t.Fatalf("cleared = %d, want 0", n)
	}
}

func TestManager_DeleteNamespace(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "argocd"}}
	m := &Manager{kubeClient: fake.NewSimpleClientset(ns)}

	if err := m.DeleteNamespace(context.Background(), "argocd"); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
	if _, err := m.kubeClient.CoreV1().Namespaces().Get(context.Background(), "argocd", metav1.GetOptions{}); err == nil {
		t.Fatal("expected namespace to be gone")
	}
}

func TestManager_DeleteNamespace_MissingIsOK(t *testing.T) {
	m := &Manager{kubeClient: fake.NewSimpleClientset()}
	if err := m.DeleteNamespace(context.Background(), "argocd"); err != nil {
		t.Fatalf("deleting a missing namespace must be a no-op, got %v", err)
	}
}
