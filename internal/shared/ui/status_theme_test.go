package ui

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

// restorePrinters snapshots the package-level pterm printers the theme
// mutates and restores them on cleanup, so theme tests don't leak styling
// into other tests in the package.
func restorePrinters(t *testing.T) {
	t.Helper()
	info, warn, errP, succ, debug := pterm.Info, pterm.Warning, pterm.Error, pterm.Success, pterm.Debug
	t.Cleanup(func() {
		pterm.Info, pterm.Warning, pterm.Error, pterm.Success, pterm.Debug = info, warn, errP, succ, debug
	})
}

// Test env has no TTY on stdout, so the theme must pick the non-interactive
// look: lowercase, column-aligned word tags — the exact words are part of the
// log contract (grep -i WARNING must still hit).
func TestApplyStatusPrefixTheme_NonInteractive(t *testing.T) {
	restorePrinters(t)

	ApplyStatusPrefixTheme()

	assert.Equal(t, "info   ", pterm.Info.Prefix.Text)
	assert.Equal(t, "warning", pterm.Warning.Prefix.Text)
	assert.Equal(t, "error  ", pterm.Error.Prefix.Text)
	assert.Equal(t, "success", pterm.Success.Prefix.Text)
	assert.Equal(t, "debug  ", pterm.Debug.Prefix.Text)

	// All tags occupy one column so messages line up.
	for _, text := range []string{
		pterm.Info.Prefix.Text, pterm.Warning.Prefix.Text, pterm.Error.Prefix.Text,
		pterm.Success.Prefix.Text, pterm.Debug.Prefix.Text,
	} {
		assert.Len(t, text, 7)
	}
}

func TestAnnotationWriter_EmitsWorkflowCommand(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")

	// WarningAnnotation prints to os.Stdout; capture it.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	aw := newAnnotationWriter(io.Discard, "warning")
	_, err := aw.Write([]byte("\x1b[93mwarning\x1b[0m  disk almost full\n"))
	assert.NoError(t, err)

	_ = w.Close()
	os.Stdout = old
	var sb strings.Builder
	_, _ = io.Copy(&sb, r)

	got := sb.String()
	assert.Contains(t, got, "::warning title=openframe::disk almost full")
	// ANSI styling and the prefix column must not leak into the annotation.
	assert.NotContains(t, got, "\x1b[")
	assert.NotContains(t, got, "::warning title=openframe::warning")
}
