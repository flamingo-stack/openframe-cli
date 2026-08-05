package eks

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

// orphanVolumeTagKey is the tag the EBS CSI driver stamps on every volume it
// provisions for this cluster (extraVolumeTags in templates/main.tf). Those
// volumes live OUTSIDE the terraform state, so this tag is the only handle the
// post-destroy sweep has on them. The template guard test pins the two sides
// together.
const orphanVolumeTagKey = "openframe:cluster"

// systemNamespaces and systemNamespacePrefixes are never deleted during
// teardown. They are the cluster's own control-plane/system namespaces (torn
// down with the cluster anyway) and — critically — kube-system hosts the EBS
// CSI controller that must keep running to delete the EBS volumes as their
// PVCs go away, plus the cloud controller that deletes the load balancers as
// their Services go away.
var systemNamespaces = map[string]struct{}{
	"default":         {},
	"kube-system":     {},
	"kube-public":     {},
	"kube-node-lease": {},
}

var systemNamespacePrefixes = []string{"kube-", "aws-", "amazon-"}

const (
	// resourceDrainTimeout bounds how long a delete waits for PVC-backed volumes
	// and Service-backed load balancers to be reclaimed before proceeding to
	// destroy anyway. A live-platform teardown must never block indefinitely;
	// volumes not drained in time are caught by the post-destroy sweep, and a
	// surviving load balancer fails the VPC destroy with a retriable error.
	resourceDrainTimeout = 5 * time.Minute
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
// PVCs whose backing volumes must be released; deleting by discovery rather
// than a hardcoded list keeps this correct as the platform layout changes.
// argocd goes first so its controller stops re-syncing before the workloads it
// manages are deleted, otherwise self-heal could recreate a StatefulSet (and
// its PVC) mid-teardown.
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
// These are the volumes whose backing EBS volume the CSI driver removes once
// their PVC is gone, so the release step waits for this to reach zero.
// Retain-policy PVs are excluded on purpose — their volumes are meant to
// survive, and the post-destroy sweep reports (never silently drops) them.
func countDeletablePVs(pvs []corev1.PersistentVolume) int {
	var n int
	for _, pv := range pvs {
		if pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			n++
		}
	}
	return n
}

// countAppLoadBalancers counts Services of type LoadBalancer outside the system
// namespaces. Each one is an ELB/NLB in AWS, created by the in-cluster cloud
// controller OUTSIDE the terraform state — and its ENIs sit in the VPC's
// subnets, so a survivor fails the whole VPC destroy with DependencyViolation.
// The service.kubernetes.io/load-balancer-cleanup finalizer keeps the Service
// object alive until the cloud load balancer is actually gone, so "no LB
// Services left" is a faithful signal that AWS is clean.
func countAppLoadBalancers(svcs []corev1.Service) int {
	var n int
	for _, svc := range svcs {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && !isSystemNamespace(svc.Namespace) {
			n++
		}
	}
	return n
}

// releaseWorkloadResources deletes every application namespace on the cluster
// and waits (bounded) for the Delete-reclaim PersistentVolumes AND the
// LoadBalancer Services to drain, so the EBS CSI driver deletes the backing
// volumes and the cloud controller deletes the ELBs/NLBs BEFORE terraform
// destroys the node group and VPC. Both live OUTSIDE the terraform state:
// without this, volumes survive as billable orphans and load balancers make
// the VPC destroy fail outright (DependencyViolation on the subnets).
//
// Deleting whole namespaces (rather than draining nodes) also avoids the
// autoscaler racing to reschedule evicted pods onto a fresh node mid-teardown.
//
// Entirely best-effort: an unreachable cluster, RBAC denial, or a drain timeout
// are logged and ignored. The infra destroy the user asked for must never be
// blocked by teardown; the post-destroy sweep reports/cleans leftover volumes,
// and a leftover load balancer surfaces as a retriable destroy error.
func releaseWorkloadResources(ctx context.Context, rec tfengine.Record) {
	restCfg, err := restConfigFor(rec)
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy resource release (no rest config): %v\n", err)
		return
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy resource release (no kube client): %v\n", err)
		return
	}

	listCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	nsList, err := client.CoreV1().Namespaces().List(listCtx, metav1.ListOptions{})
	cancel()
	if err != nil {
		pterm.Debug.Printf("skip pre-destroy resource release (cluster unreachable): %v\n", err)
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

	pterm.Info.Printf("Releasing EBS volumes and load balancers (removing namespaces: %s) before destroy...\n", strings.Join(targets, ", "))
	for _, ns := range targets {
		delCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		err := client.CoreV1().Namespaces().Delete(delCtx, ns, metav1.DeleteOptions{})
		cancel()
		if err != nil && !k8serrors.IsNotFound(err) {
			pterm.Debug.Printf("namespace %s delete: %v\n", ns, err)
		}
	}

	// Wait for Delete-reclaim PVs and app LoadBalancer Services to disappear —
	// the signal that the CSI driver has deleted the backing volumes and the
	// cloud controller has deleted the ELBs. Bounded; on timeout we proceed and
	// let the sweep/destroy-retry handle the remainder.
	_ = wait.PollUntilContextTimeout(ctx, 5*time.Second, resourceDrainTimeout, true, func(ctx context.Context) (bool, error) {
		pvCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		defer cancel()
		pvs, err := client.CoreV1().PersistentVolumes().List(pvCtx, metav1.ListOptions{})
		if err != nil {
			return false, nil // transient — keep polling until the outer timeout
		}
		if countDeletablePVs(pvs.Items) > 0 {
			return false, nil
		}
		svcCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		defer cancel()
		svcs, err := client.CoreV1().Services(metav1.NamespaceAll).List(svcCtx, metav1.ListOptions{})
		if err != nil {
			return false, nil
		}
		return countAppLoadBalancers(svcs.Items) == 0, nil
	})
}

