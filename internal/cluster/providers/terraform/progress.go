package terraform

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/pterm/pterm"
)

// Terraform's -json output is one JSON object per line (machine-readable UI
// protocol). progressWriter turns the interesting ones into short pterm lines
// so long applies show per-resource progress; pterm printers respect --silent.

// applyEvent is the subset of the terraform JSON-UI schema the writer reads.
type applyEvent struct {
	Level   string `json:"@level"`
	Message string `json:"@message"`
	Type    string `json:"type"`
}

// progressLine converts one JSON event line into a display string; ok=false
// means the event is noise (refresh/progress ticks) and prints nothing.
// Verbose forwards every event message instead.
func progressLine(line []byte, verbose bool) (string, bool) {
	var ev applyEvent
	if err := json.Unmarshal(line, &ev); err != nil {
		return "", false // not an event line (e.g. blank) — drop
	}
	if verbose {
		return ev.Message, ev.Message != ""
	}
	switch ev.Type {
	// planned_change lines ("...: Plan to create") show WHAT the apply is
	// about to do — the plan detail, not just its change_summary count.
	case "planned_change", "apply_start", "apply_complete", "apply_errored", "change_summary":
		return ev.Message, ev.Message != ""
	case "diagnostic":
		// Errors surface through the returned error too, but printing them in
		// stream order preserves the context of what was being created.
		return ev.Message, ev.Level == "error" && ev.Message != ""
	default:
		return "", false
	}
}

// progressWriter buffers partial writes into lines and prints each event.
type progressWriter struct {
	verbose bool
	buf     bytes.Buffer
}

func newProgressWriter(verbose bool) io.Writer {
	return &progressWriter{verbose: verbose}
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		line, err := w.buf.ReadBytes('\n')
		if err != nil {
			// No full line yet — keep the partial for the next Write.
			w.buf.Write(line)
			break
		}
		if msg, ok := progressLine(bytes.TrimSpace(line), w.verbose); ok {
			pterm.DefaultBasicText.Printf("  %s\n", msg)
		}
	}
	return len(p), nil
}
