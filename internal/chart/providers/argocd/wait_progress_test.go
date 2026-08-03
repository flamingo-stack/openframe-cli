package argocd

import (
	"strings"
	"testing"
	"time"
)

// TestTimeoutError_NamesTheStuckApplications (M3.2): the wait loop knows which
// applications never became ready. The old message threw that away and said
// only "timeout waiting for ArgoCD applications after 1h0m0s", leaving the user
// to go find the stuck app by hand.
func TestTimeoutError_NamesTheStuckApplications(t *testing.T) {
	err := timeoutError(30*time.Minute, 4, 6,
		[]string{"openframe-api (Health: Progressing)", "openframe-ui (Health: Degraded)"},
		[]string{"openframe-api", "openframe-ui"}, nil, "")

	msg := err.Error()
	for _, want := range []string{
		"30m0s",                               // how long it waited
		"4/6 ready",                           // how far it got
		"openframe-api (Health: Progressing)", // decorated label in the list
		"openframe-ui (Health: Degraded)",
		"kubectl get applications -n argocd", // what to run next
		"kubectl describe application openframe-api -n argocd",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout error must contain %q; got:\n%s", want, msg)
		}
	}
}

// TestTimeoutError_KubectlHintUsesBareName is the regression guard for the CI
// bug: notReady labels carry a " (Health: ...)" suffix for display, and the
// kubectl-describe hint used one verbatim, producing the un-runnable
// "kubectl describe application argocd-apps (Health: Progressing) -n argocd".
func TestTimeoutError_KubectlHintUsesBareName(t *testing.T) {
	msg := timeoutError(time.Minute, 14, 17,
		[]string{"argocd-apps (Health: Progressing)"},
		[]string{"argocd-apps"}, nil, "").Error()

	if !strings.Contains(msg, "kubectl describe application argocd-apps -n argocd") {
		t.Errorf("the kubectl hint must use the bare app name; got:\n%s", msg)
	}
	if strings.Contains(msg, "describe application argocd-apps (Health") {
		t.Errorf("the kubectl command must not contain the decorated label; got:\n%s", msg)
	}
}

// TestTimeoutError_BoundsTheApplicationList: a large platform can leave dozens
// of applications pending. The list must not bury the next-step hint.
func TestTimeoutError_BoundsTheApplicationList(t *testing.T) {
	var many []string
	for i := 0; i < 25; i++ {
		many = append(many, "app-"+string(rune('a'+i)))
	}

	msg := timeoutError(time.Minute, 0, 25, many, many, nil, "").Error()

	if !strings.Contains(msg, "and 15 more") {
		t.Errorf("the list must be truncated with a count of the remainder; got:\n%s", msg)
	}
	if strings.Contains(msg, "app-y") {
		t.Errorf("the 25th application must not be listed; got:\n%s", msg)
	}
	if !strings.Contains(msg, "kubectl get applications") {
		t.Errorf("the next-step hint must survive truncation; got:\n%s", msg)
	}
}

// TestTimeoutError_NoAppsIsStillLegible: timing out before any application was
// observed (app-of-apps never produced children) must not print an empty list.
func TestTimeoutError_NoAppsIsStillLegible(t *testing.T) {
	msg := timeoutError(time.Minute, 0, 0, nil, nil, nil, "").Error()

	if strings.Contains(msg, "still not ready:") {
		t.Errorf("an empty list must be omitted, not printed empty; got:\n%s", msg)
	}
	if strings.Contains(msg, "describe application") {
		t.Errorf("there is no application to describe; got:\n%s", msg)
	}
	if !strings.Contains(msg, "timeout after 1m0s") {
		t.Errorf("the message must still state the timeout; got:\n%s", msg)
	}
}

// The timeout error must carry per-app health messages so a Degraded app's
// failing pod is named (why it hung), not just that it hung.
func TestTimeoutError_IncludesHealthMessages(t *testing.T) {
	details := notReadyDiagnostics([]Application{
		{Name: "tenant", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncSynced,
			HealthMessage: "pod openframe-stream-abc123 is in CrashLoopBackOff"},
		{Name: "cassandra", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced}, // ready — excluded
	})
	msg := timeoutError(time.Minute, 15, 17,
		[]string{"tenant (Health: Degraded)"}, []string{"tenant"}, details, "").Error()

	for _, want := range []string{
		"Why they are not ready:",
		"tenant: health=Degraded sync=Synced",
		"openframe-stream-abc123 is in CrashLoopBackOff",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout error must contain %q; got:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "cassandra") {
		t.Errorf("a ready app must not appear in the diagnostics; got:\n%s", msg)
	}
}

