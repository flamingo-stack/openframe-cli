package eks

import (
	"context"
	"testing"

	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppNamespacesToDelete(t *testing.T) {
	all := []string{
		"kube-system", "kube-public", "kube-node-lease", "default", // system — skipped
		"aws-observability", "amazon-cloudwatch", // AWS-managed — skipped
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
	for _, ns := range []string{"kube-system", "kube-public", "default", "aws-observability", "amazon-cloudwatch"} {
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

// Every app-namespace LoadBalancer Service is an ELB/NLB whose ENIs sit in the
// VPC subnets — the drain wait must count exactly those, not ClusterIPs and not
// system-namespace LBs (their controller dies with the cluster anyway).
func TestCountAppLoadBalancers(t *testing.T) {
	svc := func(ns string, typ corev1.ServiceType) corev1.Service {
		return corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns},
			Spec:       corev1.ServiceSpec{Type: typ},
		}
	}
	svcs := []corev1.Service{
		svc("tenant", corev1.ServiceTypeLoadBalancer),      // counted
		svc("datasources", corev1.ServiceTypeLoadBalancer), // counted
		svc("tenant", corev1.ServiceTypeClusterIP),         // not an LB
		svc("kube-system", corev1.ServiceTypeLoadBalancer), // system ns
	}
	if got := countAppLoadBalancers(svcs); got != 2 {
		t.Fatalf("countAppLoadBalancers = %d, want 2", got)
	}
}

func TestParseVolumeIDs(t *testing.T) {
	// `--output text` separates values with tabs/newlines; an empty result may
	// print nothing or the literal "None".
	got := parseVolumeIDs("vol-0abc\tvol-0def\nvol-0123\n")
	if len(got) != 3 || got[0] != "vol-0abc" || got[2] != "vol-0123" {
		t.Fatalf("parseVolumeIDs = %v, want the 3 ids", got)
	}
	if len(parseVolumeIDs("")) != 0 {
		t.Fatal("empty output must yield no volumes")
	}
	if len(parseVolumeIDs("None\n")) != 0 {
		t.Fatal("the literal None must yield no volumes")
	}
}

// The post-destroy sweep must only delete cloud volumes when the caller
// consented to an unattended teardown (--force); otherwise it reports and
// touches nothing.
func TestSweepOrphanedVolumes_ConsentGating(t *testing.T) {
	rec := tfengine.Record{Name: "demo", Region: "us-east-1"}
	listResp := &executor.CommandResult{ExitCode: 0, Stdout: "vol-0abc\n"}

	t.Run("no force: reports, never deletes", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("aws ec2 describe-volumes", listResp)
		p.executor = mock

		p.sweepOrphanedVolumes(context.Background(), rec, false)
		if mock.WasCommandExecuted("aws ec2 delete-volume") {
			t.Fatal("without --force the sweep must NOT delete volumes")
		}
	})

	t.Run("force: deletes the tagged orphan", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("aws ec2 describe-volumes", listResp)
		p.executor = mock

		p.sweepOrphanedVolumes(context.Background(), rec, true)
		if !mock.WasCommandExecuted("aws ec2 delete-volume --volume-id vol-0abc --region us-east-1") {
			t.Fatalf("with --force the sweep must delete the orphan; ran: %v", mock.GetExecutedCommands())
		}
	})

	t.Run("force with profile: the profile reaches both calls", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("aws ec2 describe-volumes", listResp)
		p.executor = mock

		profiled := rec
		profiled.Profile = "staging"
		p.sweepOrphanedVolumes(context.Background(), profiled, true)
		if !mock.WasCommandExecuted("--profile staging") {
			t.Fatalf("the record's profile must scope the sweep; ran: %v", mock.GetExecutedCommands())
		}
	})

	t.Run("empty region is a no-op", func(t *testing.T) {
		p, _, _ := newTestProvider(t, nil)
		mock := executor.NewMockCommandExecutor()
		p.executor = mock
		p.sweepOrphanedVolumes(context.Background(), tfengine.Record{Name: "demo"}, true)
		if mock.GetCommandCount() != 0 {
			t.Fatal("no region → no aws calls")
		}
	})
}
