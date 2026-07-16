package models

import "fmt"

// CleanupResult records what a cleanup run ACTUALLY did, so the summary can
// report facts instead of a fixed script.
//
// Cleanup is best-effort by design: every phase swallows its own error so that
// a half-installed or partly-unreachable cluster can still be torn down. The
// old code paired that with a summary that unconditionally printed "Removed
// unused Docker images / Freed up disk space / Optimized cluster performance",
// so a run in which every phase failed was indistinguishable from a clean one.
// Counting the work and collecting the failures is what makes the best-effort
// contract honest.
type CleanupResult struct {
	ApplicationsDeleted int
	FinalizersCleared   int
	ReleasesRemoved     int
	NamespacesDeleted   int
	NodesPruned         int

	// Failures holds one human-readable line per phase that did not complete.
	// A non-empty Failures with a nil error is the normal "partial cleanup"
	// outcome: the command succeeds, but the user is told what was left behind.
	Failures []string
}

// AddFailure records a phase failure, prefixed with the phase name.
func (r *CleanupResult) AddFailure(phase string, err error) {
	r.Failures = append(r.Failures, fmt.Sprintf("%s: %v", phase, err))
}

// Removed reports the total number of objects cleanup actually removed.
func (r CleanupResult) Removed() int {
	return r.ApplicationsDeleted + r.FinalizersCleared + r.ReleasesRemoved +
		r.NamespacesDeleted + r.NodesPruned
}

// Partial reports whether at least one phase failed. Cleanup still succeeded
// overall; some resources may remain.
func (r CleanupResult) Partial() bool { return len(r.Failures) > 0 }
