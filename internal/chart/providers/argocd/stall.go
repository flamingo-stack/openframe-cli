package argocd

import (
	"sort"
	"time"
)

// Stall handling for WaitForApplications (0.4.7 verification finding N3):
// after a ref change, children with autoSync disabled settle into a persistent
// OutOfSync state that the wait would previously ride out to its full timeout
// (11+ minutes of identical "Still waiting for: [mongodb ...]" lines) with no
// hint that nothing was ever going to change.
//
// The wait loop keeps a fingerprint of the observable state; when it has not
// changed for stallAfter and OutOfSync stragglers remain, it either triggers a
// sync on exactly those stragglers (upgrade path, SyncStragglersOnStall) or
// prints an actionable hint once (install path).

// stallAfter is how long a SINGLE application may stay bit-for-bit identical
// before the wait considers it stalled. Long enough that a slow-but-live
// rollout (images pulling, probes settling) keeps resetting it via status
// transitions; short enough to leave time to act within the wait budget.
const stallAfter = 90 * time.Second

// stallTracker records, per application, how long its (health, sync) state has
// been unchanged.
//
// It replaces a single global fingerprint over the whole not-ready set (V5): a
// global timer is reset by ANY transition anywhere in the set, so one noisy
// neighbour oscillating Missing<->OutOfSync every tick reset the 90s clock
// forever while a genuinely stuck app sat Healthy+OutOfSync, bit-for-bit
// identical, and never accrued stall time. Tracking each app independently lets
// a stuck app be detected regardless of what its neighbours do.
type stallTracker struct {
	states map[string]stallEntry
}

type stallEntry struct {
	state string    // "health|sync"
	since time.Time // when the app first entered `state`
}

func newStallTracker() *stallTracker {
	return &stallTracker{states: make(map[string]stallEntry)}
}

// observe records each application's current state, resetting an app's timer
// when its state changes and forgetting apps no longer present (so a
// reappearing app starts its clock fresh rather than inheriting a stale one).
func (s *stallTracker) observe(apps []Application, now time.Time) {
	seen := make(map[string]bool, len(apps))
	for _, app := range apps {
		seen[app.Name] = true
		state := app.Health + "|" + app.Sync
		if e, ok := s.states[app.Name]; !ok || e.state != state {
			s.states[app.Name] = stallEntry{state: state, since: now}
		}
	}
	for name := range s.states {
		if !seen[name] {
			delete(s.states, name)
		}
	}
}

// stalledStragglers returns the OutOfSync-but-Healthy applications whose state
// has been identical for at least stallAfter. These are the apps a sync can
// actually move (health is already fine, so syncing won't mask a real failure)
// AND that are genuinely stuck, judged per-app rather than across the set.
//
// Callers must observe() this same tick's apps first.
func (s *stallTracker) stalledStragglers(apps []Application, now time.Time) []string {
	var names []string
	for _, app := range apps {
		if app.Sync != ArgoCDSyncOutOfSync || app.Health != ArgoCDHealthHealthy {
			continue
		}
		if e, ok := s.states[app.Name]; ok && now.Sub(e.since) >= stallAfter {
			names = append(names, app.Name)
		}
	}
	sort.Strings(names)
	return names
}
