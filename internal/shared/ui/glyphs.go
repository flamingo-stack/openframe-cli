package ui

import (
	"math"
	"os"
	"strings"
	"sync"
)

// GlyphSet is the terminal symbol vocabulary used by the richer UI surfaces
// (step trees, dashboards, plan diffs). Rich Unicode by default — every CI
// worth its salt renders UTF-8 — degrading to plain ASCII only where Unicode
// demonstrably breaks: TERM=dumb, an explicitly non-UTF-8 locale, or an
// OPENFRAME_ASCII=1 opt-out.
type GlyphSet struct {
	OK       string // finished successfully
	Fail     string // finished with an error
	Warn     string // finished with a warning
	Pending  string // not started yet
	Running  string // in progress (static marker; spinners animate separately)
	Bullet   string // list/metadata separator
	Arrow    string // "leads to"
	Bar      string // filled progress-bar cell
	BarEmpty string // empty progress-bar cell
}

var (
	glyphsOnce sync.Once
	glyphs     GlyphSet
)

// Glyphs returns the process-wide glyph set, resolved once.
func Glyphs() GlyphSet {
	glyphsOnce.Do(func() { glyphs = glyphsFor(unicodeCapable()) })
	return glyphs
}

func glyphsFor(unicode bool) GlyphSet {
	if unicode {
		return GlyphSet{OK: "✔", Fail: "✖", Warn: "▲", Pending: "○", Running: "◉", Bullet: "·", Arrow: "→", Bar: "█", BarEmpty: "░"}
	}
	return GlyphSet{OK: "ok", Fail: "x", Warn: "!", Pending: "-", Running: "*", Bullet: "-", Arrow: "->", Bar: "#", BarEmpty: "."}
}

// unicodeCapable reports whether the terminal can be trusted with Unicode
// glyphs. The default is yes — the opt-outs are explicit signals only, so CI
// logs (which usually have no locale set at all) keep the rich glyphs.
func unicodeCapable() bool {
	if os.Getenv("OPENFRAME_ASCII") == "1" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	for _, v := range []string{os.Getenv("LC_ALL"), os.Getenv("LC_CTYPE"), os.Getenv("LANG")} {
		if v == "" {
			continue
		}
		up := strings.ToUpper(v)
		return strings.Contains(up, "UTF-8") || strings.Contains(up, "UTF8")
	}
	return true
}

// ProgressBar renders a textual progress bar of the given cell width, e.g.
// "██████░░░░" for 0.6×10. fraction is clamped to [0, 1].
func ProgressBar(fraction float64, width int) string {
	if width <= 0 {
		return ""
	}
	f := math.Max(0, math.Min(1, fraction))
	filled := int(math.Round(f * float64(width)))
	g := Glyphs()
	return strings.Repeat(g.Bar, filled) + strings.Repeat(g.BarEmpty, width-filled)
}
