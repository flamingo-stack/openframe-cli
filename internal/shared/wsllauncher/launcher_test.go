package wsllauncher

import (
	"errors"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestShouldForward_OffWindowsIsFalse(t *testing.T) {
	// The test host is linux/macOS (and the Linux build runs inside WSL), so the
	// CLI must run normally, never forward.
	if runtime.GOOS == "windows" {
		t.Skip("host is Windows")
	}
	if ShouldForward() {
		t.Fatal("ShouldForward must be false off Windows")
	}
}

func TestWSLArgvWith(t *testing.T) {
	// With an explicit distro selector.
	got := wslArgvWith([]string{"-d", "Ubuntu-24.04"}, "openframe", "bootstrap", "--deployment-mode=oss-tenant")
	want := []string{"-d", "Ubuntu-24.04", "--", "openframe", "bootstrap", "--deployment-mode=oss-tenant"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv = %v, want %v", got, want)
	}

	// Default distro (nil selector) → no -d, just `-- <cmd>`.
	if got := wslArgvWith(nil, "openframe"); !reflect.DeepEqual(got, []string{"--", "openframe"}) {
		t.Fatalf("default-distro argv = %v", got)
	}
}

func TestWSLDistroArgs(t *testing.T) {
	t.Setenv(distroEnv, "")
	if got := wslDistroArgs(); got != nil {
		t.Errorf("unset OPENFRAME_WSL_DISTRO must yield nil (default distro), got %v", got)
	}
	t.Setenv(distroEnv, "Ubuntu-22.04")
	if got := wslDistroArgs(); !reflect.DeepEqual(got, []string{"-d", "Ubuntu-22.04"}) {
		t.Errorf("set OPENFRAME_WSL_DISTRO must select it, got %v", got)
	}
}

func TestMergeWSLENV(t *testing.T) {
	// Preserves existing entries (with /flags suffix) and appends new names,
	// de-duplicating by variable name.
	got := mergeWSLENV("PATH/l:GITHUB_TOKEN", []string{"GITHUB_TOKEN", "OPENFRAME_GITHUB_TOKEN"})
	parts := strings.Split(got, ":")
	if len(parts) != 3 {
		t.Fatalf("expected 3 entries, got %v", parts)
	}
	if parts[0] != "PATH/l" || parts[1] != "GITHUB_TOKEN" || parts[2] != "OPENFRAME_GITHUB_TOKEN" {
		t.Fatalf("merge = %v", parts)
	}

	// Empty existing.
	if got := mergeWSLENV("", []string{"GITHUB_TOKEN"}); got != "GITHUB_TOKEN" {
		t.Fatalf("empty-existing merge = %q", got)
	}
}

func TestWithWSLEnv_SharesOnlySetVars(t *testing.T) {
	set := map[string]string{"GITHUB_TOKEN": "ghp_x"} // OPENFRAME_GITHUB_TOKEN not set
	lookup := func(k string) (string, bool) { v, ok := set[k]; return v, ok }

	out := withWSLEnv([]string{"HOME=/h", "WSLENV=OLD"}, lookup)

	var wslenv string
	for _, kv := range out {
		if strings.HasPrefix(kv, "WSLENV=") {
			wslenv = strings.TrimPrefix(kv, "WSLENV=")
		}
	}
	if !strings.Contains(wslenv, "GITHUB_TOKEN") {
		t.Errorf("set var must be shared: %q", wslenv)
	}
	if strings.Contains(wslenv, "OPENFRAME_GITHUB_TOKEN") {
		t.Errorf("unset var must not be shared: %q", wslenv)
	}
	if !strings.Contains(wslenv, "OLD") {
		t.Errorf("existing WSLENV entries must be preserved: %q", wslenv)
	}
}

func TestWithWSLEnv_NoForwardedVarsLeavesEnvUntouched(t *testing.T) {
	lookup := func(string) (string, bool) { return "", false }
	in := []string{"HOME=/h"}
	out := withWSLEnv(in, lookup)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("env should be untouched when nothing to forward: %v", out)
	}
}

func TestExitCodeOf(t *testing.T) {
	if c := exitCodeOf(nil); c != 0 {
		t.Errorf("nil err → 0, got %d", c)
	}
	if c := exitCodeOf(errors.New("boom")); c != 1 {
		t.Errorf("generic err → 1, got %d", c)
	}
	// A real *exec.ExitError from a command that exits non-zero.
	err := exec.Command("sh", "-c", "exit 7").Run()
	if runtime.GOOS != "windows" {
		if c := exitCodeOf(err); c != 7 {
			t.Errorf("ExitError → 7, got %d", c)
		}
	}
}

func TestNotInstalledError_HasGuidance(t *testing.T) {
	msg := notInstalledError().Error()
	for _, want := range []string{"not installed inside WSL", disableEnv, localBinaryEnv, distroEnv} {
		if !strings.Contains(msg, want) {
			t.Errorf("guidance missing %q:\n%s", want, msg)
		}
	}
}

// TestWSLUnavailableError categorizes wsl failures into actionable guidance:
// missing wsl.exe vs missing/unknown distro vs "not an availability problem".
func TestWSLUnavailableError(t *testing.T) {
	// nil → nil.
	if wslUnavailableError(nil) != nil {
		t.Error("nil error must categorize to nil")
	}

	// wsl.exe missing → "WSL is not installed" guidance.
	notFound := &exec.Error{Name: "wsl", Err: exec.ErrNotFound}
	if got := wslUnavailableError(notFound); got == nil || !strings.Contains(got.Error(), "wsl --install") {
		t.Errorf("missing wsl.exe must guide to `wsl --install`, got: %v", got)
	}

	// distro-not-found (stderr) → distro guidance mentioning OPENFRAME_WSL_DISTRO.
	for _, msg := range []string{
		"There is no distribution with the supplied name.",
		"WSL_E_DISTRO_NOT_FOUND",
		"Windows Subsystem for Linux has no installed distributions.",
	} {
		ee := &exec.ExitError{Stderr: []byte(msg)}
		got := wslUnavailableError(ee)
		if got == nil || !strings.Contains(got.Error(), distroEnv) {
			t.Errorf("distro error %q must guide to %s, got: %v", msg, distroEnv, got)
		}
	}

	// A generic non-availability failure → nil (caller falls back to not-installed).
	if got := wslUnavailableError(errors.New("some other failure")); got != nil {
		t.Errorf("non-availability error must categorize to nil, got: %v", got)
	}
}

// TestWSLBinaryLookupScript locks the resolver: it must consult PATH first and
// fall back to the ~/.openframe/bin install dir (which is not on PATH), so a
// binary installed there is still found.
func TestWSLBinaryLookupScript(t *testing.T) {
	for _, want := range []string{
		"command -v " + BinaryInWSL,
		`"$HOME/.openframe/bin/` + BinaryInWSL + `"`,
		"||", // PATH first, install-dir fallback
	} {
		if !strings.Contains(wslBinaryLookupScript, want) {
			t.Errorf("lookup script missing %q:\n%s", want, wslBinaryLookupScript)
		}
	}
}
