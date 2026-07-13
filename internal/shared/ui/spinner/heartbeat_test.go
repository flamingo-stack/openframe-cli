package spinner

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuf is a mutex-guarded buffer: the heartbeat writes from its goroutine
// while the test reads, so the race detector needs the access synchronized.
type syncBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// TestHeartbeat_EmitsUntilStopped: the heartbeat writes its label periodically
// and stops cleanly. This is the liveness a non-interactive `helm --wait`
// needs — without it CI logs and piped sessions saw nothing for minutes.
func TestHeartbeat_EmitsUntilStopped(t *testing.T) {
	var buf syncBuf
	hb := startHeartbeat(&buf, "still working", time.Millisecond)

	// Give it room for several ticks, then stop (which joins the goroutine).
	time.Sleep(50 * time.Millisecond)
	hb.Stop()

	out := buf.String()
	if !strings.Contains(out, "still working") {
		t.Fatalf("heartbeat must emit its label; got %q", out)
	}
	if !strings.Contains(out, "elapsed") {
		t.Errorf("heartbeat must report elapsed time; got %q", out)
	}
}

// TestHeartbeat_StopJoinsNoEmitAfter: after Stop returns, no further line may
// be written — Stop joins the goroutine (race-free teardown).
func TestHeartbeat_StopJoinsNoEmitAfter(t *testing.T) {
	var buf syncBuf
	hb := startHeartbeat(&buf, "tick", time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	hb.Stop()
	after := buf.String()
	time.Sleep(20 * time.Millisecond) // no writer is running now
	if buf.String() != after {
		t.Error("no emit may occur after Stop returns")
	}
}

// TestHeartbeat_StopIsIdempotentAndNoOpSafe: Stop is safe to call twice, and
// safe on the no-op Heartbeat returned under --silent.
func TestHeartbeat_StopIsIdempotentAndNoOpSafe(t *testing.T) {
	hb := startHeartbeat(&syncBuf{}, "x", time.Millisecond)
	hb.Stop()
	hb.Stop() // must not panic or block

	(&Heartbeat{}).Stop() // no-op heartbeat (silent path) must be safe
}
