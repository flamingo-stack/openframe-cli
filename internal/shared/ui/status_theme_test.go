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

// A message printed repeatedly (the ArgoCD wait re-prints its stuck-app
// summary every few minutes) must annotate only once: the runner echoes every
// workflow command inline in the log, and GitHub keeps just 10 annotations per
// step, so repeats both double the log and crowd out new warnings.
func TestAnnotationWriter_DeduplicatesRepeats(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	aw := newAnnotationWriter(io.Discard, "warning")
	for range 3 {
		_, _ = aw.Write([]byte("warning  Stuck app tenant: health=Degraded\n"))
	}
	_, _ = aw.Write([]byte("warning  Stuck app mysql: health=Progressing\n"))

	_ = w.Close()
	os.Stdout = old
	var sb strings.Builder
	_, _ = io.Copy(&sb, r)

	got := sb.String()
	assert.Equal(t, 1, strings.Count(got, "Stuck app tenant"))
	assert.Equal(t, 1, strings.Count(got, "Stuck app mysql"))
}

// Exactly one severity marker is stripped, in both forms pterm prints it:
// the padded/styled tag ("warning  msg") and RawOutput's "warning: msg"
// (NO_COLOR). Stripping must not cascade into message text.
func TestAnnotationWriter_PrefixStripping(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")

	cases := []struct {
		line string
		want string
	}{
		// RawOutput (NO_COLOR) form: no leading colon may leak through.
		{"warning: disk almost full\n", "::warning title=openframe::disk almost full\n"},
		// A message that itself starts with "error..." must survive intact.
		{"warning  errors found in 3 charts\n", "::warning title=openframe::errors found in 3 charts\n"},
	}
	for _, tc := range cases {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		aw := newAnnotationWriter(io.Discard, "warning")
		_, _ = aw.Write([]byte(tc.line))

		_ = w.Close()
		os.Stdout = old
		var sb strings.Builder
		_, _ = io.Copy(&sb, r)
		assert.Equal(t, tc.want, sb.String(), "line %q", tc.line)
	}
}
