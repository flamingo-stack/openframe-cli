package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

// defaultHeartbeatInterval is how often a Heartbeat emits when no interval is
// given. Frequent enough that a CI log or a watching user sees liveness within
// a reasonable window; sparse enough not to flood a multi-minute wait.
const defaultHeartbeatInterval = 30 * time.Second

// Heartbeat periodically emits a one-line progress message while a long,
// output-less blocking operation runs. It is the NON-INTERACTIVE counterpart to
// the animated Spinner: the spinner needs a TTY to show anything, so in CI or a
// piped/`--non-interactive` session a multi-minute `helm --wait` printed
// nothing between "Installing..." and the final result — and users, assuming a
// hang, killed the process before the diagnostics ever printed.
//
// It writes to stderr so stdout stays clean for machine-readable output, and it
// is silent under --silent (whose contract is "errors only").
type Heartbeat struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

// StartHeartbeat begins emitting "<label> (<elapsed>)" to stderr every interval
// until Stop is called. interval <= 0 uses defaultHeartbeatInterval. Under
// --silent it is a no-op (Stop is still safe to call).
func StartHeartbeat(label string, interval time.Duration) *Heartbeat {
	if ui.IsSilent() {
		return &Heartbeat{} // no-op; Stop() tolerates nil channels
	}
	return startHeartbeat(os.Stderr, label, interval)
}

// startHeartbeat is the injectable core (writer + interval) behind
// StartHeartbeat, so tests can drive it with a buffer and a tiny interval.
func startHeartbeat(w io.Writer, label string, interval time.Duration) *Heartbeat {
	if interval <= 0 {
		interval = defaultHeartbeatInterval
	}
	h := &Heartbeat{stop: make(chan struct{}), done: make(chan struct{})}
	start := time.Now()
	go func() {
		defer close(h.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-h.stop:
				return
			case <-ticker.C:
				fmt.Fprintf(w, "%s (%s elapsed)\n", label, time.Since(start).Round(time.Second))
			}
		}
	}()
	return h
}

// Stop ends the heartbeat and waits for its goroutine to exit (join), so no
// emit can happen after Stop returns. Safe to call once; safe on a no-op
// Heartbeat (the --silent case).
func (h *Heartbeat) Stop() {
	if h.stop == nil {
		return
	}
	h.once.Do(func() {
		close(h.stop)
		<-h.done
	})
}
