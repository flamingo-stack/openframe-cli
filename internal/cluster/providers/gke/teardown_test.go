package gke

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func pv(ns string) corev1.PersistentVolume {
	p := corev1.PersistentVolume{}
	if ns != "" {
		p.Spec.ClaimRef = &corev1.ObjectReference{Namespace: ns}
	}
	return p
}

func TestCountClaimedPVs(t *testing.T) {
	pvs := []corev1.PersistentVolume{
		pv("openframe"),   // in scope
		pv("openframe"),   // in scope
		pv("argocd"),      // in scope
		pv("kube-system"), // out of scope — must not be counted
		pv(""),            // unbound (no claimRef) — must not be counted
	}
	if got := countClaimedPVs(pvs, []string{"argocd", "openframe"}); got != 3 {
		t.Fatalf("countClaimedPVs = %d, want 3 (only app-namespace claims)", got)
	}
	if got := countClaimedPVs(nil, []string{"openframe"}); got != 0 {
		t.Fatalf("countClaimedPVs(nil) = %d, want 0", got)
	}
}

func TestParseDiskList(t *testing.T) {
	// gcloud value() output: one disk per line, blank lines when there are none.
	out := "pvc-abc\tus-central1-a\t20\npvc-def\tus-central1-b\t8\n\n"
	got := parseDiskList(out)
	if len(got) != 2 {
		t.Fatalf("parseDiskList returned %d entries, want 2: %v", len(got), got)
	}
	if parseDiskList("") != nil || len(parseDiskList("\n \n")) != 0 {
		t.Fatal("empty/whitespace gcloud output must yield no disks")
	}
}
