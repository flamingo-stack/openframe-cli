package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyHint(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantSub string // substring the hint should contain ("" → expect no hint)
	}{
		{"nil", nil, ""},
		{"connection refused", errors.New("dial tcp 127.0.0.1:6550: connect: connection refused"), "isn't reachable"},
		{"server unreachable", errors.New("Unable to connect to the server: EOF"), "isn't reachable"},
		{"no such host", errors.New("lookup api.example.com: no such host"), "couldn't be resolved"},
		{"timeout", errors.New("context deadline exceeded"), "timed out"},
		// Finding 7: a Helm CRD-ownership failure must get its own actionable hint,
		// not the misleading "timed out / unreachable" one.
		{"crd ownership", errors.New(`CustomResourceDefinition "applications.argoproj.io" exists and cannot be imported into the current release: invalid ownership metadata`), "already exists without Helm ownership"},
		{"crd ownership beats timeout wording", errors.New(`operation timed out: invalid ownership metadata; missing key "meta.helm.sh/release-name"`), "already exists without Helm ownership"},
		{"permission denied", errors.New("pods is forbidden: User cannot list"), "Permission was denied"},
		{"missing context", errors.New(`context "k3d-foo" does not exist`), "kube-context doesn't exist"},
		{"docker down", errors.New("Cannot connect to the Docker daemon"), "Docker"},
		{"unknown error", errors.New("some totally unrelated failure"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hint := friendlyHint(tc.err)
			if tc.wantSub == "" {
				assert.Empty(t, hint)
			} else {
				assert.Contains(t, hint, tc.wantSub)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	assert.True(t, containsAny("hello world", "nope", "world"))
	assert.False(t, containsAny("hello world", "nope", "zzz"))
	assert.False(t, containsAny("hello"))
}

// selfDiagStub mimics an error that carries its own diagnosis (like the ArgoCD
// wait's stuck-app error embedding pod crash logs).
type selfDiagStub struct{ msg string }

func (s *selfDiagStub) Error() string       { return s.msg }
func (s *selfDiagStub) SelfDiagnosed() bool { return true }

// A self-diagnosed error must get NO generic hint: its embedded pod logs
// ("connect: connection refused" from a probe failure) previously misfired the
// "cluster isn't reachable" hint under a perfectly reachable cluster.
func TestFriendlyHint_SilentForSelfDiagnosed(t *testing.T) {
	err := &selfDiagStub{msg: "2 application(s) are Degraded...\n" +
		"  Unhealthy Pod/x: Startup probe failed: dial tcp 10.0.0.1:8096: connect: connection refused"}
	assert.Empty(t, friendlyHint(err), "self-diagnosed errors must suppress pattern-matched hints")

	// And the marker must survive wrapping (the chart/service layers wrap it).
	wrapped := &wrapErr{cause: err}
	assert.Empty(t, friendlyHint(wrapped), "the marker must be found through Unwrap chains")

	// A plain error with the same text still gets the hint — only the marker
	// changes behavior.
	plain := errors.New("dial tcp 10.0.0.1:8096: connect: connection refused")
	assert.Contains(t, friendlyHint(plain), "isn't reachable")
}

type wrapErr struct{ cause error }

func (w *wrapErr) Error() string { return "waiting failed: " + w.cause.Error() }
func (w *wrapErr) Unwrap() error { return w.cause }
