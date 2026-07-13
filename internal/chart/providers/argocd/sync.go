package argocd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// refreshAnnotationKey is the annotation ArgoCD watches to force a refresh; its
// controller clears it once the refresh has been processed.
const refreshAnnotationKey = "argocd.argoproj.io/refresh"

// AppOfAppsName is the root Application that owns the whole platform. Refreshing
// and syncing it cascades to every child Application, so the force-sync path only
// has to act on this one object. openframe-oss-tenant renamed it from
// "app-of-apps" to "argocd-apps".
const AppOfAppsName = "argocd-apps"

// refreshHardPatch forces the repo-server to re-read git and drop its cached
// manifests (a plain "normal" refresh can serve stale cache for a moved ref).
// CONSTANT JSON — no interpolation, no injection surface.
const refreshHardPatch = `{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}`

// syncOperationPatch builds the top-level .operation sync patch — the exact
// mechanism `argocd app sync` uses, driven through the CRD so no argocd CLI or
// API port is needed. `prune` controls whether ArgoCD DELETES resources no
// longer present in git; it is off by default because a force-sync of a moved
// ref must never silently delete workloads (deleting a child Application
// cascades to its resources). Only the boolean is interpolated, so there is no
// injection surface.
func syncOperationPatch(prune bool) string {
	return fmt.Sprintf(`{"operation":{"initiatedBy":{"username":"openframe-cli"},"sync":{"prune":%t,"syncStrategy":{"apply":{"force":false}}}}}`, prune)
}

// RefreshAndSync forces ArgoCD to re-read git for the root app-of-apps
// Application and trigger a sync, WITHOUT changing its targetRevision. This is
// the "force-sync" mode of `app upgrade`: use it to pull a moved floating ref
// (e.g. a branch whose HEAD advanced) and roll it out when auto-sync is off or
// the repo-server is serving stale manifests.
//
// prune enables deletion of resources removed from git (off by default — see
// syncOperationPatch). It applies two dynamic-client patches to app-of-apps: a
// hard refresh, then a sync operation. ArgoCD reconciles the cascade to child
// Applications itself.
func (m *Manager) RefreshAndSync(ctx context.Context, prune bool) error {
	if m.dynamicClient == nil {
		if err := m.initKubernetesClients(); err != nil {
			return err
		}
	}
	if m.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	apps := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)

	// 1) Hard refresh — repo-server re-reads git, bypassing the manifest cache.
	if _, err := apps.Patch(ctx, AppOfAppsName, types.MergePatchType, []byte(refreshHardPatch), metav1.PatchOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("%s not found in namespace %q — is OpenFrame installed?", AppOfAppsName, ArgoCDNamespace)
		}
		return fmt.Errorf("refreshing %s: %w", AppOfAppsName, err)
	}

	budget := m.syncWait
	if budget <= 0 {
		budget = 30 * time.Second
	}

	// 1b) Wait for the controller to process the hard refresh — it clears the
	// refresh annotation once it has re-read git — so the sync below runs against
	// fresh manifests, not the pre-refresh cache. Best-effort: proceed when the
	// annotation clears or the budget elapses (never fatal).
	m.waitRefreshCleared(ctx, budget)

	// 2) Don't clobber an operation already in flight — setting .operation over a
	// running one is racy; wait briefly, then refuse rather than stomp it.
	if err := m.ensureNoRunningOperation(ctx, budget); err != nil {
		return err
	}

	// 3) Trigger the sync via the top-level .operation field.
	if _, err := apps.Patch(ctx, AppOfAppsName, types.MergePatchType, []byte(syncOperationPatch(prune)), metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("triggering sync of %s: %w", AppOfAppsName, err)
	}

	// 4) Sync the child Applications too. Syncing only the root updates child
	// specs but does not roll them out: children that are not auto-sync would
	// stay OutOfSync and WaitForApplications would then block until its timeout.
	// Best-effort — a child that already has an operation in flight is skipped.
	return m.syncChildApplications(ctx, prune)
}

// trackingInstanceLabel is ArgoCD's LABEL resource-tracking marker: resources
// managed by an Application carry app.kubernetes.io/instance=<app>. Child
// Applications created by the app-of-apps therefore carry the root's name.
const trackingInstanceLabel = "app.kubernetes.io/instance"

// trackingIDAnnotation is ArgoCD's ANNOTATION resource-tracking marker, used
// when resourceTrackingMethod is "annotation" or "annotation+label" (the label
// above is then absent). Its value is "<owner-app>:<group>/<kind>:<ns>/<name>",
// e.g. "argocd-apps:argoproj.io/Application:argocd/openframe-api".
const trackingIDAnnotation = "argocd.argoproj.io/tracking-id"

// trackingOwner returns the owning Application name encoded in either tracking
// marker, or "" if neither is present. The label wins when set; otherwise the
// annotation's owner is the segment before the first ":". Splitting (rather
// than prefix-matching) avoids mistaking "argocd-apps-foo" for "argocd-apps".
func trackingOwner(labels, annotations map[string]string) string {
	if v := labels[trackingInstanceLabel]; v != "" {
		return v
	}
	if id := annotations[trackingIDAnnotation]; id != "" {
		return strings.SplitN(id, ":", 2)[0]
	}
	return ""
}

