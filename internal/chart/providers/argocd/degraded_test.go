package argocd

import (
	stderrors "errors"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func degApp(name, health, sync string) Application {
	return Application{Name: name, Health: health, Sync: sync, Namespace: name}
}

// A Degraded+Synced app must be reported only after it persists past BOTH
// thresholds; a recovering or Progressing app resets its clock.
func TestDegradedTracker_FiresOnlyAfterPersistence(t *testing.T) {
	tr := &degradedTracker{entries: map[string]degradedEntry{}, after: 2 * time.Minute}
	start := time.Unix(1000, 0)

	// One sighting: not enough (needs minChecks + wall-clock).
	if out := tr.observe([]Application{degApp("tenant", ArgoCDHealthDegraded, ArgoCDSyncSynced)}, start); len(out) != 0 {
		t.Fatalf("single observation must not fire, got %v", out)
	}
	// Enough checks AND enough wall-clock later → fires.
	var out []Application
	for i := 1; i <= degradedMinChecks+1; i++ {
		out = tr.observe([]Application{degApp("tenant", ArgoCDHealthDegraded, ArgoCDSyncSynced)}, start.Add(time.Duration(i)*30*time.Second))
	}
	if len(out) != 1 || out[0].Name != "tenant" {
		t.Fatalf("expected tenant to fire after persistence, got %v", out)
	}
}

func TestDegradedTracker_IgnoresNonTerminalStates(t *testing.T) {
	tr := newDegradedTracker()
	now := time.Unix(2000, 0)
	// Progressing (transient) and OutOfSync are not Degraded+Synced → never tracked.
	for i := 0; i < degradedMinChecks+5; i++ {
		out := tr.observe([]Application{
			degApp("a", ArgoCDHealthProgressing, ArgoCDSyncSynced),
			degApp("b", ArgoCDHealthDegraded, ArgoCDSyncOutOfSync),
		}, now.Add(time.Duration(i)*time.Minute))
		if len(out) != 0 {
			t.Fatalf("non Degraded+Synced apps must never fire, got %v", out)
		}
	}
}

func TestDegradedTracker_ResetsWhenAppRecovers(t *testing.T) {
	tr := &degradedTracker{entries: map[string]degradedEntry{}, after: time.Minute}
	base := time.Unix(3000, 0)
	// Build up some persistence...
	for i := 0; i < 5; i++ {
		tr.observe([]Application{degApp("x", ArgoCDHealthDegraded, ArgoCDSyncSynced)}, base.Add(time.Duration(i)*20*time.Second))
	}
	// ...then it recovers (disappears from the Degraded set) → forgotten.
	tr.observe([]Application{degApp("x", ArgoCDHealthHealthy, ArgoCDSyncSynced)}, base.Add(2*time.Minute))
	if _, ok := tr.entries["x"]; ok {
		t.Fatal("a recovered app must be forgotten so its clock restarts")
	}
}

func TestDegradedFailAfter_EnvOverride(t *testing.T) {
	t.Run("default 8m", func(t *testing.T) {
		t.Setenv("OPENFRAME_DEGRADED_FAIL_AFTER", "")
		if got := degradedFailAfter(); got != 8*time.Minute {
			t.Fatalf("default = %v, want 8m", got)
		}
	})
	t.Run("valid override", func(t *testing.T) {
		t.Setenv("OPENFRAME_DEGRADED_FAIL_AFTER", "3m")
		if got := degradedFailAfter(); got != 3*time.Minute {
			t.Fatalf("override = %v, want 3m", got)
		}
	})
	t.Run("garbage falls back", func(t *testing.T) {
		t.Setenv("OPENFRAME_DEGRADED_FAIL_AFTER", "soon")
		if got := degradedFailAfter(); got != 8*time.Minute {
			t.Fatalf("garbage = %v, want fallback 8m", got)
		}
	})
}

func TestDegradedAppError_EmbedsDiagnostic(t *testing.T) {
	err := degradedAppError(
		[]Application{{Name: "tenant", Namespace: "tenant"}},
		"\ntenant (namespace tenant):\n  pod openframe-stream-x / stream: CrashLoopBackOff (restarts=4)",
	)
	msg := err.Error()
	for _, want := range []string{"tenant", "CrashLoopBackOff", "terminal", "kubectl get pods -n tenant"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error must contain %q; got:\n%s", want, msg)
		}
	}
}

func waitingPod(name, container, reason, image string, restarts int32) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name: container, Image: image, RestartCount: restarts,
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
		}}},
	}
}

