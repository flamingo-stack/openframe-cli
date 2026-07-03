package services

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
)

// TestInstallChartsWithConfigContext_CancelledContextIsHonored proves the P3
// win: the install path respects its context. Previously the command layer
// passed context.Background(), so Ctrl-C could never cancel the install; now
// cmd.Context() (signal-cancelled at the root via ExecuteContext) flows through
// and a cancelled context short-circuits before any real work.
func TestInstallChartsWithConfigContext_CancelledContextIsHonored(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled, as if Ctrl-C was pressed

	err := InstallChartsWithConfigContext(ctx, types.InstallationRequest{})
	if err == nil {
		t.Fatal("expected a cancellation error from an already-cancelled context")
	}
	if !stderrors.Is(err, context.Canceled) {
		t.Fatalf("error must wrap context.Canceled, got %v", err)
	}
}
