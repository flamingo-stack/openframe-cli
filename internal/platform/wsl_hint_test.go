package platform

import (
	"errors"
	"strings"
	"testing"
)

func TestWSLClusterHint_NilOffWindows(t *testing.T) {
	// The test suite runs on linux/macOS, so the guard must be a no-op there and
	// the native client-go path proceeds.
	if IsWindows() {
		t.Skip("host is Windows")
	}
	if err := WSLClusterHint("do a thing"); err != nil {
		t.Fatalf("off Windows WSLClusterHint must return nil, got %v", err)
	}
}

func TestWindowsWSLError_WrapsSentinelAndGuides(t *testing.T) {
	err := windowsWSLError("wait for ArgoCD")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, ErrWindowsNeedsWSL) {
		t.Error("must wrap ErrWindowsNeedsWSL so callers can detect it")
	}
	msg := err.Error()
	for _, want := range []string{"wait for ArgoCD", "wsl -d Ubuntu", "openframe", "inside WSL"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q:\n%s", want, msg)
		}
	}
}
