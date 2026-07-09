// Package spinner provides a small, race-free status spinner.
//
// It replaces direct use of pterm's SpinnerPrinter, whose internal animation
// goroutine races with its own Stop() (flagged by `go test -race`). Here we own
// the animation goroutine and, crucially, Stop() JOINS it (waits for it to
// exit) before returning — so no read of shared state can happen after teardown.
// All mutable state is guarded by a mutex.
package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// finalStyle selects how the spinner's final line is rendered.
type finalStyle int

const (
	styleNone finalStyle = iota
	styleSuccess
	styleFail
	styleWarning
	styleInfo
)

var defaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner is a race-free status spinner. The zero value is not usable; use New.
type Spinner struct {
	out      io.Writer
	isTTY    bool
	interval time.Duration
	frames   []string

	showTimer bool

	mu        sync.Mutex
	text      string
	active    bool
	startedAt time.Time
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// WithTimer makes the spinner show elapsed time next to its text.
func (s *Spinner) WithTimer() *Spinner {
	s.showTimer = true
	return s
}

// Start creates a spinner and immediately starts it with the given text — the
// one-line equivalent of New() followed by Start().
func Start(text string) *Spinner {
	s := New()
	s.Start(text)
	return s
}

// New returns a Spinner that writes to stdout (animated only on a real terminal).
func New() *Spinner {
	s := NewWithWriter(os.Stdout)
	if f, ok := any(os.Stdout).(*os.File); ok {
		s.isTTY = term.IsTerminal(int(f.Fd()))
	}
	return s
}

// NewWithWriter returns a Spinner writing to w (non-TTY: no animation frames).
// Used by tests to capture output.
func NewWithWriter(w io.Writer) *Spinner {
	return &Spinner{
		out:      w,
		interval: 100 * time.Millisecond,
		frames:   defaultFrames,
	}
}

// Start begins the spinner with the given text. Calling Start on an already
// running spinner just updates the text.
func (s *Spinner) Start(text string) {
	s.mu.Lock()
	if s.active {
		s.text = text
		s.mu.Unlock()
		return
	}
	s.text = text
	s.active = true
	s.startedAt = time.Now()
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go s.animate()
}

// UpdateText changes the spinner text while it is running.
func (s *Spinner) UpdateText(text string) {
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
}

// Stop stops the spinner without a final message.
func (s *Spinner) Stop() { s.finish("", styleNone) }

// Success stops the spinner and prints a styled success line (pterm SUCCESS box).
func (s *Spinner) Success(text string) { s.finish(text, styleSuccess) }

// Fail stops the spinner and prints a styled failure line (pterm ERROR box).
func (s *Spinner) Fail(text string) { s.finish(text, styleFail) }

// Warning stops the spinner and prints a styled warning line (pterm WARNING box).
func (s *Spinner) Warning(text string) { s.finish(text, styleWarning) }

// Info stops the spinner and prints a styled info line (pterm INFO box).
func (s *Spinner) Info(text string) { s.finish(text, styleInfo) }

// animate is the single writer of frames; it reads text under the mutex and
// exits promptly when signalled, closing doneCh so Stop can join.
func (s *Spinner) animate() {
	defer close(s.doneCh)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			text := s.text
			started := s.startedAt
			s.mu.Unlock()
			if s.isTTY {
				if s.showTimer {
					fmt.Fprintf(s.out, "\r%s %s (%s) ", s.frames[i%len(s.frames)], text, time.Since(started).Round(time.Second))
				} else {
					fmt.Fprintf(s.out, "\r%s %s ", s.frames[i%len(s.frames)], text)
				}
			}
		}
	}
}

// finish stops the animation goroutine, waits for it to exit (join), then prints
// the final line via pterm's styled printers (so the look matches the rest of
// the CLI). Joining before writing is what makes teardown race-free.
func (s *Spinner) finish(text string, style finalStyle) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	stopCh, doneCh := s.stopCh, s.doneCh
	s.mu.Unlock()

	close(stopCh) // tell animate to stop
	<-doneCh      // JOIN: animate has fully exited; no more reads happen

	if s.isTTY {
		fmt.Fprint(s.out, "\r\033[K") // clear the spinner line
	}

	switch style {
	case styleSuccess:
		pterm.Success.WithWriter(s.out).Println(text)
	case styleFail:
		pterm.Error.WithWriter(s.out).Println(text)
	case styleWarning:
		pterm.Warning.WithWriter(s.out).Println(text)
	case styleInfo:
		pterm.Info.WithWriter(s.out).Println(text)
	case styleNone:
		if text != "" {
			pterm.Info.WithWriter(s.out).Println(text)
		}
	}
}