// The per-app block must name the unhealthy CHILD resources — the exact
// Deployment/child-Application failing inside a Degraded app — because the
// app-level health message is frequently empty (the user's report showed bare
// "health=Degraded sync=Synced" lines and nothing else).
func TestNotReadyDiagnostics_NamesUnhealthyResources(t *testing.T) {
	details := notReadyDiagnostics([]Application{
		{Name: "tenant", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncSynced,
			UnhealthyResources: []ResourceIssue{
				{Kind: "Deployment", Name: "openframe-api", Namespace: "tenant",
					Health: "Degraded", Message: "Deployment \"openframe-api\" exceeded its progress deadline"},
				{Kind: "StatefulSet", Name: "kafka", Namespace: "datasources", Health: "Progressing"},
			}},
	})
	if len(details) != 1 {
		t.Fatalf("expected one diagnostic block, got %d", len(details))
	}
	for _, want := range []string{
		"Deployment tenant/openframe-api: Degraded — Deployment \"openframe-api\" exceeded its progress deadline",
		"StatefulSet datasources/kafka: Progressing",
	} {
		if !strings.Contains(details[0], want) {
			t.Errorf("diagnostics must contain %q; got:\n%s", want, details[0])
		}
	}
}

// One wedged node can degrade dozens of resources at once; the list must be
// bounded with a count of the remainder.
func TestNotReadyDiagnostics_BoundsResourceList(t *testing.T) {
	var many []ResourceIssue
	for i := 0; i < maxResourcesPerAppInDiagnostics+4; i++ {
		many = append(many, ResourceIssue{Kind: "Deployment", Name: "d-" + string(rune('a'+i)), Health: "Degraded"})
	}
	details := notReadyDiagnostics([]Application{
		{Name: "tenant", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncSynced, UnhealthyResources: many},
	})
	if !strings.Contains(details[0], "and 4 more unhealthy resource(s)") {
		t.Errorf("resource list must be truncated with a remainder count; got:\n%s", details[0])
	}
}

// The timeout error must embed the pod-level workload diagnosis (crash logs,
// events) when one was gathered — and be SelfDiagnosed so the generic handler
// does not pattern-match the embedded logs into a bogus hint.
func TestTimeoutError_IncludesWorkloadDiagnosis(t *testing.T) {
	diag := "\ntenant (namespace tenant):\n  pod api-abc / api: CrashLoopBackOff (restarts=7)\n    last log lines:\n      fatal: MONGODB_URI is not set"
	err := timeoutError(time.Minute, 15, 17,
		[]string{"tenant (Health: Degraded)"}, []string{"tenant"}, nil, diag)

	msg := err.Error()
	for _, want := range []string{
		"Failing workloads (pod errors, last logs, warning events):",
		"CrashLoopBackOff (restarts=7)",
		"MONGODB_URI is not set",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout error must contain %q; got:\n%s", want, msg)
		}
	}
	var sd interface{ SelfDiagnosed() bool }
	if !errorsAs(err, &sd) || !sd.SelfDiagnosed() {
		t.Fatal("a timeout error carrying a workload diagnosis must be SelfDiagnosed")
	}
}

// notReadyApplications feeds the timeout diagnosis: only not-ready apps, capped.
func TestNotReadyApplications_FiltersAndCaps(t *testing.T) {
	apps := []Application{
		{Name: "ready", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced},
		{Name: "a", Health: ArgoCDHealthDegraded, Sync: ArgoCDSyncSynced},
		{Name: "b", Health: ArgoCDHealthProgressing, Sync: ArgoCDSyncSynced},
		{Name: "c", Health: ArgoCDHealthHealthy, Sync: "OutOfSync"},
	}
	got := notReadyApplications(apps, 2)
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" {
		t.Fatalf("expected the first 2 not-ready apps, got %+v", got)
	}
	if got := notReadyApplications(apps, 10); len(got) != 3 {
		t.Fatalf("expected all 3 not-ready apps under a large cap, got %d", len(got))
	}
}

func TestDefaultAppWaitTimeout_EnvOverride(t *testing.T) {
	t.Run("defaults to 60m when unset", func(t *testing.T) {
		t.Setenv("OPENFRAME_APP_WAIT_TIMEOUT", "")
		if got := defaultAppWaitTimeout(); got != 60*time.Minute {
			t.Fatalf("default = %v, want 60m", got)
		}
	})
	t.Run("honours a valid duration", func(t *testing.T) {
		t.Setenv("OPENFRAME_APP_WAIT_TIMEOUT", "35m")
		if got := defaultAppWaitTimeout(); got != 35*time.Minute {
			t.Fatalf("override = %v, want 35m", got)
		}
	})
	t.Run("ignores garbage and falls back", func(t *testing.T) {
		t.Setenv("OPENFRAME_APP_WAIT_TIMEOUT", "not-a-duration")
		if got := defaultAppWaitTimeout(); got != 60*time.Minute {
			t.Fatalf("garbage override = %v, want fallback 60m", got)
		}
	})
}