// syncChildApplications triggers a sync on the root's child Applications.
// Children are selected by ArgoCD's resource tracking, checking BOTH markers:
// the label (app.kubernetes.io/instance=argocd-apps) and the annotation
// (argocd.argoproj.io/tracking-id, owner before the first ":"). The annotation
// is required because ArgoCD's "annotation" / "annotation+label" tracking
// methods leave the label empty — the case the verification run hit, where the
// primary selector matched nothing and the fallback synced everything.
//
// Only when NEITHER marker is present on any Application does it fall back to
// every Application except the root, rather than silently syncing nothing —
// but that may touch Applications that are not OpenFrame-owned (a real risk on
// a shared cluster), which the fallback warning makes visible.
//
// Per-child failures no longer vanish (audit F8): individual errors are
// counted and surfaced — a warning on partial failure, an error when NOT ONE
// child could be synced (previously `app upgrade --sync` would then "succeed"
// into a 15-minute wait timeout with no hint).
func (m *Manager) syncChildApplications(ctx context.Context, prune bool) error {
	apps := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)
	list, err := apps.List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing applications to sync: %w", err)
	}

	var children []string
	for i := range list.Items {
		name := list.Items[i].GetName()
		if name == AppOfAppsName {
			continue
		}
		if trackingOwner(list.Items[i].GetLabels(), list.Items[i].GetAnnotations()) == AppOfAppsName {
			children = append(children, name)
		}
	}
	if children == nil {
		for i := range list.Items {
			if name := list.Items[i].GetName(); name != AppOfAppsName {
				children = append(children, name)
			}
		}
		if len(children) > 0 {
			pterm.Warning.Printf("No applications carry the %s=%s tracking label; syncing all %d applications in %q\n",
				trackingInstanceLabel, AppOfAppsName, len(children), ArgoCDNamespace)
		}
	}

	patched, failed, firstErr := m.syncApplicationsByName(ctx, children, prune)
	if failed > 0 && patched == 0 {
		return fmt.Errorf("could not trigger a sync on any of the %d child applications (first error: %w)", failed, firstErr)
	}
	if failed > 0 {
		pterm.Warning.Printf("Triggered sync on %d application(s); %d failed (first error: %v)\n", patched, failed, firstErr)
	}
	return nil
}

// syncApplicationsByName applies the sync-operation patch to each named
// Application, returning how many were patched, how many failed, and the first
// failure. Lazily initializes the Kubernetes clients like RefreshAndSync.
func (m *Manager) syncApplicationsByName(ctx context.Context, names []string, prune bool) (patched, failed int, firstErr error) {
	if m.dynamicClient == nil {
		if err := m.initKubernetesClients(); err != nil {
			return 0, len(names), err
		}
	}
	apps := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)
	patch := []byte(syncOperationPatch(prune))
	for _, name := range names {
		if _, err := apps.Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
			failed++
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", name, err)
			}
			continue
		}
		patched++
	}
	return patched, failed, firstErr
}

// appOfAppsObject fetches the current app-of-apps Application (unstructured).
func (m *Manager) appOfAppsObject(ctx context.Context) (*unstructured.Unstructured, error) {
	return m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace).
		Get(ctx, AppOfAppsName, metav1.GetOptions{})
}

// waitRefreshCleared polls until ArgoCD clears the refresh annotation (refresh
// processed) or the budget elapses. Best-effort — it never returns an error; the
// sync still triggers a comparison, this just avoids racing the pre-refresh cache.
func (m *Manager) waitRefreshCleared(ctx context.Context, budget time.Duration) {
	pollUntil(ctx, budget, func() bool {
		obj, err := m.appOfAppsObject(ctx)
		if err != nil {
			return true // can't read — stop waiting, let the sync proceed
		}
		v, _, _ := unstructured.NestedString(obj.Object, "metadata", "annotations", refreshAnnotationKey)
		return v == "" // cleared
	})
}

// ensureNoRunningOperation waits (up to budget) for any in-flight sync to finish,
// then errors if one is still running rather than clobbering it.
func (m *Manager) ensureNoRunningOperation(ctx context.Context, budget time.Duration) error {
	running := false
	pollUntil(ctx, budget, func() bool {
		obj, err := m.appOfAppsObject(ctx)
		if err != nil {
			running = false
			return true // can't read — stop; the sync patch will surface real errors
		}
		phase, _, _ := unstructured.NestedString(obj.Object, "status", "operationState", "phase")
		running = phase == "Running"
		return !running
	})
	if running {
		return fmt.Errorf("%s already has a sync operation in progress; try again shortly", AppOfAppsName)
	}
	return nil
}

// pollUntil calls done() immediately and then every ~500ms until it returns true,
// the context is cancelled, or the budget elapses.
func pollUntil(ctx context.Context, budget time.Duration, done func() bool) {
	if done() {
		return
	}
	interval := 500 * time.Millisecond
	if budget < interval {
		interval = budget
	}
	if interval <= 0 {
		interval = time.Millisecond
	}
	timer := time.NewTimer(budget)
	defer timer.Stop()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			return
		case <-ticker.C:
			if done() {
				return
			}
		}
	}
}
