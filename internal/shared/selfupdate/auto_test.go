package selfupdate

import (
	"context"
	"testing"
)

func TestAutoUpdateEnabled(t *testing.T) {
	cases := map[string]bool{"1": true, "true": true, "YES": true, "on": true, "": false, "0": false, "off": false}
	for val, want := range cases {
		t.Setenv(autoUpdateEnv, val)
		if got := AutoUpdateEnabled(); got != want {
			t.Errorf("AutoUpdateEnabled with %q = %v, want %v", val, got, want)
		}
	}
}

func TestSameMajor(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v1.2.3", "v1.9.0", true},
		{"v1.2.3", "v2.0.0", false},
		{"1.0.0", "v1.0.1", true},
		{"dev", "v1.0.0", false},
	}
	for _, c := range cases {
		if got := sameMajor(c.a, c.b); got != c.want {
			t.Errorf("sameMajor(%q,%q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// TestMaybeAutoUpdateGating confirms the cheap gates return early (no network):
// disabled by default, and suppressed when non-interactive even if opted in.
func TestMaybeAutoUpdateGating(t *testing.T) {
	t.Setenv(autoUpdateEnv, "") // not opted in
	if msg := MaybeAutoUpdate(context.Background(), "v1.0.0", true, nil); msg != "" {
		t.Fatalf("disabled auto-update should be a no-op, got %q", msg)
	}
	t.Setenv(autoUpdateEnv, "1") // opted in but non-interactive
	if msg := MaybeAutoUpdate(context.Background(), "v1.0.0", false, nil); msg != "" {
		t.Fatalf("non-interactive auto-update should be a no-op, got %q", msg)
	}
	// Dev build never auto-updates, even opted in + interactive.
	if msg := MaybeAutoUpdate(context.Background(), "dev", true, nil); msg != "" {
		t.Fatalf("dev build auto-update should be a no-op, got %q", msg)
	}
}
