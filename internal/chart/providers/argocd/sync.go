package argocd

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// AppOfAppsName is the root app-of-apps Application that owns the whole
// platform. Refreshing and syncing it cascades to every child Application, so
// the force-sync path only has to act on this one object.
const AppOfAppsName = "app-of-apps"

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
			return fmt.Errorf("app-of-apps not found in namespace %q — is OpenFrame installed?", ArgoCDNamespace)
		}
		return fmt.Errorf("refreshing app-of-apps: %w", err)
	}

	// 2) Trigger the sync via the top-level .operation field.
	if _, err := apps.Patch(ctx, AppOfAppsName, types.MergePatchType, []byte(syncOperationPatch(prune)), metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("triggering sync of app-of-apps: %w", err)
	}
	return nil
}
