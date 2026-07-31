package argocd

import (
	"context"
	"testing"
	"time"
)

// The wait's own budget must elapse BEFORE the caller's context deadline, or
// the diagnostic timeoutError (not-ready apps, per-app health) is unreachable
// and the user gets a bare "context deadline exceeded". The install path runs
// the whole flow under one 60m deadline while the wait default is also 60m
// measured from a later point — exactly that dead zone.
func TestCapToDeadline(t *testing.T) {
	from := time.Now()

	t.Run("no deadline keeps the configured budget", func(t *testing.T) {
		if got := capToDeadline(context.Background(), time.Hour, from); got != time.Hour {
			t.Fatalf("got %v, want 1h", got)
		}
	})

	t.Run("nearer deadline caps the budget with margin", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), from.Add(30*time.Minute))
		defer cancel()
		got := capToDeadline(ctx, time.Hour, from)
		if got <= 0 || got >= 30*time.Minute {
			t.Fatalf("got %v, want (0, 30m) — deadline minus margin", got)
		}
	})

	t.Run("farther deadline leaves the budget alone", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), from.Add(2*time.Hour))
		defer cancel()
		if got := capToDeadline(ctx, time.Hour, from); got != time.Hour {
			t.Fatalf("got %v, want 1h", got)
		}
	})

	t.Run("exhausted deadline yields zero, not negative", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), from.Add(10*time.Second))
		defer cancel()
		if got := capToDeadline(ctx, time.Hour, from); got != 0 {
			t.Fatalf("got %v, want 0", got)
		}
	})
}
