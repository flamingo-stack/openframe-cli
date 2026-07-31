package gke

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	corev1 "k8s.io/api/core/v1"
)

func TestAppNamespacesToDelete(t *testing.T) {
	all := []string{
		"kube-system", "kube-public", "kube-node-lease", "default", // system — skipped
		"gke-managed-system", "gmp-system", // GKE-managed — skipped
		"tenant", "datasources", "platform", "argocd", // app — kept
	}
	got := appNamespacesToDelete(all)

	// argocd must come first so its controller stops re-syncing before its
	// managed workloads (and their PVCs) are deleted.
	if len(got) == 0 || got[0] != "argocd" {
		t.Fatalf("argocd must be deleted first, got %v", got)
	}
	want := map[string]bool{"argocd": true, "datasources": true, "platform": true, "tenant": true}
	if len(got) != len(want) {
		t.Fatalf("expected exactly the app namespaces, got %v", got)
	}
	for _, ns := range got {
		if !want[ns] {
			t.Fatalf("system namespace %q must never be deleted; got %v", ns, got)
		}
	}
}

func TestIsSystemNamespace(t *testing.T) {
	for _, ns := range []string{"kube-system", "kube-public", "default", "gke-managed-cim", "gmp-public"} {
		if !isSystemNamespace(ns) {
			t.Errorf("%q must be treated as a system namespace", ns)
		}
	}
	for _, ns := range []string{"datasources", "tenant", "argocd", "platform"} {
		if isSystemNamespace(ns) {
			t.Errorf("%q is an app namespace, not system", ns)
		}
	}
}

func TestCountDeletablePVs(t *testing.T) {
	pvs := []corev1.PersistentVolume{
		{Spec: corev1.PersistentVolumeSpec{PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete}},
		{Spec: corev1.PersistentVolumeSpec{PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete}},
		{Spec: corev1.PersistentVolumeSpec{PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain}}, // not counted
	}
	if got := countDeletablePVs(pvs); got != 2 {
		t.Fatalf("countDeletablePVs = %d, want 2 (Retain excluded)", got)
	}
}

func TestParseDisks(t *testing.T) {
	// value() output is TAB-separated with empty columns kept: a regional disk
	// has an empty zone column before its region.
	out := "pvc-abc\tus-central1-a\t\npvc-def\tus-central1-b\t\nregional-disk\t\tus-central1\n\n"
	got := parseDisks(out)
	if len(got) != 3 {
		t.Fatalf("parseDisks returned %d, want 3: %+v", len(got), got)
	}
	if got[0].name != "pvc-abc" || got[0].zone != "us-central1-a" || got[0].region != "" {
		t.Fatalf("first disk = %+v", got[0])
	}
	if got[2].name != "regional-disk" || got[2].zone != "" || got[2].region != "us-central1" {
		t.Fatalf("regional disk must parse with empty zone and its region, got %+v", got[2])
	}
	if len(parseDisks("")) != 0 {
		t.Fatal("empty output must yield no disks")
	}
}

// GKE cluster names are unique per LOCATION, not per project: the sweep must
// never touch a same-named cluster's disks in another region.
func TestDisksInLocation(t *testing.T) {
	all := []disk{
		{name: "mine-zonal", zone: "us-central1-a"},
		{name: "other-zonal", zone: "europe-west1-b"},
		{name: "mine-regional", region: "us-central1"},
		{name: "other-regional", region: "europe-west1"},
	}
	got := disksInLocation(all, "us-central1")
	if len(got) != 2 || got[0].name != "mine-zonal" || got[1].name != "mine-regional" {
		t.Fatalf("expected only us-central1 disks, got %+v", got)
	}
	// A record without a location (predates the field) keeps everything.
	if got := disksInLocation(all, ""); len(got) != 4 {
		t.Fatalf("empty location must keep all disks, got %+v", got)
	}
	// A region name must not prefix-match a lookalike ("us-central12-a").
	lookalike := []disk{{name: "x", zone: "us-central12-a"}}
	if got := disksInLocation(lookalike, "us-central1"); len(got) != 0 {
		t.Fatalf("us-central12-a must not match us-central1, got %+v", got)
	}
}

// The post-destroy sweep must only delete cloud disks when the caller consented
// to an unattended teardown (--force); otherwise it reports and touches nothing.
func TestSweepOrphanedDisks_ConsentGating(t *testing.T) {
	listResp := &executor.CommandResult{ExitCode: 0, Stdout: "pvc-abc\tus-central1-a\n"}

	t.Run("no force: reports, never deletes", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud compute disks list", listResp)
		p.executor = mock

		p.sweepOrphanedDisks(context.Background(), "proj", "demo", "us-central1", false)
		if mock.WasCommandExecuted("gcloud compute disks delete") {
			t.Fatal("without --force the sweep must NOT delete disks")
		}
	})

	t.Run("force: deletes the labeled orphan", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud compute disks list", listResp)
		p.executor = mock

		p.sweepOrphanedDisks(context.Background(), "proj", "demo", "us-central1", true)
		if !mock.WasCommandExecuted("gcloud compute disks delete pvc-abc --zone us-central1-a") {
			t.Fatalf("with --force the sweep must delete the orphan; ran: %v", mock.GetExecutedCommands())
		}
	})

	t.Run("force: never deletes a same-named other-location cluster's disks", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud compute disks list", &executor.CommandResult{
			ExitCode: 0,
			Stdout:   "pvc-mine\tus-central1-a\t\npvc-other\teurope-west1-b\t\n",
		})
		p.executor = mock

		p.sweepOrphanedDisks(context.Background(), "proj", "demo", "us-central1", true)
		if !mock.WasCommandExecuted("gcloud compute disks delete pvc-mine --zone us-central1-a") {
			t.Fatalf("the deleted cluster's own disk must be swept; ran: %v", mock.GetExecutedCommands())
		}
		if mock.WasCommandExecuted("gcloud compute disks delete pvc-other") {
			t.Fatal("a disk of the same-named cluster in another location must never be deleted")
		}
	})

	t.Run("empty project is a no-op", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		p.executor = mock
		p.sweepOrphanedDisks(context.Background(), "", "demo", "us-central1", true)
		if mock.GetCommandCount() != 0 {
			t.Fatal("no project → no gcloud calls")
		}
	})
}
