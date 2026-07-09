package spinner

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinner_SuccessPrintsLine(t *testing.T) {
	var buf bytes.Buffer
	s := NewWithWriter(&buf)
	s.Start("working")
	s.Success("done") // rendered via pterm's styled SUCCESS printer
	assert.Contains(t, buf.String(), "done")
}

func TestSpinner_FailPrintsLine(t *testing.T) {
	var buf bytes.Buffer
	s := NewWithWriter(&buf)
	s.Start("working")
	s.Fail("nope") // rendered via pterm's styled ERROR printer
	assert.Contains(t, buf.String(), "nope")
}

func TestSpinner_StopIsQuietAndIdempotent(t *testing.T) {
	var buf bytes.Buffer
	s := NewWithWriter(&buf)
	s.Start("working")
	s.Stop()
	s.Stop()          // second stop must be a no-op, not a panic/hang
	s.Success("late") // after stop → no-op
	assert.Empty(t, buf.String(), "plain Stop prints nothing; calls after stop are no-ops")
}

func TestSpinner_StartTwiceJustUpdatesText(t *testing.T) {
	s := NewWithWriter(io.Discard)
	s.Start("first")
	s.Start("second") // must not spawn a second goroutine
	s.mu.Lock()
	assert.Equal(t, "second", s.text)
	s.mu.Unlock()
	s.Stop()
}

// TestSpinner_ConcurrentUpdateAndStopRaceFree is the point of this package: run
// the animation goroutine while many goroutines update the text, then Stop.
// Under `go test -race` this must be clean (Stop joins the animation goroutine).
func TestSpinner_ConcurrentUpdateAndStopRaceFree(t *testing.T) {
	s := NewWithWriter(io.Discard)
	s.interval = time.Millisecond // tick fast so animate reads text during the test
	s.Start("initial")

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				s.UpdateText(fmt.Sprintf("update-%d-%d", n, j))
			}
		}(i)
	}
	wg.Wait()

	done := make(chan struct{})
	go func() { s.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return — animation goroutine was not joined")
	}
}
