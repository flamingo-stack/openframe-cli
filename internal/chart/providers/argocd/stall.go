package argocd

import (
	"fmt"
	"sort"
	"strings"
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

// stallAfter is how long the application set may stay bit-for-bit identical
// before the wait considers itself stalled. Long enough that a slow-but-live
// rollout (images pulling, probes settling) keeps resetting it via status
// transitions; short enough to leave time to act within the wait budget.
const stallAfter = 90 * time.Second

// stallFingerprint captures the observable wait state: ready count plus the
// sorted not-ready list (names with their status strings). Any transition —
// an app appearing, progressing, or changing status — changes the fingerprint.
func stallFingerprint(ready int, notReady []string) string {
	sorted := append([]string(nil), notReady...)
	sort.Strings(sorted)
	return fmt.Sprintf("%d|%s", ready, strings.Join(sorted, ","))
}

// outOfSyncStragglers returns the names of applications that are not ready
// solely because they are OutOfSync (health already fine). These are the ones
// a sync operation can actually move; apps with health problems are excluded —
// syncing them would mask the real failure.
func outOfSyncStragglers(apps []Application) []string {
	var names []string
	for _, app := range apps {
		if app.Sync == ArgoCDSyncOutOfSync && app.Health == ArgoCDHealthHealthy {
			names = append(names, app.Name)
		}
	}
	sort.Strings(names)
	return names
}
