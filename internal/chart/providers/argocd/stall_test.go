package argocd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

// labeledAppObj is appObj plus ArgoCD's tracking label marking the app as a
// child of the root app-of-apps.
func labeledAppObj(name, health, sync string) *unstructured.Unstructured {
	o := appObj(name, health, sync)
	o.SetLabels(map[string]string{trackingInstanceLabel: AppOfAppsName})
	return o
}

func TestStallFingerprint(t *testing.T) {
	// Order-insensitive: the same set in any order is the same fingerprint.
	a := stallFingerprint(3, []string{"b (Sync: OutOfSync)", "a (Sync: OutOfSync)"})
	b := stallFingerprint(3, []string{"a (Sync: OutOfSync)", "b (Sync: OutOfSync)"})
	if a != b {
		t.Errorf("fingerprint must be order-insensitive: %q vs %q", a, b)
	}
	// Any transition changes it: ready count, membership, or status.
	if a == stallFingerprint(4, []string{"a (Sync: OutOfSync)", "b (Sync: OutOfSync)"}) {
		t.Error("ready-count change must change the fingerprint")
	}
	if a == stallFingerprint(3, []string{"a (Sync: OutOfSync)", "b (Health: Degraded)"}) {
		t.Error("status change must change the fingerprint")
	}
}

// TestOutOfSyncStragglers: only healthy-but-OutOfSync apps qualify — an app
// with health problems must not be auto-synced (it would mask a real failure).
func TestOutOfSyncStragglers(t *testing.T) {
	apps := []Application{
		{Name: "mongodb", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync},
		{Name: "ready", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced},
		{Name: "broken", Health: "Degraded", Sync: ArgoCDSyncOutOfSync},
		{Name: "zoo", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync},
	}
	got := outOfSyncStragglers(apps)
	want := []string{"mongodb", "zoo"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("stragglers = %v, want %v", got, want)
	}
}

// TestSyncChildApplications_PrefersTrackingLabel: only children carrying the
// app-of-apps tracking label are synced; a foreign Application in the argocd
// namespace is left alone (verification report: force-sync used to patch every
// Application, OpenFrame-owned or not).
func TestSyncChildApplications_PrefersTrackingLabel(t *testing.T) {
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		labeledAppObj("child-a", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
		appObj("foreign-app", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync), // no tracking label
	)

	var patched []string
	dc := m.dynamicClient.(interface {
		PrependReactor(verb, resource string, fn k8stesting.ReactionFunc)
	})
	dc.PrependReactor("patch", "applications", func(action k8stesting.Action) (bool, runtime.Object, error) {
		patched = append(patched, action.(k8stesting.PatchAction).GetName())
		return false, nil, nil // fall through to the default reactor
	})

	if err := m.syncChildApplications(context.Background(), false); err != nil {
		t.Fatalf("syncChildApplications: %v", err)
	}
	for _, name := range patched {
		if name == "foreign-app" {
			t.Error("foreign (untracked) application must not be synced when labeled children exist")
		}
	}
	if len(patched) == 0 || patched[0] != "child-a" {
		t.Errorf("labeled child must be synced, patched=%v", patched)
	}
}

// TestSyncChildApplications_AllFailuresSurface is the F8 guard: when not one
// child can be synced, the caller must get an error — previously every patch
// error was discarded and `--sync` "succeeded" into a guaranteed wait timeout.
func TestSyncChildApplications_AllFailuresSurface(t *testing.T) {
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		labeledAppObj("child-a", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
		labeledAppObj("child-b", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
	)
	dc := m.dynamicClient.(interface {
		PrependReactor(verb, resource string, fn k8stesting.ReactionFunc)
	})
	dc.PrependReactor("patch", "applications", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("rbac: denied")
	})

	err := m.syncChildApplications(context.Background(), false)
	if err == nil {
		t.Fatal("total patch failure must surface as an error")
	}
	if !strings.Contains(err.Error(), "rbac: denied") {
		t.Errorf("error should carry the first cause, got: %v", err)
	}
}
