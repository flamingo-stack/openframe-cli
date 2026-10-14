package gke

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/shared"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// systemNamespacePrefixes lists the GKE-specific system namespace prefixes
// (in addition to the shared "kube-" prefix) that are never deleted during
// teardown. They are the cluster's own control-plane/system namespaces (torn
// down with the cluster anyway) and — critically — kube-system hosts the GKE PD
// CSI controller that must keep running to delete the Persistent Disks as their
// PVCs go away.
var systemNamespacePrefixes = []string{"gke-", "gmp-"}

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
	targets := shared.AppNamespacesToDelete(names, systemNamespacePrefixes)
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
		return shared.CountDeletablePVs(pvs.Items) == 0, nil
	})
}

// disk is one labeled Persistent Disk found by the post-destroy sweep.
type disk struct {
	name   string
	zone   string // basename, e.g. "us-central1-a"; empty for a regional disk
	region string // basename, e.g. "us-central1"; set for regional disks only
}

// location returns where the disk lives: its zone, or its region for a
// regional disk.
func (d disk) location() string {
	if d.zone != "" {
		return d.zone
	}
	return d.region
}

// disksInLocation keeps only disks that can belong to the deleted cluster.
// GKE cluster names are unique per LOCATION, not per project, so the
// name-label filter alone would also match — and with --force delete — the
// still-attached disks of a same-named, live cluster in another region. The
// record stores the cluster's region; a zonal disk of that cluster lives in
// "<region>-<suffix>". An empty location (a record predating this field)
// keeps everything, degrading to the unscoped behavior.
func disksInLocation(disks []disk, location string) []disk {
	if location == "" {
		return disks
	}
	var out []disk
	for _, d := range disks {
		loc := d.location()
		if loc == location || strings.HasPrefix(loc, location+"-") {
			out = append(out, d)
		}
	}
	return out
}

// sweepOrphanedDisks finds Persistent Disks still labeled for this cluster after
// the destroy. Post-destroy they are unambiguous orphans of a cluster that no
// longer exists. It deletes them when the operator consents — standing consent
// via --force, or an interactive yes to the prompt — so the cluster leaves zero
// billable leftovers; otherwise it reports them with the exact cleanup command
// and never deletes cloud data without consent. Best-effort throughout.
func (p *Provider) sweepOrphanedDisks(ctx context.Context, project, name, location string, force bool) {
	if project == "" {
		return
	}
	res, err := p.executor.Execute(ctx, "gcloud", "compute", "disks", "list",
		"--project", project,
		"--filter", "labels.goog-k8s-cluster-name="+name,
		"--format=value(name,zone.basename(),region.basename())")
	if err != nil || res == nil {
		return
	}
	disks := disksInLocation(parseDisks(res.Stdout), location)
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
		// The location tells the operator WHICH cluster's disks these are —
		// GKE cluster names repeat across locations.
		// List items go through DefaultBasicText (the repo's header+items
		// pattern): the Warning header above carries the severity, and
		// repeating the warning tag on every item just breaks the column.
		if loc := d.location(); loc != "" {
			pterm.DefaultBasicText.Printf("  - %s (%s)\n", d.name, loc)
		} else {
			pterm.DefaultBasicText.Printf("  - %s\n", d.name)
		}
	}
}

// printOrphanCleanupHint prints the manual cleanup command for the no-consent path.
func printOrphanCleanupHint(project, name string) {
	pterm.Warning.Printf("Left in place. Remove them with:\n"+
		"  gcloud compute disks list --project %s --filter=\"labels.goog-k8s-cluster-name=%s\"\n"+
		"  (or re-run delete with --force to have them cleaned up automatically)\n", project, name)
}

// parseDisks turns `gcloud compute disks list
// --format=value(name,zone.basename(),region.basename())` output into disks —
// one per non-empty line. value() output is TAB-separated with empty columns
// kept: a regional disk has an empty zone column, so a whitespace split would
// shift its region into the zone slot.
func parseDisks(out string) []disk {
	var disks []disk
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		d := disk{name: strings.TrimSpace(fields[0])}
		if len(fields) > 1 {
			d.zone = strings.TrimSpace(fields[1])
		}
		if len(fields) > 2 {
			d.region = strings.TrimSpace(fields[2])
		}
		disks = append(disks, d)
	}
	return disks
}
