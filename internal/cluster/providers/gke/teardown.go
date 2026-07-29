package gke

import (
	"context"
	"fmt"
	"strings"
	"time"

	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// systemNamespaces and systemNamespacePrefixes are never deleted during
// teardown. They are the cluster's own control-plane/system namespaces (torn
// down with the cluster anyway) and — critically — kube-system hosts the GKE PD
// CSI controller that must keep running to delete the Persistent Disks as their
// PVCs go away.
var systemNamespaces = map[string]struct{}{
	"default":         {},
	"kube-system":     {},
	"kube-public":     {},
	"kube-node-lease": {},
}

var systemNamespacePrefixes = []string{"kube-", "gke-", "gmp-"}

const (
	// diskDrainTimeout bounds how long a delete waits for PVC-backed disks to be
	// reclaimed before proceeding to destroy anyway. A live-platform teardown
	// must never block indefinitely; anything not drained in time is caught by
	// the post-destroy sweep.
	diskDrainTimeout = 5 * time.Minute
	// kubeCallTimeout bounds each individual API call so an unreachable cluster
	// fails fast instead of hanging the whole delete.
	kubeCallTimeout = 20 * time.Second
)

// isSystemNamespace reports whether ns is a cluster/system namespace that
// teardown must never delete.
func isSystemNamespace(ns string) bool {
	if _, ok := systemNamespaces[ns]; ok {
		return true
	}
	for _, p := range systemNamespacePrefixes {
		if strings.HasPrefix(ns, p) {
			return true
		}
	}
	return false
}

// appNamespacesToDelete returns the application namespaces (everything that is
// not a system namespace), with argocd first. OpenFrame's stateful services
// (Kafka, MongoDB, Cassandra, Pinot, … in the 'datasources' namespace) hold the
// PVCs whose backing disks must be released; deleting by discovery rather than a
// hardcoded list keeps this correct as the platform layout changes. argocd goes
// first so its controller stops re-syncing before the workloads it manages are
// deleted, otherwise self-heal could recreate a StatefulSet (and its PVC)
// mid-teardown.
func appNamespacesToDelete(all []string) []string {
	var argocd []string
	var rest []string
	for _, ns := range all {
		if isSystemNamespace(ns) {
			continue
		}
		if ns == "argocd" {
			argocd = append(argocd, ns)
		} else {
			rest = append(rest, ns)
		}
	}
	return append(argocd, rest...)
}

// countDeletablePVs counts PersistentVolumes whose reclaim policy is Delete.
// These are the volumes whose backing cloud disk the CSI driver removes once
// their PVC is gone, so the release step waits for this to reach zero.
// Retain-policy PVs are excluded on purpose — their disks are meant to survive,
// and the post-destroy sweep reports (never silently drops) them.
func countDeletablePVs(pvs []corev1.PersistentVolume) int {
	var n int
	for _, pv := range pvs {
		if pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			n++
		}
	}
	return n
}

// releaseWorkloadDisks deletes every application namespace on the cluster and
// waits (bounded) for the Delete-reclaim PersistentVolumes to drain, so the GKE
// CSI driver deletes the backing Persistent Disks BEFORE terraform destroys the
// node pool. Those disks are PVC-provisioned and live OUTSIDE the terraform
// state, so without this they detach and survive as billable orphans.
//
// Deleting whole namespaces (rather than draining nodes) also avoids the
// autoscaler racing to reschedule evicted pods onto a fresh node mid-teardown.
//
// Entirely best-effort: an unreachable cluster, RBAC denial, or a drain timeout
// are logged and ignored. The infra destroy the user asked for must never be
// blocked by teardown, and the post-destroy sweep reports/cleans anything this
// misses.
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

	listCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	nsList, err := client.CoreV1().Namespaces().List(listCtx, metav1.ListOptions{})
	cancel()
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy disk release (cluster unreachable): %v\n", err)
		return
	}
	names := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}
	targets := appNamespacesToDelete(names)
	if len(targets) == 0 {
		return // no application namespaces — nothing to release
	}

	pterm.Info.Printf("Releasing persistent disks (removing namespaces: %s) before destroy...\n", strings.Join(targets, ", "))
	for _, ns := range targets {
		delCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		err := client.CoreV1().Namespaces().Delete(delCtx, ns, metav1.DeleteOptions{})
		cancel()
		if err != nil && !k8serrors.IsNotFound(err) {
			pterm.Debug.Printf("namespace %s delete: %v\n", ns, err)
		}
	}

	// Wait for Delete-reclaim PVs to disappear — the signal that the CSI driver
	// has deleted the backing disks. Bounded; on timeout we proceed and let the
	// sweep handle the remainder.
	_ = wait.PollUntilContextTimeout(ctx, 5*time.Second, diskDrainTimeout, true, func(ctx context.Context) (bool, error) {
		pvCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		defer cancel()
		pvs, err := client.CoreV1().PersistentVolumes().List(pvCtx, metav1.ListOptions{})
		if err != nil {
			return false, nil // transient — keep polling until the outer timeout
		}
		return countDeletablePVs(pvs.Items) == 0, nil
	})
}

