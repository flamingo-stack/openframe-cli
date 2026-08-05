package ui

import (
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

// ApplyStatusPrefixTheme restyles the package-level pterm status printers that
// the whole CLI prints through (pterm.Info/Warning/Error/Success/Debug), so a
// single call here re-themes every call site.
//
// Interactive terminals get quiet glyph prefixes from the shared GlyphSet
// (which already degrades to ASCII under OPENFRAME_ASCII/TERM=dumb): a dim
// bullet for Info, ▲/✖/✔ for Warning/Error/Success — no background badges.
// Non-interactive output (CI, pipes, --plain) keeps word tags for
// grep-ability, but lowercase, column-aligned, and foreground-colored instead
// of the block badges; the words themselves stay info/warning/error/success.
//
// Inside GitHub Actions, Warning and Error additionally emit
// ::warning::/::error:: workflow commands so they surface as job/PR
// annotations instead of sinking into the log.
//
// It mutates pterm's package-level printers (like SetSilent), so it runs once,
// early, from ApplyGlobalOutputFlags — after the --plain/--silent flags are
// applied, because both the mode choice and the silent writers must win.
// pterm's With* helpers copy the printer struct, so composing with SetSilent's
// io.Discard writers and the --verbose timestamp writer is order-safe either
// way; running last just keeps the reasoning simple.
func ApplyStatusPrefixTheme() {
	interactive := IsTerminal() && !IsPlain()

	type look struct {
		printer *pterm.PrefixPrinter
		glyph   string // interactive prefix
		word    string // non-interactive prefix, column-aligned
		style   *pterm.Style
	}
	g := Glyphs()
	looks := []look{
		{&pterm.Info, g.Bullet, "info   ", pterm.NewStyle(pterm.FgGray)},
		{&pterm.Warning, g.Warn, "warning", pterm.NewStyle(pterm.FgLightYellow)},
		{&pterm.Error, g.Fail, "error  ", pterm.NewStyle(pterm.FgLightRed)},
		{&pterm.Success, g.OK, "success", pterm.NewStyle(pterm.FgLightGreen)},
		{&pterm.Debug, g.Bullet, "debug  ", pterm.NewStyle(pterm.FgGray)},
	}
	for _, l := range looks {
		text := l.word
		if interactive {
			text = l.glyph
		}
		*l.printer = *l.printer.WithPrefix(pterm.Prefix{Text: text, Style: l.style})
	}
	// Info's non-interactive tag stays cyan (its message color family), not the
	// gray the interactive bullet uses — in a colorless-context log the tag is
	// the only severity signal.
	if !interactive {
		pterm.Info = *pterm.Info.WithPrefix(pterm.Prefix{Text: "info   ", Style: pterm.NewStyle(pterm.FgLightCyan)})
	}

	// Surface warnings as Actions annotations. Errors are NOT teed here: the
	// shared error handler already emits a richer ::error:: annotation
	// (headline as title, cause as message) for every command failure, and a
	// second writer-level annotation would duplicate it. The tee respects
	// --silent, which discarded Warning's writer above — silent means errors
	// only, annotations included.
	if InGitHubActions() && !IsSilent() {
		pterm.Warning = *pterm.Warning.WithWriter(newAnnotationWriter(pterm.Warning.Writer, "warning"))
	}
}

// annotationWriter tees a status printer's output into a GitHub Actions
// ::warning:: or ::error:: workflow command, with ANSI styling and the
// printer's own prefix column stripped.
//
// Each distinct message is annotated ONCE per process: the runner echoes every
// workflow command inline in the log ("Warning: …"), so re-annotating a
// repeating message (the ArgoCD wait re-prints its stuck-app summary every few
// minutes) would double a growing share of the log — and GitHub keeps only 10
// annotations per step, so repeats also crowd out genuinely new warnings.
type annotationWriter struct {
	inner io.Writer
	level string
	mu    sync.Mutex
	seen  map[string]struct{}
}

func newAnnotationWriter(inner io.Writer, level string) io.Writer {
	if inner == nil {
		inner = os.Stdout
	}
	return &annotationWriter{inner: inner, level: level, seen: make(map[string]struct{})}
}

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (a *annotationWriter) Write(p []byte) (int, error) {
	n, err := a.inner.Write(p)
	msg := strings.TrimSpace(ansiSeq.ReplaceAllString(string(p), ""))
	// Drop the prefix column ("warning"/"error  "/▲/✖) — the annotation level
	// already carries the severity.
	for _, prefix := range []string{"warning", "error", Glyphs().Warn, Glyphs().Fail} {
		msg = strings.TrimSpace(strings.TrimPrefix(msg, prefix))
	}
	if msg == "" {
		return n, err
	}
	a.mu.Lock()
	_, dup := a.seen[msg]
	if !dup {
		a.seen[msg] = struct{}{}
	}
	a.mu.Unlock()
	if dup {
		return n, err
	}
	// Reuse the escaped emitters so runner parsing rules live in one place.
	if a.level == "warning" {
		WarningAnnotation("openframe", msg)
	} else {
		ErrorAnnotation("openframe", msg)
	}
	return n, err
}
