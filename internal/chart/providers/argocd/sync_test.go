package argocd

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

// TestRefreshAndSync_SyncsChildren proves force-sync rolls out child Applications
// too (not just the root), so non-auto-sync children don't stay OutOfSync.
func TestRefreshAndSync_SyncsChildren(t *testing.T) {
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObj("core-api", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
		appObj("nats", ArgoCDHealthHealthy, ArgoCDSyncOutOfSync),
	)

	if err := m.RefreshAndSync(context.Background(), false); err != nil {
		t.Fatalf("RefreshAndSync: %v", err)
	}

	res := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)
	for _, name := range []string{AppOfAppsName, "core-api", "nats"} {
		got, err := res.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("get %s: %v", name, err)
		}
		if _, ok := got.Object["operation"]; !ok {
			t.Errorf("%s must have a sync .operation set", name)
		}
	}
}

// TestRefreshAndSync_InFlightOperation refuses to clobber a sync that is already
// running, returning an error instead of racing it.
func TestRefreshAndSync_InFlightOperation(t *testing.T) {
	app := appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced)
	status := app.Object["status"].(map[string]interface{})
	status["operationState"] = map[string]interface{}{"phase": "Running"}
	m := fakeManager(app)

	err := m.RefreshAndSync(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), "in progress") {
		t.Fatalf("want in-progress error, got: %v", err)
	}

	// It must NOT have set a new .operation over the running one.
	got, gerr := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace).
		Get(context.Background(), AppOfAppsName, metav1.GetOptions{})
	if gerr != nil {
		t.Fatalf("get app-of-apps: %v", gerr)
	}
	if _, ok := got.Object["operation"]; ok {
		t.Error("must not set .operation while another operation is running")
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

// appObjInGroup is appObj plus the deploy-ordering SyncGroupLabel.
func appObjInGroup(name, health, sync, group string) *unstructured.Unstructured {
	obj := appObj(name, health, sync)
	obj.SetLabels(map[string]string{SyncGroupLabel: group})
	return obj
}

// TestGroupChildren_SortsAndDefaults locks the grouping rules: groups sort
// lowest-first, an unlabeled child on a labeled install falls into
// defaultSyncGroup, and a fully unlabeled install reports labeled=false with
// every child in one group (legacy single-pass behaviour).
func TestGroupChildren_SortsAndDefaults(t *testing.T) {
	children := []unstructured.Unstructured{
		*appObjInGroup("tenant", ArgoCDHealthHealthy, ArgoCDSyncSynced, "5"),
		*appObjInGroup("kafka", ArgoCDHealthHealthy, ArgoCDSyncSynced, "3"),
		*appObjInGroup("zookeeper", ArgoCDHealthHealthy, ArgoCDSyncSynced, "2"),
		*appObjInGroup("cassandra", ArgoCDHealthHealthy, ArgoCDSyncSynced, "2"),
		*appObj("unlabeled", ArgoCDHealthHealthy, ArgoCDSyncSynced), // → defaultSyncGroup
	}

	groups, labeled := groupChildren(children)
	if !labeled {
		t.Fatal("labeled = false, want true")
	}
	wantNumbers := []int{2, 3, 5}
	if len(groups) != len(wantNumbers) {
		t.Fatalf("got %d groups, want %d: %+v", len(groups), len(wantNumbers), groups)
	}
	for i, n := range wantNumbers {
		if groups[i].number != n {
			t.Errorf("groups[%d].number = %d, want %d", i, groups[i].number, n)
		}
	}
	if got := strings.Join(groups[0].names, ","); got != "cassandra,zookeeper" {
		t.Errorf("group 2 = %q, want cassandra,zookeeper", got)
	}
	if got := strings.Join(groups[1].names, ","); got != "kafka,unlabeled" {
		t.Errorf("group 3 = %q, want kafka,unlabeled (unlabeled must default to group %d)", got, defaultSyncGroup)
	}

	// Fully unlabeled install → single group, labeled=false.
	groups, labeled = groupChildren(children[4:])
	if labeled || len(groups) != 1 || len(groups[0].names) != 1 {
		t.Errorf("unlabeled install: labeled=%v groups=%+v, want one unlabeled group", labeled, groups)
	}
}

// TestRefreshAndSync_GroupedChildrenAllSynced proves the grouped path still
// triggers a sync operation on every labeled child (groups gate the ORDER,
// never drop apps), including a child whose group never converges — the gate
// is best-effort and must not block later groups.
func TestRefreshAndSync_GroupedChildrenAllSynced(t *testing.T) {
	m := fakeManager(
		appObj(AppOfAppsName, ArgoCDHealthHealthy, ArgoCDSyncSynced),
		appObjInGroup("ingress-nginx", ArgoCDHealthHealthy, ArgoCDSyncSynced, "1"),
		// Group 2 never reaches Healthy against the controllerless fake:
		// the gate must time out (tiny groupWait) and move on, not hang.
		appObjInGroup("zookeeper", ArgoCDHealthProgressing, ArgoCDSyncOutOfSync, "2"),
		appObjInGroup("kafka", ArgoCDHealthHealthy, ArgoCDSyncSynced, "3"),
	)
	m.groupWait = 10 * time.Millisecond

	if err := m.RefreshAndSync(context.Background(), false); err != nil {
		t.Fatalf("RefreshAndSync: %v", err)
	}

	res := m.dynamicClient.Resource(applicationGVR).Namespace(ArgoCDNamespace)
	for _, name := range []string{"ingress-nginx", "zookeeper", "kafka"} {
		got, err := res.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("get %s: %v", name, err)
		}
		if _, ok := got.Object["operation"]; !ok {
			t.Errorf("%s must have a sync .operation set", name)
		}
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
