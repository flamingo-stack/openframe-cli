package argocd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

// TestStallTracker_OnlyHealthyOutOfSyncQualify: only healthy-but-OutOfSync apps
// can be stalled stragglers — an app with health problems must not be
// auto-synced (it would mask a real failure).
func TestStallTracker_OnlyHealthyOutOfSyncQualify(t *testing.T) {
	apps := []Application{
		{Name: "mongodb", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync},
		{Name: "ready", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced},
		{Name: "broken", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncOutOfSync},
		{Name: "zoo", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync},
	}
	s := newStallTracker()
	t0 := time.Unix(0, 0)
	s.observe(apps, t0)

	// Before stallAfter elapses, nothing is stalled.
	if got := s.stalledStragglers(apps, t0.Add(stallAfter-time.Second)); len(got) != 0 {
		t.Errorf("nothing may be stalled before stallAfter, got %v", got)
	}
	// After it, only the healthy+OutOfSync apps qualify.
	got := s.stalledStragglers(apps, t0.Add(stallAfter))
	want := []string{"mongodb", "zoo"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("stalled stragglers = %v, want %v", got, want)
	}
}

// TestStallTracker_NoisyNeighbourDoesNotResetStuckApp is the V5 regression: a
// genuinely stuck app (Healthy+OutOfSync, never changing) must be detected even
// while a neighbour oscillates Missing<->OutOfSync every tick. The old global
// fingerprint reset its single timer on every neighbour transition, so the
// stuck app never accrued stall time and the wait rode to its full timeout.
func TestStallTracker_NoisyNeighbourDoesNotResetStuckApp(t *testing.T) {
	s := newStallTracker()
	base := time.Unix(0, 0)

	// The neighbour flaps on every tick; the stuck app never changes. Tick every
	// 10s across more than stallAfter.
	flapStates := []struct{ health, sync string }{
		{ArgoCDHealthMissing, ArgoCDSyncOutOfSync},
		{ArgoCDHealthProgressing, ArgoCDSyncOutOfSync},
	}
	var last []string
	for i := 0; i*10 <= int((stallAfter + 10*time.Second).Seconds()); i++ {
		now := base.Add(time.Duration(i*10) * time.Second)
		flap := flapStates[i%len(flapStates)]
		apps := []Application{
			{Name: "argocd-apps", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync}, // stone-cold stuck
			{Name: "ingress-nginx", Health: flap.health, Sync: flap.sync},                 // noisy
		}
		s.observe(apps, now)
		last = s.stalledStragglers(apps, now)
	}

	// The stuck app must be reported despite the neighbour's constant churn.
	found := false
	for _, n := range last {
		if n == "argocd-apps" {
			found = true
		}
	}
	if !found {
		t.Errorf("the stuck app must be detected regardless of a flapping neighbour; got %v", last)
	}
}

// TestStallTracker_ResetsOnStateChange: an app that genuinely progresses (its
// own state changes) restarts its clock and is not reported as stalled.
func TestStallTracker_ResetsOnStateChange(t *testing.T) {
	s := newStallTracker()
	t0 := time.Unix(0, 0)

	appsV1 := []Application{{Name: "mongodb", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync}}
	s.observe(appsV1, t0)

	// Just before it would stall, its sync flips (a real transition).
	appsV2 := []Application{{Name: "mongodb", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced}}
	s.observe(appsV2, t0.Add(stallAfter-time.Second))
	// Then back to OutOfSync — the clock restarted at the transition.
	appsV3 := appsV1
	s.observe(appsV3, t0.Add(stallAfter))
	if got := s.stalledStragglers(appsV3, t0.Add(stallAfter+time.Second)); len(got) != 0 {
		t.Errorf("an app that transitioned must not be immediately stalled, got %v", got)
	}
	// But if it now sits still for stallAfter, it stalls.
	if got := s.stalledStragglers(appsV3, t0.Add(2*stallAfter)); len(got) != 1 {
		t.Errorf("after stallAfter of no change it must stall, got %v", got)
	}
}

// TestStallTracker_ForgetsVanishedApps: an app that disappears and later
// reappears starts a fresh clock rather than inheriting a stale one.
func TestStallTracker_ForgetsVanishedApps(t *testing.T) {
	s := newStallTracker()
	t0 := time.Unix(0, 0)
	apps := []Application{{Name: "mongodb", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncOutOfSync}}
	s.observe(apps, t0)

	// It vanishes from the listing.
	s.observe(nil, t0.Add(30*time.Second))
	// It reappears much later — its clock must start now, not at t0.
	s.observe(apps, t0.Add(10*time.Minute))
	if got := s.stalledStragglers(apps, t0.Add(10*time.Minute+stallAfter-time.Second)); len(got) != 0 {
		t.Errorf("a reappearing app must start a fresh clock, got %v", got)
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
