package argocd

import (
	"context"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// appObj builds an ArgoCD Application unstructured object with the given health
// and sync status.
func appObj(name, health, sync string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata":   map[string]interface{}{"name": name, "namespace": ArgoCDNamespace},
		"status": map[string]interface{}{
			"health": map[string]interface{}{"status": health},
			"sync":   map[string]interface{}{"status": sync},
		},
	}}
}

func fakeManager(objs ...*unstructured.Unstructured) *Manager {
	items := make([]runtime.Object, len(objs))
	for i, o := range objs {
		items[i] = o
	}
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"},
		items...,
	)
	// clientsInitialized=true so parseApplications skips real client init and uses
	// the fake dynamic client. syncWait is tiny so RefreshAndSync's waits (which
	// never complete against the controllerless fake) don't slow the tests.
	return &Manager{dynamicClient: dc, clientsInitialized: true, syncWait: 10 * time.Millisecond}
}

// TestParseApplications_UsesDynamicClient proves the kubectl→client-go migration:
// applications are listed through the dynamic client, no kubectl involved.
func TestParseApplications_UsesDynamicClient(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(
		appObj("core-api", ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObj("nats", ArgoCDHealthProgressing, ArgoCDSyncOutOfSync),
	)

	apps, err := m.parseApplications(context.Background(), false)
	if err != nil {
		t.Fatalf("parseApplications: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("got %d apps, want 2: %+v", len(apps), apps)
	}
	byName := map[string]Application{apps[0].Name: apps[0], apps[1].Name: apps[1]}
	if byName["core-api"].Health != ArgoCDHealthHealthy || byName["core-api"].Sync != ArgoCDSyncSynced {
		t.Errorf("core-api = %+v", byName["core-api"])
	}
	if byName["nats"].Sync != ArgoCDSyncOutOfSync {
		t.Errorf("nats = %+v", byName["nats"])
	}
}

func TestParseApplications_Empty(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	apps, err := fakeManager().parseApplications(context.Background(), false)
	if err != nil {
		t.Fatalf("parseApplications on empty: %v", err)
	}
	if len(apps) != 0 {
		t.Fatalf("want no apps, got %d", len(apps))
	}
}

// TestGetTotalExpectedApplications_CountsViaDynamicClient: Method 2 (list all
// apps except the root app-of-apps) via the dynamic client.
func TestGetTotalExpectedApplications_CountsViaDynamicClient(t *testing.T) {
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObj("child-1", ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObj("child-2", ArgoCDHealthProgressing, ArgoCDSyncOutOfSync),
	)
	got := m.getTotalExpectedApplications(context.Background(), config.ChartInstallConfig{})
	if got != 2 {
		t.Fatalf("expected 2 children (%s excluded), got %d", AppOfAppsName, got)
	}
}

func TestGetTotalExpectedApplications_UnknownWhenNoClient(t *testing.T) {
	// No dynamic client and not initialized → best-effort returns 0 (unknown),
	// the caller discovers the count while polling.
	m := &Manager{clientsInitialized: true}
	if got := m.getTotalExpectedApplications(context.Background(), config.ChartInstallConfig{}); got != 0 {
		t.Fatalf("want 0 (unknown), got %d", got)
	}
}