func TestFailingContainers_ClassifiesTerminalStates(t *testing.T) {
	// CrashLoopBackOff → terminal.
	got := failingContainers(waitingPod("p", "c", "CrashLoopBackOff", "img", 4))
	if len(got) != 1 || !got[0].terminal || got[0].reason != "CrashLoopBackOff" {
		t.Fatalf("CrashLoopBackOff must be terminal, got %+v", got)
	}
	// ImagePullBackOff → terminal + image-pull.
	got = failingContainers(waitingPod("p", "c", "ImagePullBackOff", "ghcr.io/x:latest", 0))
	if len(got) != 1 || !got[0].terminal || !isImagePullReason(got[0].reason) {
		t.Fatalf("ImagePullBackOff must be terminal+imagepull, got %+v", got)
	}
	// Benign startup states are NOT failures.
	if got := failingContainers(waitingPod("p", "c", "ContainerCreating", "img", 0)); len(got) != 0 {
		t.Fatalf("ContainerCreating must not be a failure, got %+v", got)
	}
	if got := failingContainers(waitingPod("p", "c", "PodInitializing", "img", 0)); len(got) != 0 {
		t.Fatalf("PodInitializing must not be a failure, got %+v", got)
	}
}

func TestIsImagePullReason(t *testing.T) {
	for _, r := range []string{"ImagePullBackOff", "ErrImagePull", "InvalidImageName"} {
		if !isImagePullReason(r) {
			t.Errorf("%q must be an image-pull reason", r)
		}
	}
	if isImagePullReason("CrashLoopBackOff") {
		t.Error("CrashLoopBackOff is not an image-pull reason")
	}
}

// Both diagnostic-carrying errors must be marked SelfDiagnosed so the generic
// handler suppresses its pattern-matched hint (which misfires on embedded pod
// logs like "connect: connection refused").
func TestDiagnosticErrors_AreSelfDiagnosed(t *testing.T) {
	type selfDiagnosed interface{ SelfDiagnosed() bool }

	dErr := degradedAppError([]Application{{Name: "tenant", Namespace: "tenant"}}, "\n  pod x: CrashLoopBackOff")
	var sd selfDiagnosed
	if !errorsAs(dErr, &sd) || !sd.SelfDiagnosed() {
		t.Fatal("degradedAppError must be SelfDiagnosed")
	}

	tErr := timeoutError(time.Minute, 1, 2, []string{"a"}, []string{"a"}, []string{"  - a: health=Degraded"})
	sd = nil
	if !errorsAs(tErr, &sd) || !sd.SelfDiagnosed() {
		t.Fatal("timeoutError WITH diagnostics must be SelfDiagnosed")
	}

	// Without diagnostics the timeout error stays plain — generic hints stay useful.
	plain := timeoutError(time.Minute, 1, 2, []string{"a"}, []string{"a"}, nil)
	sd = nil
	if errorsAs(plain, &sd) {
		t.Fatal("timeoutError WITHOUT diagnostics must stay a plain error")
	}
}

func errorsAs(err error, target any) bool { return stderrors.As(err, target) }

// Once past its thresholds, a candidate must be re-reported at the recheck
// interval, not on every 2s tick: each report triggers a full pod diagnosis
// (pod list + log streams + events), ~1800 rounds an hour otherwise.
func TestDegradedTracker_ThrottlesRecheck(t *testing.T) {
	tr := &degradedTracker{entries: map[string]degradedEntry{}, after: time.Minute}
	base := time.Unix(5000, 0)
	app := []Application{degApp("x", ArgoCDHealthDegraded, ArgoCDSyncSynced)}

	// Drive past both thresholds at a 2s cadence and count the reports.
	fired := 0
	var lastFire time.Time
	var prevFire time.Time
	for i := 0; i < 400; i++ { // ~13 minutes of ticks
		now := base.Add(time.Duration(i) * 2 * time.Second)
		if out := tr.observe(app, now); len(out) > 0 {
			fired++
			prevFire, lastFire = lastFire, now
		}
	}
	if fired == 0 {
		t.Fatal("a persistently Degraded app must still be reported")
	}
	if fired > 15 {
		t.Fatalf("reports must be throttled to ~1/min, got %d in ~13m", fired)
	}
	if !prevFire.IsZero() && lastFire.Sub(prevFire) < degradedRecheckInterval {
		t.Fatalf("consecutive reports %v apart, want >= %v", lastFire.Sub(prevFire), degradedRecheckInterval)
	}
}
