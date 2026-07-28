package gke

import (
	"context"
	"strings"
	"time"

	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// appNamespaces are the namespaces OpenFrame installs into. They are deleted
// (argocd first, so its controller stops re-syncing, then openframe) before the
// infra is destroyed, so the PVCs living in them are removed and the GKE CSI
// driver deletes their backing Persistent Disks while the cluster is still
// alive. Order matters — see releaseWorkloadDisks.
var appNamespaces = []string{"argocd", "openframe"}

// diskDrainTimeout bounds how long a delete waits for PVC-backed disks to be
// reclaimed before proceeding to destroy anyway. A live-platform teardown must
// never block indefinitely; anything not drained in time is surfaced by the
// post-destroy sweep instead.
const diskDrainTimeout = 4 * time.Minute

// releaseWorkloadDisks deletes the OpenFrame app namespaces on the cluster and
// waits (bounded) for their PVC-backed PersistentVolumes to drain, so the GKE
// CSI driver deletes the underlying Persistent Disks BEFORE terraform destroys
// the node pool. Without this, PVC-provisioned disks — which live outside the
// terraform state — detach and survive as billable orphans.
//
// Entirely best-effort: an unreachable cluster, missing namespaces, RBAC
// denial, or a drain timeout are all logged and ignored. The infra destroy the
// user asked for must never be blocked by teardown, and the post-destroy sweep
// reports anything this misses.
func releaseWorkloadDisks(ctx context.Context, rec tfengine.Record) {
	restCfg, err := restConfigFor(rec)
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy disk release (no rest config): %v\n", err)
		return
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy disk release (no kube client): %v\n", err)
		return
	}

	var deleting []string
	for _, ns := range appNamespaces {
		// Short per-call context so an unreachable API server fails fast rather
		// than hanging the whole delete.
		getCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		_, err := client.CoreV1().Namespaces().Get(getCtx, ns, metav1.GetOptions{})
		cancel()
		if err != nil {
			continue // absent or unreachable — nothing to release here
		}
		delCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err = client.CoreV1().Namespaces().Delete(delCtx, ns, metav1.DeleteOptions{})
		cancel()
		if err == nil || k8serrors.IsNotFound(err) {
			deleting = append(deleting, ns)
		}
	}
	if len(deleting) == 0 {
		return // no OpenFrame workloads on the cluster; nothing to drain
	}
	pterm.Info.Printf("Releasing persistent disks (removing %s) before destroy...\n", strings.Join(deleting, ", "))

	_ = wait.PollUntilContextTimeout(ctx, 5*time.Second, diskDrainTimeout, true, func(ctx context.Context) (bool, error) {
		listCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		pvs, err := client.CoreV1().PersistentVolumes().List(listCtx, metav1.ListOptions{})
		if err != nil {
			return false, nil // transient — keep polling until the outer timeout
		}
		return countClaimedPVs(pvs.Items, deleting) == 0, nil
	})
}

// countClaimedPVs counts PersistentVolumes still claimed by any of the given
// namespaces. Once a PVC is deleted its PV is released and (reclaimPolicy
// Delete) removed, so this reaching zero means the backing disks are gone.
func countClaimedPVs(pvs []corev1.PersistentVolume, namespaces []string) int {
	inScope := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		inScope[ns] = struct{}{}
	}
	var n int
	for _, pv := range pvs {
		ref := pv.Spec.ClaimRef
		if ref == nil {
			continue
		}
		if _, ok := inScope[ref.Namespace]; ok {
			n++
		}
	}
	return n
}

// reportOrphanedDisks lists any Persistent Disks still labeled for this cluster
// after the destroy and prints them with a cleanup command. It is report-only:
// the CLI never deletes cloud disks directly, so a user is told the truth
// ("these survived") instead of a false "fully cleaned up". Best-effort — a
// gcloud hiccup is silently skipped.
func (p *Provider) reportOrphanedDisks(ctx context.Context, project, name string) {
	if project == "" {
		return
	}
	res, err := p.executor.Execute(ctx, "gcloud", "compute", "disks", "list",
		"--project", project,
		"--filter", "labels.goog-k8s-cluster-name="+name,
		"--format=value(name,location,sizeGb)")
	if err != nil || res == nil {
		return
	}
	disks := parseDiskList(res.Stdout)
	if len(disks) == 0 {
		return
	}
	pterm.Warning.Printf("%d Persistent Disk(s) labeled for cluster %q survived the destroy (PVC-provisioned, outside terraform state):\n", len(disks), name)
	for _, d := range disks {
		pterm.Warning.Printf("  - %s\n", d)
	}
	pterm.Warning.Printf("Review and remove them with:\n  gcloud compute disks list --project %s --filter=\"labels.goog-k8s-cluster-name=%s\"\n", project, name)
}

// parseDiskList turns `gcloud compute disks list --format=value(...)` output
// into one entry per non-empty line.
func parseDiskList(out string) []string {
	var disks []string
	for _, line := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			disks = append(disks, s)
		}
	}
	return disks
}
