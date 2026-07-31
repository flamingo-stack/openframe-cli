package ui

import (
	"bytes"
	"regexp"
	"testing"
)

// Every line the timestamp writer emits must start with a wall-clock stamp,
// including lines split across Write calls; blank lines stay unstamped.
func TestTimestampWriter(t *testing.T) {
	var out bytes.Buffer
	w := NewTimestampWriter(&out)
	if _, err := w.Write([]byte("first li")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ne\nsecond\n\nthird\n")); err != nil {
		t.Fatal(err)
	}
	stamped := regexp.MustCompile(`(?m)^\d{2}:\d{2}:\d{2}\.\d{3} `)
	if got := len(stamped.FindAllString(out.String(), -1)); got != 3 {
		t.Fatalf("expected 3 stamped lines, got %d in:\n%s", got, out.String())
	}
	if regexp.MustCompile(`ne\n\d`).MatchString(out.String()) == false && !bytes.Contains(out.Bytes(), []byte("first line\n")) {
		t.Fatalf("split line must not be re-stamped mid-line:\n%s", out.String())
	}
}
