package argocd

import (
	"fmt"
	"sort"
	"strings"
)

// refMismatch is one child application that ArgoCD is tracking at a different
// git revision than the ref the user asked the CLI to deploy.
type refMismatch struct {
	App  string
	Want string // requested ref
	Got  string // the app's actual spec.source.targetRevision
}

// defaultRefs are the branch names indistinguishable from "no pinning": if the
// user requests one of these and a legacy chart ignores the pin, the children
// land on main anyway — the same place — so there is nothing to warn about.
var defaultRefs = map[string]bool{"": true, "main": true, "master": true, "head": true}

// verifyRefPinning reports the OSS-repo child applications whose targetRevision
// does not match the requested ref.
//
// It exists for the V3 silent-failure: `app install --ref <old-branch>` writes
// the flattened repository.branch, but a branch whose chart predates that key
// ignores it — children render from main, the CLI waits, everything goes
// Healthy+Synced, and "17/17 ready … SUCCESS" is printed for a deployment of
// main, not the requested ref. Comparing what ArgoCD actually tracks against
// what was asked turns that into a loud, specific failure.
//
// Only children pointing at repoURL are considered: a child that legitimately
// sources a different repository (third-party) has its own revision and must
// not be flagged. A child with an empty targetRevision is skipped (unknowable),
// as is any request for a default ref (see defaultRefs).
func verifyRefPinning(apps []Application, repoURL, requestedRef string) []refMismatch {
	if defaultRefs[strings.ToLower(strings.TrimSpace(requestedRef))] {
		return nil
	}
	want := normalizeRef(requestedRef)
	repo := normalizeRepoURL(repoURL)

	var out []refMismatch
	for _, app := range apps {
		if repo != "" && normalizeRepoURL(app.RepoURL) != repo {
			continue // different repository — not ours to judge
		}
		got := strings.TrimSpace(app.TargetRevision)
		if got == "" {
			continue // no declared revision — nothing to compare
		}
		if normalizeRef(got) != want {
			out = append(out, refMismatch{App: app.Name, Want: requestedRef, Got: got})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].App < out[j].App })
	return out
}

// normalizeRef trims a git ref for comparison. Branch and tag references are
// compared by their short name; a "v"-prefixed tag and its bare form are NOT
// unified (they are distinct git refs).
func normalizeRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "refs/tags/")
	return ref
}

// normalizeRepoURL canonicalizes a git remote for comparison: lowercased, no
// scheme, no embedded credentials, no trailing ".git" or slash. Enough to match
// "https://user:tok@github.com/org/repo.git" against "github.com/org/repo".
func normalizeRepoURL(u string) string {
	u = strings.TrimSpace(strings.ToLower(u))
	if u == "" {
		return ""
	}
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	u = strings.TrimPrefix(u, "git@")
	if at := strings.LastIndex(u, "@"); at >= 0 { // strip user:token@
		u = u[at+1:]
	}
	u = strings.ReplaceAll(u, ":", "/") // scp-style host:org/repo
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimSuffix(u, ".git")
	return u
}

// refMismatchError renders the mismatches into a loud, actionable error. The
// workloads are running — but not from the ref the user asked for, so the
// requested operation did not do what it said.
func refMismatchError(requestedRef string, m []refMismatch) error {
	var b strings.Builder
	fmt.Fprintf(&b, "requested ref %q was NOT deployed: this branch's chart predates ref pinning, "+
		"so its child applications ignored the pin and ArgoCD is tracking a different revision.\n",
		requestedRef)
	b.WriteString("The workloads are running, but they do NOT reflect the requested ref:\n")
	for _, x := range m {
		fmt.Fprintf(&b, "  - %s is on %q, not %q\n", x.App, x.Got, x.Want)
	}
	b.WriteString("Use a branch whose chart reads repository.branch, or pin these applications' targetRevision by hand.")
	return fmt.Errorf("%s", b.String())
}
