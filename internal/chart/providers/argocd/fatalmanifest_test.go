package argocd

import (
	"strings"
	"testing"
	"time"
)

// legacyRefCondition is the real-world shape of the ComparisonError a legacy
// ref produces: the repo is reachable, the chart path simply does not exist at
// the pinned revision.
const legacyRefCondition = "Failed to load target state: failed to generate manifest for source 1 of 1: " +
	"rpc error: code = Unknown desc = Manifest generation error (cached): " +
	"/tmp/repo/manifests/oss: no such file or directory"

func TestIsDeterministicManifestError(t *testing.T) {
	cases := []struct {
		name      string
		condition string
		want      bool
	}{
		{"legacy ref: path missing at revision", legacyRefCondition, true},
		{"app path does not exist variant",
			"Failed to load target state: Unable to generate manifests in manifests/oss: " +
				"rpc error: code = Unknown desc = manifests/oss: app path does not exist", true},
		{"transient repo-server EOF stays recoverable",
			"failed to generate manifest for source 1 of 1: rpc error: EOF", false},
		{"transient Unavailable stays recoverable",
			"Manifest generation error: rpc error: code = Unavailable desc = connection refused", false},
		{"deterministic fragment without manifest context does not trip",
			"hook failed: /scripts/setup.sh: no such file or directory", false},
		{"empty condition", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isDeterministicManifestError(c.condition); got != c.want {
				t.Fatalf("isDeterministicManifestError(%q) = %v, want %v", c.condition, got, c.want)
			}
		})
	}
}

// TestFatalManifestTracker_FailsAfterPersistence is the motivating case: the
// same deterministic error, tick after tick, must eventually be reported
// instead of riding the wait to its full timeout.
func TestFatalManifestTracker_FailsAfterPersistence(t *testing.T) {
	tr := newFatalManifestTracker()
	app := []Application{{Name: "oss", Health: "Degraded", Sync: "OutOfSync", Condition: legacyRefCondition}}
	start := time.Now()

	// Enough checks but not enough wall-clock time → not fatal yet.
	for i := 0; i < fatalManifestMinChecks+2; i++ {
		if fatal := tr.observe(app, start.Add(time.Duration(i)*time.Second)); len(fatal) != 0 {
			t.Fatalf("fatal before fatalManifestAfter elapsed (tick %d)", i)
		}
	}

	// Past the wall-clock threshold with the check count already met → fatal.
	fatal := tr.observe(app, start.Add(fatalManifestAfter+time.Second))
	if len(fatal) != 1 || fatal[0].Name != "oss" {
		t.Fatalf("fatal = %v, want [oss]", names(fatal))
	}
}

// TestFatalManifestTracker_TimeAloneIsNotEnough: two isolated sightings around
// a long gap must not trip the fail-fast — the check count guards it.
func TestFatalManifestTracker_TimeAloneIsNotEnough(t *testing.T) {
	tr := newFatalManifestTracker()
	app := []Application{{Name: "oss", Health: "Degraded", Sync: "OutOfSync", Condition: legacyRefCondition}}
	start := time.Now()

	tr.observe(app, start)
	if fatal := tr.observe(app, start.Add(fatalManifestAfter+time.Minute)); len(fatal) != 0 {
		t.Fatalf("fatal after only 2 checks, want none (min %d)", fatalManifestMinChecks)
	}
}

// TestFatalManifestTracker_ClearsWhenErrorResolves: an app whose condition
// goes away (or that becomes ready) starts a fresh clock if it ever returns.
func TestFatalManifestTracker_ClearsWhenErrorResolves(t *testing.T) {
	tr := newFatalManifestTracker()
	broken := []Application{{Name: "oss", Health: "Degraded", Sync: "OutOfSync", Condition: legacyRefCondition}}
	start := time.Now()

	for i := 0; i < fatalManifestMinChecks; i++ {
		tr.observe(broken, start.Add(time.Duration(i)*time.Second))
	}

	// Error clears for one tick — the entry must be forgotten.
	tr.observe([]Application{{Name: "oss", Health: "Progressing", Sync: "OutOfSync"}}, start.Add(time.Minute))

	// Error returns after the original wall-clock threshold: the fresh clock
	// must keep it non-fatal.
	if fatal := tr.observe(broken, start.Add(fatalManifestAfter+time.Minute)); len(fatal) != 0 {
		t.Fatal("tracker must reset after the error resolves")
	}
}

// TestFatalManifestTracker_ReadyAppNeverFatal: a stale condition on an app
// that is currently Healthy+Synced must never fail the install.
func TestFatalManifestTracker_ReadyAppNeverFatal(t *testing.T) {
	tr := newFatalManifestTracker()
	app := []Application{{Name: "oss", Health: ArgoCDHealthHealthy, Sync: ArgoCDSyncSynced, Condition: legacyRefCondition}}
	start := time.Now()

	for i := 0; i < fatalManifestMinChecks+1; i++ {
		tr.observe(app, start.Add(time.Duration(i)*time.Second))
	}
	if fatal := tr.observe(app, start.Add(fatalManifestAfter+time.Minute)); len(fatal) != 0 {
		t.Fatal("ready app must never be reported fatal")
	}
}

func TestFatalManifestError_NonDefaultRefNamesTheRef(t *testing.T) {
	err := fatalManifestError("release-1.0", []Application{
		{Name: "oss", Condition: legacyRefCondition},
	})
	msg := err.Error()
	for _, want := range []string{"oss", "deterministic", `"release-1.0"`, "no such file or directory"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error must contain %q, got:\n%s", want, msg)
		}
	}
}

func TestFatalManifestError_DefaultRefSkipsRefHint(t *testing.T) {
	err := fatalManifestError("main", []Application{{Name: "oss", Condition: legacyRefCondition}})
	msg := err.Error()
	if strings.Contains(msg, "requested ref") {
		t.Errorf("default ref must not produce the ref hint, got:\n%s", msg)
	}
	if !strings.Contains(msg, "kubectl describe application oss -n argocd") {
		t.Errorf("default-ref message must give the inspect command, got:\n%s", msg)
	}
}

func TestFatalManifestError_TruncatesLongConditions(t *testing.T) {
	long := legacyRefCondition + strings.Repeat(" pad", 200)
	err := fatalManifestError("release-1.0", []Application{{Name: "oss", Condition: long}})
	if len(err.Error()) > maxConditionInError+500 {
		t.Fatalf("condition not truncated, message length %d", len(err.Error()))
	}
}

// TestClassifyAppIssues_DeterministicErrorsBypassRecovery locks the recovery
// exclusion: a deterministic manifest error matches the broad "failed to
// generate manifest" repo-server pattern, but restarting the repo-server
// cannot fix it, so it must not be counted as a repo-server issue.
func TestClassifyAppIssues_DeterministicErrorsBypassRecovery(t *testing.T) {
	counts := map[string]int{}
	apps := []Application{
		{Name: "legacy", Health: "Degraded", Sync: "OutOfSync", Condition: legacyRefCondition},
		{Name: "transient", Health: "Progressing", Sync: "OutOfSync",
			Condition: "failed to generate manifest for source 1 of 1: rpc error: EOF"},
	}

	_, condErrs := classifyAppIssues(apps, counts)

	if len(condErrs) != 1 || condErrs[0].Name != "transient" {
		t.Fatalf("conditionErrors = %v, want [transient]", names(condErrs))
	}
	if _, ok := counts["legacy"]; ok {
		t.Error("deterministic error must not accumulate a recovery count")
	}
	if counts["transient"] != 1 {
		t.Errorf("transient error must still count, got %v", counts)
	}
}