// disk is one labeled Persistent Disk found by the post-destroy sweep.
type disk struct {
	name string
	zone string // basename, e.g. "us-central1-a"; empty for a regional disk
}

// sweepOrphanedDisks finds Persistent Disks still labeled for this cluster after
// the destroy. Post-destroy they are unambiguous orphans of a cluster that no
// longer exists. It deletes them when the operator consents — standing consent
// via --force, or an interactive yes to the prompt — so the cluster leaves zero
// billable leftovers; otherwise it reports them with the exact cleanup command
// and never deletes cloud data without consent. Best-effort throughout.
func (p *Provider) sweepOrphanedDisks(ctx context.Context, project, name string, force bool) {
	if project == "" {
		return
	}
	res, err := p.executor.Execute(ctx, "gcloud", "compute", "disks", "list",
		"--project", project,
		"--filter", "labels.goog-k8s-cluster-name="+name,
		"--format=value(name,zone.basename())")
	if err != nil || res == nil {
		return
	}
	disks := parseDisks(res.Stdout)
	if len(disks) == 0 {
		return
	}

	// --force is standing consent; otherwise, in an interactive session, ask.
	// The disks are unambiguous orphans of a cluster that no longer exists, but
	// they hold data, so deletion always requires consent.
	del := force
	interactive := !force && !sharedUI.IsNonInteractive()
	if !force {
		printOrphanList(disks, name)
	}
	if interactive {
		confirmed, cErr := sharedUI.ConfirmActionInteractive(
			fmt.Sprintf("Delete these %d orphaned disk(s) now?", len(disks)), false)
		del = cErr == nil && confirmed
	}
	if !del {
		printOrphanCleanupHint(project, name)
		return
	}

	pterm.Info.Printf("Cleaning up %d orphaned Persistent Disk(s) from cluster %q...\n", len(disks), name)
	var failed []string
	for _, d := range disks {
		if d.zone == "" {
			failed = append(failed, d.name+" (regional — delete manually)")
			continue
		}
		if _, err := p.executor.Execute(ctx, "gcloud", "compute", "disks", "delete", d.name,
			"--zone", d.zone, "--project", project, "--quiet"); err != nil {
			failed = append(failed, d.name)
		}
	}
	if len(failed) > 0 {
		pterm.Warning.Printf("Could not delete %d disk(s): %s\n  Remove manually: gcloud compute disks delete <name> --zone <zone> --project %s\n",
			len(failed), strings.Join(failed, ", "), project)
		return
	}
	pterm.Success.Printf("Removed %d orphaned Persistent Disk(s)\n", len(disks))
}

// printOrphanList shows the surviving disks (what a delete/report decision is about).
func printOrphanList(disks []disk, name string) {
	pterm.Warning.Printf("%d Persistent Disk(s) labeled for cluster %q survived the destroy (PVC-provisioned, outside terraform state):\n", len(disks), name)
	for _, d := range disks {
		pterm.Warning.Printf("  - %s\n", d.name)
	}
}

// printOrphanCleanupHint prints the manual cleanup command for the no-consent path.
func printOrphanCleanupHint(project, name string) {
	pterm.Warning.Printf("Left in place. Remove them with:\n"+
		"  gcloud compute disks list --project %s --filter=\"labels.goog-k8s-cluster-name=%s\"\n"+
		"  (or re-run delete with --force to have them cleaned up automatically)\n", project, name)
}

// parseDisks turns `gcloud compute disks list
// --format=value(name,zone.basename())` output into disks — one per non-empty
// line, fields separated by whitespace (name, then optional zone).
func parseDisks(out string) []disk {
	var disks []disk
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		d := disk{name: fields[0]}
		if len(fields) > 1 {
			d.zone = fields[1]
		}
		disks = append(disks, d)
	}
	return disks
}
