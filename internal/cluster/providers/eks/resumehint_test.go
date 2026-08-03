package eks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// The exact chain CreateCluster returns on an interrupted apply must satisfy
// BOTH: (a) errors.Is(context.Canceled), so the interruption handler runs, and
// (b) errors.As to a ResumeHint carrier, so the handler can surface the hint it
// otherwise drops with the swallowed error text. If either breaks, an
// interrupted create goes silent again (the GKE regression M2's twin).
func TestResumeHintError_SurvivesCancellationChain(t *testing.T) {
	applyErr := fmt.Errorf("terraform apply failed: %w", context.Canceled)
	hint := "The terraform state is kept in /ws/demo; re-run create to resume"
	opErr := models.NewClusterOperationError("create", "demo",
		withResumeHint(fmt.Errorf("%w\n%s", applyErr, hint), hint))

	if !errors.Is(opErr, context.Canceled) {
		t.Fatal("chain must still satisfy errors.Is(context.Canceled) so it is treated as an interruption")
	}

	var rh interface{ ResumeHint() string }
	if !errors.As(opErr, &rh) {
		t.Fatal("resume hint must be discoverable via errors.As despite the wrapping")
	}
	if !strings.Contains(rh.ResumeHint(), "re-run create to resume") {
		t.Fatalf("unexpected resume hint: %q", rh.ResumeHint())
	}
}

// A non-cancellation error path must not accidentally look interrupted.
func TestResumeHintError_UnwrapsToCause(t *testing.T) {
	cause := errors.New("quota exceeded")
	err := withResumeHint(cause, "state kept")
	if !errors.Is(err, cause) {
		t.Fatal("withResumeHint must keep the underlying cause reachable")
	}
	if errors.Is(err, context.Canceled) {
		t.Fatal("a non-cancellation cause must not be classified as cancelled")
	}
}
