// Package steps renders a buildkit-style stage checklist for multi-stage
// flows (bootstrap: validate → cluster → platform):
//
//	▶ [2/3] Create cluster k3d-openframe-dev
//	  ... the stage's own output scrolls here ...
//	✔ [2/3] Create cluster k3d-openframe-dev · 42s
//
// Stage lines are plain sequential prints, never in-place rewrites: the
// stages run spinners and scrolling output of their own, so a rewriting
// tracker would fight them for the cursor. That also makes the output
// CI-log friendly for free. Durations are recorded for the final summary.
package steps

import (
	"fmt"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// Timing is one finished stage's outcome for the final summary.
type Timing struct {
	Title    string
	Duration time.Duration
	Failed   bool
}

// Tracker tracks a fixed sequence of stages. Not safe for concurrent use —
// stages are sequential by design.
type Tracker struct {
	titles  []string
	started []time.Time
	timings []Timing
	begun   time.Time
}

// NewTracker creates a tracker over the given ordered stage titles.
func NewTracker(titles ...string) *Tracker {
	return &Tracker{
		titles:  titles,
		started: make([]time.Time, len(titles)),
		begun:   time.Now(),
	}
}

// Begin announces stage i (0-based). detail, when non-empty, is appended dim.
func (t *Tracker) Begin(i int, detail string) {
	if i < 0 || i >= len(t.titles) {
		return
	}
	t.started[i] = time.Now()
	line := fmt.Sprintf("%s %s %s", ui.Glyphs().Running, t.progress(i), t.titles[i])
	if detail != "" {
		line += " " + pterm.FgGray.Sprint(detail)
	}
	pterm.DefaultBasicText.Println(line)
}

// Done closes stage i successfully, printing its duration.
func (t *Tracker) Done(i int) { t.finish(i, false) }

// Fail closes stage i as failed, printing its duration.
func (t *Tracker) Fail(i int) { t.finish(i, true) }

func (t *Tracker) finish(i int, failed bool) {
	if i < 0 || i >= len(t.titles) || t.started[i].IsZero() {
		return
	}
	d := time.Since(t.started[i]).Round(100 * time.Millisecond)
	t.timings = append(t.timings, Timing{Title: t.titles[i], Duration: d, Failed: failed})
	g := ui.Glyphs()
	if failed {
		pterm.DefaultBasicText.Printf("%s %s %s %s %s\n",
			pterm.FgRed.Sprint(g.Fail), t.progress(i), t.titles[i], g.Bullet, FormatDuration(d))
		return
	}
	pterm.DefaultBasicText.Printf("%s %s %s %s %s\n",
		pterm.FgGreen.Sprint(g.OK), t.progress(i), t.titles[i], g.Bullet, FormatDuration(d))
}

func (t *Tracker) progress(i int) string {
	return pterm.FgGray.Sprintf("[%d/%d]", i+1, len(t.titles))
}

// Timings returns the finished stages in completion order.
func (t *Tracker) Timings() []Timing { return t.timings }

// Total is the wall-clock time since the tracker was created.
func (t *Tracker) Total() time.Duration { return time.Since(t.begun) }

// TimingsLine renders "validate 0.1s · cluster 42s · platform 17m20s" for a
// summary row.
func TimingsLine(timings []Timing) string {
	g := ui.Glyphs()
	out := ""
	for i, tm := range timings {
		if i > 0 {
			out += " " + g.Bullet + " "
		}
		out += fmt.Sprintf("%s %s", tm.Title, FormatDuration(tm.Duration))
	}
	return out
}

// FormatDuration renders durations the way humans read elapsed time: 0.8s,
// 42s, 17m20s, 1h02m.
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	}
}
