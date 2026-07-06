package argocd

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestRefreshAndSync_PatchesAppOfApps proves force-sync applies both the hard
// refresh annotation and the top-level .operation sync to the app-of-apps
// Application via the dynamic client.
func TestRefreshAndSync_PatchesAppOfApps(t *testing.T) {
	// prune defaults to false: a force-sync must not delete workloads.
	for _, prune := range []bool{false, true} {
		m := fakeManager(appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced))

		if err := m.RefreshAndSync(context.Background(), prune); err != nil {
			t.Fatalf("RefreshAndSync(prune=%v): %v", prune, err)
		}

		got, err := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace).
			Get(context.Background(), AppOfAppsName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("get app-of-apps: %v", err)
		}

		// 1) Hard refresh annotation is set.
		ann, _, _ := unstructuredNestedString(got.Object, "metadata", "annotations", "argocd.argoproj.io/refresh")
		if ann != "hard" {
			t.Errorf("refresh annotation = %q, want hard", ann)
		}

		// 2) A sync operation is present with prune matching the requested value.
		op, ok := got.Object["operation"].(map[string]interface{})
		if !ok {
			t.Fatalf("no .operation set; object=%v", got.Object)
		}
		sync, ok := op["sync"].(map[string]interface{})
		if !ok {
			t.Fatalf(".operation.sync missing: %v", op)
		}
		if gotPrune, _ := sync["prune"].(bool); gotPrune != prune {
			t.Errorf(".operation.sync.prune = %v, want %v", sync["prune"], prune)
		}
	}
}

// TestSyncOperationPatch_PruneDefault locks that the default patch does NOT prune.
func TestSyncOperationPatch_PruneDefault(t *testing.T) {
	if !strings.Contains(syncOperationPatch(false), `"prune":false`) {
		t.Errorf("default sync patch must not prune: %s", syncOperationPatch(false))
	}
	if !strings.Contains(syncOperationPatch(true), `"prune":true`) {
		t.Errorf("prune=true patch must prune: %s", syncOperationPatch(true))
	}
}

// TestRefreshAndSync_NotInstalled surfaces a friendly error when app-of-apps is
// absent (OpenFrame not installed).
func TestRefreshAndSync_NotInstalled(t *testing.T) {
	m := fakeManager() // no applications

	err := m.RefreshAndSync(context.Background(), false)
	if err == nil {
		t.Fatal("expected an error when app-of-apps is missing")
	}
	if !strings.Contains(err.Error(), "is OpenFrame installed") {
		t.Errorf("want friendly not-installed error, got: %v", err)
	}
}

// TestRefreshAndSync_NoClient errors clearly when the dynamic client is
// unavailable and cannot be initialized.
func TestRefreshAndSync_NoClient(t *testing.T) {
	m := &Manager{clientsInitialized: true} // initialized but nil dynamicClient

	if err := m.RefreshAndSync(context.Background(), false); err == nil {
		t.Fatal("expected an error when dynamic client is nil")
	}
}

// unstructuredNestedString reads a nested string value from an unstructured map.
func unstructuredNestedString(obj map[string]interface{}, fields ...string) (string, bool, error) {
	cur := obj
	for i, f := range fields {
		if i == len(fields)-1 {
			s, ok := cur[f].(string)
			return s, ok, nil
		}
		next, ok := cur[f].(map[string]interface{})
		if !ok {
			return "", false, nil
		}
		cur = next
	}
	return "", false, nil
}