// sweepOrphanedVolumes finds EBS volumes still tagged for this cluster after
// the destroy. Post-destroy they are unambiguous orphans of a cluster that no
// longer exists (describe-volumes is region-scoped, so a same-named cluster in
// another region is out of reach by construction; the status=available filter
// additionally refuses anything still attached). It deletes them when the
// operator consents — standing consent via --force, or an interactive yes to
// the prompt — so the cluster leaves zero billable leftovers; otherwise it
// reports them with the exact cleanup command and never deletes cloud data
// without consent. Best-effort throughout. (The GKE twin sweeps Persistent
// Disks by label.)
func (p *Provider) sweepOrphanedVolumes(ctx context.Context, rec tfengine.Record, force bool) {
	if rec.Region == "" {
		return
	}
	args := []string{"ec2", "describe-volumes",
		"--region", rec.Region,
		"--filters",
		fmt.Sprintf("Name=tag:%s,Values=%s", orphanVolumeTagKey, rec.Name),
		"Name=status,Values=available",
		"--query", "Volumes[].VolumeId", "--output", "text"}
	if rec.Profile != "" {
		args = append(args, "--profile", rec.Profile)
	}
	res, err := p.executor.Execute(ctx, "aws", args...)
	if err != nil || res == nil {
		return
	}
	volumes := parseVolumeIDs(res.Stdout)
	if len(volumes) == 0 {
		return
	}

	// --force is standing consent; otherwise, in an interactive session, ask.
	// The volumes are unambiguous orphans of a cluster that no longer exists,
	// but they hold data, so deletion always requires consent.
	del := force
	interactive := !force && !sharedUI.IsNonInteractive()
	if !force {
		printOrphanList(volumes, rec.Name)
	}
	if interactive {
		confirmed, cErr := sharedUI.ConfirmActionInteractive(
			fmt.Sprintf("Delete these %d orphaned EBS volume(s) now?", len(volumes)), false)
		del = cErr == nil && confirmed
	}
	if !del {
		printOrphanCleanupHint(rec)
		return
	}

	pterm.Info.Printf("Cleaning up %d orphaned EBS volume(s) from cluster %q...\n", len(volumes), rec.Name)
	var failed []string
	for _, id := range volumes {
		delArgs := []string{"ec2", "delete-volume", "--volume-id", id, "--region", rec.Region}
		if rec.Profile != "" {
			delArgs = append(delArgs, "--profile", rec.Profile)
		}
		if _, err := p.executor.Execute(ctx, "aws", delArgs...); err != nil {
			failed = append(failed, id)
		}
	}
	if len(failed) > 0 {
		pterm.Warning.Printf("Could not delete %d volume(s): %s\n  Remove manually: aws ec2 delete-volume --volume-id <id> --region %s\n",
			len(failed), strings.Join(failed, ", "), rec.Region)
		return
	}
	pterm.Success.Printf("Removed %d orphaned EBS volume(s)\n", len(volumes))
}

// printOrphanList shows the surviving volumes (what a delete/report decision is
// about).
func printOrphanList(volumes []string, name string) {
	pterm.Warning.Printf("%d EBS volume(s) tagged for cluster %q survived the destroy (PVC-provisioned, outside terraform state):\n", len(volumes), name)
	for _, id := range volumes {
		// Items under the Warning header go through DefaultBasicText so the
		// warning tag isn't repeated per line (same pattern as cleanup lists).
		pterm.DefaultBasicText.Printf("  - %s\n", id)
	}
}

// printOrphanCleanupHint prints the manual cleanup command for the no-consent
// path.
func printOrphanCleanupHint(rec tfengine.Record) {
	pterm.Warning.Printf("Left in place. Remove them with:\n"+
		"  aws ec2 describe-volumes --region %s --filters \"Name=tag:%s,Values=%s\"\n"+
		"  (or re-run delete with --force to have them cleaned up automatically)\n",
		rec.Region, orphanVolumeTagKey, rec.Name)
}

// parseVolumeIDs turns `aws ec2 describe-volumes --query Volumes[].VolumeId
// --output text` output into volume ids. text output separates values with
// tabs/newlines; an empty result prints nothing (or the literal "None").
func parseVolumeIDs(out string) []string {
	var ids []string
	for _, field := range strings.Fields(out) {
		if field == "" || field == "None" {
			continue
		}
		ids = append(ids, field)
	}
	return ids
}
