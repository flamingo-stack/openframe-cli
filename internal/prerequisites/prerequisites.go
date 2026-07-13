// Package prerequisites is the unified, OS-aware framework for the tools the CLI
// needs (Docker, kubectl, k3d, helm, …).
//
// It is one of the CLI's three abstractions (cluster, app, prerequisites). A
// Prerequisite knows how to check itself, optionally install itself, and where
// its manual docs are. A Runner executes a named Set with OS-aware behavior:
//
//   - macOS / Linux: missing prerequisites are auto-installed.
//   - Windows:       nothing is auto-installed; the Runner reports each missing
//     item with a docs link for the user to follow (see req 21).
//
// The concrete cluster-set and app-set are assembled from the per-tool
// installers and wired into the `openframe prerequisites` command separately.
package prerequisites

import "context"

// Prerequisite is a single tool or condition the CLI requires.
type Prerequisite struct {
	// Name is the human-facing label (e.g. "Docker").
	Name string
	// IsSatisfied reports whether the prerequisite is met (installed and, where
	// relevant, running). Required.
	IsSatisfied func() bool
	// Install attempts to install/enable the prerequisite on macOS/Linux. It may
	// be nil for things that cannot be auto-installed (then DocsURL is used).
	Install func(ctx context.Context) error
	// DocsURL points to manual setup instructions. Shown on Windows, when there
	// is no installer, or when an install attempt fails.
	DocsURL string
	// Detail, when set, explains WHY the prerequisite is unsatisfied more
	// precisely than the default "not installed" — e.g. Docker returns
	// "installed but not running" when the binary is present but the daemon is
	// down. Optional; nil means the generic "not installed" wording is used.
	Detail func() string
}

// Set is a named group of prerequisites, e.g. "cluster" or "app".
type Set struct {
	Name  string
	Items []Prerequisite
}

// MissingItem is a prerequisite that is still not satisfied after a run.
type MissingItem struct {
	Name    string
	DocsURL string
	// Reason, when set, is a specific explanation (from Prerequisite.Detail) of
	// why the item is unsatisfied — e.g. "installed but not running". Empty means
	// the item is genuinely absent and the generic "not installed" wording fits.
	Reason string
	Err    error // why it could not be auto-installed (nil on Windows / no installer)
}

// Result summarizes running a Set.
type Result struct {
	Satisfied []string      // already present
	Installed []string      // installed during this run
	Missing   []MissingItem // still missing afterwards
}

// OK reports whether every prerequisite is satisfied (nothing missing).
func (r Result) OK() bool { return len(r.Missing) == 0 }
