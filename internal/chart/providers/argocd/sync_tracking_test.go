package argocd

import (
	"context"
	goruntime "runtime"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

// annotatedAppObj builds an Application tracked by ArgoCD's ANNOTATION method:
// the app.kubernetes.io/instance label is absent, and ownership lives only in
// argocd.argoproj.io/tracking-id ("<owner>:<group>/<kind>:<ns>/<name>").
func annotatedAppObj(name, owner string) *unstructured.Unstructured {
	o := appObj(name, ArgoCDHealthHealthy, ArgoCDSyncOutOfSync)
	meta := o.Object["metadata"].(map[string]interface{})
	meta["annotations"] = map[string]interface{}{
		trackingIDAnnotation: owner + ":argoproj.io/Application:" + ArgoCDNamespace + "/" + name,
	}
	return o
}

// patchedNames captures the names of every Application the sync patches.
func patchedNames(m *Manager) *[]string {
	var names []string
	dc := m.dynamicClient.(interface {
		PrependReactor(verb, resource string, fn k8stesting.ReactionFunc)
	})
	dc.PrependReactor("patch", "applications", func(action k8stesting.Action) (bool, runtime.Object, error) {
		names = append(names, action.(k8stesting.PatchAction).GetName())
		return false, nil, nil
	})
	return &names
}

// TestTrackingOwner is the pure selection logic: label wins; otherwise the
// annotation's owner is the segment before the first ":"; a look-alike name is
// not matched; neither marker yields "".
func TestTrackingOwner(t *testing.T) {
	cases := []struct {
		name        string
		labels      map[string]string
		annotations map[string]string
		want        string
	}{
		{"label", map[string]string{trackingInstanceLabel: "argocd-apps"}, nil, "argocd-apps"},
		{"annotation", nil, map[string]string{trackingIDAnnotation: "argocd-apps:argoproj.io/Application:argocd/api"}, "argocd-apps"},
		{"label wins over annotation", map[string]string{trackingInstanceLabel: "argocd-apps"}, map[string]string{trackingIDAnnotation: "other:x/y:z/w"}, "argocd-apps"},
		{"lookalike not matched", nil, map[string]string{trackingIDAnnotation: "argocd-apps-foo:argoproj.io/Application:argocd/api"}, "argocd-apps-foo"},
		{"neither", nil, nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trackingOwner(tc.labels, tc.annotations); got != tc.want {
				t.Errorf("trackingOwner = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestSyncChildApplications_SelectsByAnnotation is the finding fix: on an
// annotation-tracking cluster (no instance label) children are still selected
// by the tracking-id annotation, a foreign Application (different owner) is
// left alone, and the "sync everything" fallback does NOT fire.
func TestSyncChildApplications_SelectsByAnnotation(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced), // root, no tracking marker
		annotatedAppObj("openframe-api", AppOfAppsName),
		annotatedAppObj("openframe-ui", AppOfAppsName),
		annotatedAppObj("tenants-billing", "some-other-root"), // foreign owner
	)
	got := patchedNames(m)

	if err := m.syncChildApplications(context.Background(), false); err != nil {
		t.Fatalf("syncChildApplications: %v", err)
	}
	sort.Strings(*got)
	want := []string{"openframe-api", "openframe-ui"}
	if len(*got) != 2 || (*got)[0] != want[0] || (*got)[1] != want[1] {
		t.Errorf("synced %v, want %v (annotation-owned children only)", *got, want)
	}
	for _, n := range *got {
		if n == "tenants-billing" {
			t.Error("a foreign-owned Application must not be synced")
		}
		if n == AppOfAppsName {
			t.Error("the root must never sync itself here")
		}
	}
}

// TestSyncChildApplications_FallbackWhenNoTracking preserves the safety net:
// when NEITHER marker is present anywhere, every non-root Application is synced
// (better than silently syncing nothing), with the visible warning.
func TestSyncChildApplications_FallbackWhenNoTracking(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObj("untracked-a", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
		appObj("untracked-b", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
	)
	got := patchedNames(m)

	if err := m.syncChildApplications(context.Background(), false); err != nil {
		t.Fatalf("syncChildApplications: %v", err)
	}
	sort.Strings(*got)
	if len(*got) != 2 || (*got)[0] != "untracked-a" || (*got)[1] != "untracked-b" {
		t.Errorf("fallback must sync all non-root apps, got %v", *got)
	}
}
