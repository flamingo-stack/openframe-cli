package argocd

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Fail-fast for deterministic manifest-generation failures (0.4.9 verification
// observation): a legacy ref whose chart path does not exist at the pinned
// revision produces a persistent ComparisonError ("Manifest generation error
// ... no such file or directory"). No retry, repo-server restart, or amount of
// waiting can change that outcome — the repo-server fetched the repo fine, the
// content simply isn't there. The wait previously classified it as repo-server
// trouble (restarting the repo-server up to 3 times) and then rode out the
// full timeout (20+ minutes observed) before failing.

// manifestErrorContexts identify a condition as a manifest-generation failure
// at all; only then are the deterministic patterns consulted. This keeps a
// stray "no such file or directory" from an unrelated condition from tripping
// the fail-fast.
var manifestErrorContexts = []string{
	"failed to generate manifest",
	"Manifest generation error",
	"Unable to generate manifests",
}

// deterministicManifestPatterns are fragments that make a manifest-generation
// failure deterministic: the requested path/revision content does not exist,
// so the error is permanent for this install. Transient repo-server trouble
// (EOF, Unavailable, connection resets) must NOT be listed here — those stay
// on the recovery path in classifyAppIssues.
var deterministicManifestPatterns = []string{
	"no such file or directory",
	"app path does not exist",
}

// isDeterministicManifestError reports whether a condition message describes a
// manifest-generation failure that waiting cannot fix.
func isDeterministicManifestError(condition string) bool {
	if condition == "" {
		return false
	}
	inManifestContext := false
	for _, c := range manifestErrorContexts {
		if strings.Contains(condition, c) {
			inManifestContext = true
			break
		}
	}
	if !inManifestContext {
		return false
	}
	for _, p := range deterministicManifestPatterns {
		if strings.Contains(condition, p) {
			return true
		}
	}
	return false
}

// fatalManifestAfter is how long an application's deterministic manifest error
// must persist before the wait fails fast. Long enough to survive a one-off
// condition observed mid-sync (ArgoCD caches and re-evaluates manifests during
// the first waves); short enough to save the 20+ minute ride to the timeout.
const fatalManifestAfter = 2 * time.Minute

// fatalManifestMinChecks is the minimum number of consecutive observations of
// the same deterministic error. Guards against a wall-clock threshold being
// crossed by two isolated sightings around a long connectivity gap.
const fatalManifestMinChecks = 5

// fatalManifestTracker records, per application, how long a deterministic
// manifest error has persisted. Mirrors stallTracker's shape: reset on change,
// forget on disappearance.
type fatalManifestTracker struct {
	entries map[string]fatalManifestEntry
}

type fatalManifestEntry struct {
	since  time.Time // when the deterministic error was first observed
	checks int       // consecutive observations
}

func newFatalManifestTracker() *fatalManifestTracker {
	return &fatalManifestTracker{entries: make(map[string]fatalManifestEntry)}
}

// observe records this tick's applications and returns those whose
// deterministic manifest error has persisted past both thresholds. An app that
// stops showing the error (or becomes ready) is forgotten, so its clock starts
// fresh if the error ever returns.
func (t *fatalManifestTracker) observe(apps []Application, now time.Time) []Application {
	var fatal []Application
	seen := make(map[string]bool, len(apps))
	for _, app := range apps {
		ready := app.Health == ArgoCDHealthHealthy && app.Sync == ArgoCDSyncSynced
		if ready || !isDeterministicManifestError(app.Condition) {
			continue
		}
		seen[app.Name] = true
		e, ok := t.entries[app.Name]
		if !ok {
			e = fatalManifestEntry{since: now}
		}
		e.checks++
		t.entries[app.Name] = e
		if e.checks >= fatalManifestMinChecks && now.Sub(e.since) >= fatalManifestAfter {
			fatal = append(fatal, app)
		}
	}
	for name := range t.entries {
		if !seen[name] {
			delete(t.entries, name)
		}
	}
	sort.Slice(fatal, func(i, j int) bool { return fatal[i].Name < fatal[j].Name })
	return fatal
}

// maxConditionInError bounds how much of a condition message lands in the
// fail-fast error; ArgoCD prefixes them with several layers of rpc wrapping.
const maxConditionInError = 300

// fatalManifestError renders the fail-fast error. When a non-default ref was
// requested it names the likely cause — the same legacy-ref situation
// refMismatchError covers, except here the children DID honor the pin and the
// pinned revision simply lacks the expected chart content.
func fatalManifestError(requestedRef string, apps []Application) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%d application(s) cannot render their manifests from the deployed revision; "+
		"this error is deterministic — retrying or waiting cannot fix it:\n", len(apps))
	for _, app := range apps {
		cond := app.Condition
		if len(cond) > maxConditionInError {
			cond = cond[:maxConditionInError] + "..."
		}
		fmt.Fprintf(&b, "  - %s: %s\n", app.Name, cond)
	}
	if !defaultRefs[strings.ToLower(strings.TrimSpace(requestedRef))] {
		fmt.Fprintf(&b, "The requested ref %q likely predates the chart layout this CLI expects "+
			"(the chart path does not exist at that revision).\n"+
			"Use a branch whose chart matches this CLI, or fix the applications' path/targetRevision by hand.",
			requestedRef)
	} else {
		b.WriteString("The chart path does not exist at the deployed revision. " +
			"Inspect the application source with: kubectl describe application " + apps[0].Name + " -n argocd")
	}
	return fmt.Errorf("%s", b.String())
}
