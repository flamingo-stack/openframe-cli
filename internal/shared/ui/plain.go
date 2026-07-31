package ui

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/pterm/pterm"
)

// plainMode records --plain: keep the output sequential — no spinners, no
// in-place areas, no self-rewriting progress lines — while keeping colors.
// For terminal users running the CLI under wrappers (watch, script, tmux
// panes being logged) where cursor rewriting garbles the capture; non-TTY
// output degrades the same way automatically, --plain just makes it explicit.
var plainMode bool

// SetPlain enables plain mode for the process (called from the global flag
// hook, not meant to be reversed).
func SetPlain() { plainMode = true }

// IsPlain reports whether --plain sequential output was requested.
func IsPlain() bool { return plainMode }

// Animated reports whether in-place terminal animation (spinners, live
// areas, self-rewriting progress) is appropriate: an interactive terminal,
// not --plain, not --silent.
func Animated() bool { return IsTerminal() && !plainMode && !IsSilent() }

// ApplyColorContract honors the standard color environment variables:
// NO_COLOR (any non-empty value) strips all styling, unless CLICOLOR_FORCE=1
// insists. ANSI stays ON for plain non-TTY output by default — modern CI log
// renderers (GitHub Actions among them) display it, and NO_COLOR is the
// explicit opt-out for consumers that don't.
func ApplyColorContract() {
	if os.Getenv("CLICOLOR_FORCE") == "1" {
		return
	}
	if os.Getenv("NO_COLOR") != "" {
		pterm.DisableStyling()
	}
}

// NewTimestampWriter wraps w so every line starts with a wall-clock stamp —
// applied to the debug printer under --verbose, where lines exist to be
// correlated with cluster events ("which pod restarted while helm hung?")
// and were previously impossible to place in time.
func NewTimestampWriter(w io.Writer) io.Writer {
	return &timestampWriter{out: w, atLineStart: true}
}

type timestampWriter struct {
	out         io.Writer
	atLineStart bool
}

func (t *timestampWriter) Write(p []byte) (int, error) {
	var b bytes.Buffer
	for _, c := range p {
		if t.atLineStart && c != '\n' {
			b.WriteString(time.Now().Format("15:04:05.000 "))
			t.atLineStart = false
		}
		b.WriteByte(c)
		if c == '\n' {
			t.atLineStart = true
		}
	}
	if _, err := t.out.Write(b.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}
