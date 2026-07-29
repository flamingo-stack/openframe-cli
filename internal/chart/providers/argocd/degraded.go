package argocd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Fail-fast for terminally-Degraded applications. A Degraded+Synced app means
// ArgoCD applied every manifest but a workload is unhealthy and has already
// passed its progress deadline; it will not become Healthy on its own. Previously
// the wait rode such an app out to the full timeout (silently — the CI job cap
// killed it first with no diagnostic). This detects a *persistent* Degraded+Synced
// app and, only when a pod is genuinely stuck (CrashLoop / ImagePull / repeated
// crashes), aborts with the pod diagnostic instead of hanging.

// degradedFailAfter is how long an app must stay continuously Degraded+Synced
// before it is a fail-fast candidate. Generous by default so a slow-but-eventually
// -healthy workload is not cut off (Degraded already means past the progress
// deadline, so this is *additional* stuck time); overridable via
// OPENFRAME_DEGRADED_FAIL_AFTER (a Go duration, e.g. "5m").
func degradedFailAfter() time.Duration {
	const fallback = 8 * time.Minute
	if v := strings.TrimSpace(os.Getenv("OPENFRAME_DEGRADED_FAIL_AFTER")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}

// degradedMinChecks guards the wall-clock threshold against being crossed by two
// isolated sightings around a long gap (mirrors fatalManifestMinChecks).
const degradedMinChecks = 10

// degradedTracker records, per application, how long it has stayed Degraded+
// Synced. Reset on change, forget on disappearance — an app that recovers or
// goes back to Progressing starts its clock fresh.
type degradedTracker struct {
	entries map[string]degradedEntry
	after   time.Duration
}

type degradedEntry struct {
	since  time.Time
	checks int
}

func newDegradedTracker() *degradedTracker {
	return &degradedTracker{entries: make(map[string]degradedEntry), after: degradedFailAfter()}
}

// observe returns apps that have been Degraded+Synced continuously past both
// thresholds. Being a candidate is necessary but not sufficient to abort — the
// caller confirms a genuinely-stuck pod first (diagnoseFailingApps.terminal).
func (t *degradedTracker) observe(apps []Application, now time.Time) []Application {
	var out []Application
	seen := make(map[string]bool, len(apps))
	for _, app := range apps {
		if app.Health != ArgoCDHealthDegraded || app.Sync != ArgoCDSyncSynced {
			continue
		}
		seen[app.Name] = true
		e, ok := t.entries[app.Name]
		if !ok {
			e = degradedEntry{since: now}
		}
		e.checks++
		t.entries[app.Name] = e
		if e.checks >= degradedMinChecks && now.Sub(e.since) >= t.after {
			out = append(out, app)
		}
	}
	for name := range t.entries {
		if !seen[name] {
			delete(t.entries, name)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// degradedAppError renders the fail-fast error, embedding the pod diagnostic
// (crash logs / events) so it says WHY the app is stuck, not just that it is.
func degradedAppError(apps []Application, diag string) error {
	names := make([]string, len(apps))
	for i, a := range apps {
		names[i] = a.Name
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d application(s) are Degraded with a workload that will not recover on its own: %s\n",
		len(apps), strings.Join(names, ", "))
	if d := strings.TrimSpace(diag); d != "" {
		b.WriteString("Why:" + diag + "\n")
	}
	b.WriteString("This is terminal — retrying or waiting cannot fix it.\n")
	fmt.Fprintf(&b, "Inspect: kubectl get pods -n %s   (and: kubectl describe application %s -n argocd)",
		apps[0].Namespace, apps[0].Name)
	return fmt.Errorf("%s", b.String())
}
