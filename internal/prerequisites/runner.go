package prerequisites

import (
	"context"
	"fmt"
	"runtime"
)

// Runner executes a Set with OS-aware behavior. The zero value uses the host OS;
// tests set OS explicitly.
type Runner struct {
	// OS overrides the detected operating system ("darwin", "linux", "windows").
	// Empty means use runtime.GOOS.
	OS string
}

// NewRunner returns a Runner for the current host OS.
func NewRunner() Runner { return Runner{} }

// os returns the effective OS.
func (r Runner) os() string {
	if r.OS != "" {
		return r.OS
	}
	return runtime.GOOS
}

// AutoInstalls reports whether missing prerequisites are auto-installed on this
// OS. True on macOS/Linux, false on Windows (and anything else).
func (r Runner) AutoInstalls() bool {
	switch r.os() {
	case "darwin", "linux":
		return true
	default:
		return false
	}
}

// Check reports which prerequisites are satisfied WITHOUT installing anything.
// Use it for `prerequisites check`; use Run to actually install.
func (r Runner) Check(set Set) Result {
	var res Result
	for _, item := range set.Items {
		if satisfied(item) {
			res.Satisfied = append(res.Satisfied, item.Name)
		} else {
			res.Missing = append(res.Missing, MissingItem{Name: item.Name, DocsURL: item.DocsURL})
		}
	}
	return res
}

// Run checks every prerequisite in the set and, on supported OSes, installs the
// missing ones. It never returns an error itself — inspect Result.OK() /
// Result.Missing to decide whether to proceed.
func (r Runner) Run(ctx context.Context, set Set) Result {
	auto := r.AutoInstalls()
	var res Result

	for _, item := range set.Items {
		if satisfied(item) {
			res.Satisfied = append(res.Satisfied, item.Name)
			continue
		}

		// Not satisfied. Auto-install only when the OS supports it and an
		// installer exists.
		if auto && item.Install != nil {
			if err := item.Install(ctx); err != nil {
				res.Missing = append(res.Missing, MissingItem{Name: item.Name, DocsURL: item.DocsURL, Err: err})
				continue
			}
			if satisfied(item) {
				res.Installed = append(res.Installed, item.Name)
				continue
			}
			res.Missing = append(res.Missing, MissingItem{
				Name:    item.Name,
				DocsURL: item.DocsURL,
				Err:     fmt.Errorf("%s was installed but is still not satisfied", item.Name),
			})
			continue
		}

		// Windows, or no installer available: report as manual (docs link).
		res.Missing = append(res.Missing, MissingItem{Name: item.Name, DocsURL: item.DocsURL})
	}

	return res
}

// satisfied is nil-safe: a prerequisite with no check is treated as satisfied.
func satisfied(p Prerequisite) bool {
	return p.IsSatisfied == nil || p.IsSatisfied()
}
