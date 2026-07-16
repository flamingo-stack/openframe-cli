package errors

import (
	"fmt"
	"strings"
	"testing"
)

// friendlyHint matches on substrings of OTHER PROJECTS' error strings (helm,
// client-go, Docker). Nothing stops those projects from rewording a message in
// a patch release, and when they do, the hint silently disappears — the CLI
// keeps working, just less helpfully, and no test notices.
//
// These cases pin verbatim messages, each transcribed from the tool that emits
// it. If an upstream rewording breaks a hint, this test says so, with the
// message that must be re-checked. It is a canary, not a correctness proof.
func TestFriendlyHint_MatchesRepresentativeUpstreamMessages(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    string // a distinctive fragment of the expected hint
	}{
		{
			name:    "helm: CRDs left by an aborted install",
			message: `rendered manifests contain a resource that already exists. Unable to continue with install: CustomResourceDefinition "applications.argoproj.io" in namespace "" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"`,
			want:    "already exists without Helm ownership metadata",
		},
		{
			name:    "client-go: apiserver down",
			message: `Get "https://0.0.0.0:6550/api?timeout=32s": dial tcp 0.0.0.0:6550: connect: connection refused`,
			want:    "isn't reachable",
		},
		{
			name:    "kubectl: stale kubeconfig host",
			message: `Get "https://k3d-dev-server-0:6443/version": dial tcp: lookup k3d-dev-server-0: no such host`,
			want:    "couldn't be resolved",
		},
		{
			name:    "client-go: request budget exhausted",
			message: `client rate limiter Wait returned an error: context deadline exceeded`,
			want:    "timed out",
		},
		{
			name:    "apiserver: RBAC denial",
			message: `applications.argoproj.io is forbidden: User "system:serviceaccount:argocd:default" cannot list resource "applications"`,
			want:    "Permission was denied",
		},
		{
			name:    "kubectl: missing context",
			message: `error: context "k3d-missing" does not exist`,
			want:    "kube-context doesn't exist",
		},
		{
			name:    "docker: daemon not started",
			message: `Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?`,
			want:    "Docker doesn't appear to be running",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := friendlyHint(fmt.Errorf("operation failed: %s", tc.message))
			if !strings.Contains(got, tc.want) {
				t.Errorf("no hint for a real upstream message.\nmessage: %s\nwant hint containing: %q\ngot: %q\n"+
					"If the tool reworded its error, update the patterns in friendly.go.", tc.message, tc.want, got)
			}
		})
	}
}

// TestFriendlyHint_OwnershipBeatsTimeout pins the documented precedence: the
// ownership failure often surfaces inside a message that also mentions a
// timeout, and the ownership hint is the actionable one.
func TestFriendlyHint_OwnershipBeatsTimeout(t *testing.T) {
	err := fmt.Errorf("timed out waiting for the condition: invalid ownership metadata")

	if got := friendlyHint(err); !strings.Contains(got, "Helm ownership metadata") {
		t.Errorf("the ownership hint must win over the generic timeout hint; got %q", got)
	}
}

// TestFriendlyHint_NoHintForUnknownErrors: an unrecognized failure must produce
// no hint rather than a misleading one.
func TestFriendlyHint_NoHintForUnknownErrors(t *testing.T) {
	if got := friendlyHint(fmt.Errorf("chart values are malformed at line 12")); got != "" {
		t.Errorf("expected no hint for an unrecognized error, got %q", got)
	}
}

// TestFriendlyHint_PendingReleaseSuggestsRollback (V4): a helm release left in
// pending-* by an interrupted operation must point at rollback, NOT the generic
// "wait and retry" timeout hint — retrying hits the same wedged release. The
// pending case is ordered before the timeout case for exactly this reason.
func TestFriendlyHint_PendingReleaseSuggestsRollback(t *testing.T) {
	cases := []string{
		`Error: UPGRADE FAILED: another operation (install/upgrade/rollback) is in progress`,
		`Error: release app-of-apps failed, status: pending-upgrade`,
		`cannot patch: release in pending-install state`,
		`Error: release argo-cd failed, status: pending-rollback`,
	}
	for _, msg := range cases {
		got := friendlyHint(fmt.Errorf("op failed: %s", msg))
		if !strings.Contains(got, "rollback") {
			t.Errorf("pending-release must suggest rollback.\nmessage: %s\ngot: %q", msg, got)
		}
		if strings.Contains(got, "cluster may be slow or unreachable") {
			t.Errorf("pending-release must NOT get the generic timeout hint (retry is wrong here).\nmessage: %s\ngot: %q", msg, got)
		}
	}
}
