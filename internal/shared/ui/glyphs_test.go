package ui

import (
	"strings"
	"testing"
)

func TestUnicodeCapable(t *testing.T) {
	clear := func(t *testing.T) {
		t.Helper()
		for _, k := range []string{"OPENFRAME_ASCII", "TERM", "LC_ALL", "LC_CTYPE", "LANG"} {
			t.Setenv(k, "")
		}
	}

	t.Run("default is unicode (CI has no locale set)", func(t *testing.T) {
		clear(t)
		if !unicodeCapable() {
			t.Fatal("no signals → unicode")
		}
	})
	t.Run("TERM=dumb opts out", func(t *testing.T) {
		clear(t)
		t.Setenv("TERM", "dumb")
		if unicodeCapable() {
			t.Fatal("dumb terminal must degrade to ASCII")
		}
	})
	t.Run("OPENFRAME_ASCII=1 opts out", func(t *testing.T) {
		clear(t)
		t.Setenv("OPENFRAME_ASCII", "1")
		if unicodeCapable() {
			t.Fatal("explicit opt-out must degrade to ASCII")
		}
	})
	t.Run("non-UTF8 locale opts out", func(t *testing.T) {
		clear(t)
		t.Setenv("LANG", "C")
		if unicodeCapable() {
			t.Fatal("LANG=C must degrade to ASCII")
		}
	})
	t.Run("UTF-8 locale is unicode", func(t *testing.T) {
		clear(t)
		t.Setenv("LANG", "en_US.UTF-8")
		if !unicodeCapable() {
			t.Fatal("UTF-8 locale must keep unicode")
		}
	})
}

func TestGlyphsFor(t *testing.T) {
	if g := glyphsFor(true); g.OK != "✔" || g.Fail != "✖" {
		t.Fatalf("unicode set = %+v", g)
	}
	// The ASCII set must be pure 7-bit so TERM=dumb logs never see mojibake.
	g := glyphsFor(false)
	for _, s := range []string{g.OK, g.Fail, g.Warn, g.Pending, g.Running, g.Bullet, g.Arrow, g.Bar, g.BarEmpty} {
		for _, r := range s {
			if r > 127 {
				t.Fatalf("ASCII set contains non-ASCII %q", s)
			}
		}
	}
}

func TestProgressBar(t *testing.T) {
	g := glyphsFor(unicodeCapable())
	bar := ProgressBar(0.6, 10)
	if strings.Count(bar, g.Bar) != 6 || strings.Count(bar, g.BarEmpty) != 4 {
		t.Fatalf("0.6×10 = %q", bar)
	}
	if ProgressBar(-1, 5) != strings.Repeat(g.BarEmpty, 5) {
		t.Fatal("negative fraction must clamp to empty")
	}
	if ProgressBar(2, 5) != strings.Repeat(g.Bar, 5) {
		t.Fatal("fraction above 1 must clamp to full")
	}
	if ProgressBar(0.5, 0) != "" {
		t.Fatal("zero width → empty string")
	}
}

func TestHyperlink_FallsBackOutsideTerminal(t *testing.T) {
	// Tests never run on a TTY, so the fallback path is deterministic here.
	if got := Hyperlink("https://x.dev", "docs"); got != "docs (https://x.dev)" {
		t.Fatalf("got %q", got)
	}
	if got := Hyperlink("https://x.dev", "https://x.dev"); got != "https://x.dev" {
		t.Fatalf("same text and url must not duplicate: %q", got)
	}
	if got := Hyperlink("", "docs"); got != "docs" {
		t.Fatalf("no url → plain text, got %q", got)
	}
}
