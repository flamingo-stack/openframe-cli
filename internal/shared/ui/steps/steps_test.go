package steps

import (
	"strings"
	"testing"
	"time"
)

func TestTracker_TimingsAndTotal(t *testing.T) {
	tr := NewTracker("one", "two")
	tr.Begin(0, "")
	tr.Done(0)
	tr.Begin(1, "detail")
	tr.Fail(1)

	timings := tr.Timings()
	if len(timings) != 2 {
		t.Fatalf("timings = %+v", timings)
	}
	if timings[0].Title != "one" || timings[0].Failed {
		t.Fatalf("first timing = %+v", timings[0])
	}
	if timings[1].Title != "two" || !timings[1].Failed {
		t.Fatalf("second timing = %+v", timings[1])
	}
	if tr.Total() <= 0 {
		t.Fatal("total must be positive")
	}

	// Out-of-range and never-begun stages are ignored, not panics.
	tr.Begin(9, "")
	tr.Done(9)
	tr.Done(-1)
	if len(tr.Timings()) != 2 {
		t.Fatal("out-of-range stages must not record timings")
	}
}

func TestTimingsLine(t *testing.T) {
	line := TimingsLine([]Timing{
		{Title: "validate", Duration: 100 * time.Millisecond},
		{Title: "cluster", Duration: 42 * time.Second},
	})
	if !strings.Contains(line, "validate 0.1s") || !strings.Contains(line, "cluster 42s") {
		t.Fatalf("line = %q", line)
	}
}

func TestFormatDuration(t *testing.T) {
	for in, want := range map[time.Duration]string{
		800 * time.Millisecond:          "0.8s",
		42 * time.Second:                "42s",
		17*time.Minute + 20*time.Second: "17m20s",
		time.Hour + 2*time.Minute:       "1h02m",
	} {
		if got := FormatDuration(in); got != want {
			t.Errorf("FormatDuration(%v) = %q, want %q", in, got, want)
		}
	}
}
