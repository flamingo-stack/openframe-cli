package argocd

import (
	"strings"
	"testing"
	"time"
)

// TestTimeoutError_NamesTheStuckApplications (M3.2): the wait loop knows which
// applications never became ready. The old message threw that away and said
// only "timeout waiting for ArgoCD applications after 1h0m0s", leaving the user
// to go find the stuck app by hand.
func TestTimeoutError_NamesTheStuckApplications(t *testing.T) {
	err := timeoutError(30*time.Minute, 4, 6, []string{"openframe-api", "openframe-ui"})

	msg := err.Error()
	for _, want := range []string{
		"30m0s",                              // how long it waited
		"4/6 ready",                          // how far it got
		"openframe-api",                      // which apps are stuck
		"openframe-ui",                       //
		"kubectl get applications -n argocd", // what to run next
		"kubectl describe application openframe-api -n argocd",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout error must contain %q; got:\n%s", want, msg)
		}
	}
}

// TestTimeoutError_BoundsTheApplicationList: a large platform can leave dozens
// of applications pending. The list must not bury the next-step hint.
func TestTimeoutError_BoundsTheApplicationList(t *testing.T) {
	var many []string
	for i := 0; i < 25; i++ {
		many = append(many, "app-"+string(rune('a'+i)))
	}

	msg := timeoutError(time.Minute, 0, 25, many).Error()

	if !strings.Contains(msg, "and 15 more") {
		t.Errorf("the list must be truncated with a count of the remainder; got:\n%s", msg)
	}
	if strings.Contains(msg, "app-y") {
		t.Errorf("the 25th application must not be listed; got:\n%s", msg)
	}
	if !strings.Contains(msg, "kubectl get applications") {
		t.Errorf("the next-step hint must survive truncation; got:\n%s", msg)
	}
}

// TestTimeoutError_NoAppsIsStillLegible: timing out before any application was
// observed (app-of-apps never produced children) must not print an empty list.
func TestTimeoutError_NoAppsIsStillLegible(t *testing.T) {
	msg := timeoutError(time.Minute, 0, 0, nil).Error()

	if strings.Contains(msg, "still not ready:") {
		t.Errorf("an empty list must be omitted, not printed empty; got:\n%s", msg)
	}
	if strings.Contains(msg, "describe application") {
		t.Errorf("there is no application to describe; got:\n%s", msg)
	}
	if !strings.Contains(msg, "timeout after 1m0s") {
		t.Errorf("the message must still state the timeout; got:\n%s", msg)
	}
}
