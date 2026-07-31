package download

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
)

// Download progress: a single self-rewriting line on stderr —
//
//	⤓ helm-v3.16.2.tar.gz  ██████████░░░░░░ 62% · 8.4 MB/s
//
// shown only where it can't do harm: an interactive terminal, not --silent,
// and no spinner currently animating (both rewrite the same cursor line and
// would garble each other — the spinner already tells that flow's story).

const progressRenderInterval = 120 * time.Millisecond

// progressEnabled gates the bar to safe contexts (see package comment above).
func progressEnabled() bool {
	return ui.Animated() && !spinner.AnyActive()
}

// announceEnabled gates the sequential begin/done lines used where the live
// bar cannot run (non-TTY, --plain): a 50 MB download used to be complete
// silence there. When a spinner is animating, its own text already covers
// the story; --silent stays silent.
func announceEnabled() bool {
	return !ui.IsSilent() && !spinner.AnyActive() && !progressEnabled()
}

// progressReader wraps a download body and renders the progress line as the
// bytes flow through. total < 0 means unknown length (no percentage, just
// volume and speed). Rendering goes to stderr so stdout stays machine-clean.
type progressReader struct {
	r          io.Reader
	label      string
	total      int64
	read       int64
	start      time.Time
	lastRender time.Time
	out        io.Writer
	rendered   bool
}

func newProgressReader(r io.Reader, label string, total int64) *progressReader {
	return &progressReader{r: r, label: label, total: total, start: time.Now(), out: os.Stderr}
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.read += int64(n)
	if now := time.Now(); now.Sub(p.lastRender) >= progressRenderInterval {
		p.lastRender = now
		p.render()
	}
	return n, err
}

func (p *progressReader) render() {
	p.rendered = true
	speed := humanBytes(int64(float64(p.read)/time.Since(p.start).Seconds())) + "/s"
	g := ui.Glyphs()
	if p.total > 0 {
		frac := float64(p.read) / float64(p.total)
		fmt.Fprintf(p.out, "\r\033[K  %s %s  %s %3.0f%% %s %s",
			g.Arrow, p.label, ui.ProgressBar(frac, 20), frac*100, g.Bullet, speed)
		return
	}
	fmt.Fprintf(p.out, "\r\033[K  %s %s  %s %s %s", g.Arrow, p.label, humanBytes(p.read), g.Bullet, speed)
}

// done clears the progress line so the caller's own output starts clean.
func (p *progressReader) done() {
	if p.rendered {
		fmt.Fprint(p.out, "\r\033[K")
	}
}

// announceStart/announceDone are the sequential begin/done pair for contexts
// without the live bar. Duration-only on done — the byte count already ran in
// the start line when the server sent a length.
func announceStart(label string, total int64) {
	if total > 0 {
		pterm.Info.Printf("Downloading %s (%s)...\n", label, humanBytes(total))
		return
	}
	pterm.Info.Printf("Downloading %s...\n", label)
}

func announceDone(label string, took time.Duration) {
	pterm.Info.Printf("Downloaded %s in %s\n", label, took.Round(100*time.Millisecond))
}

// humanBytes renders a byte count for humans: 512 B, 3.4 MB, 1.2 GB.
func humanBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// assetLabel is the human name of a download: the URL path's base name.
func assetLabel(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		if base := path.Base(u.Path); base != "." && base != "/" {
			return base
		}
	}
	return "download"
}
