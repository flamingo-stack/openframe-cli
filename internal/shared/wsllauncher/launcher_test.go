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

func TestBuildForwardArgv(t *testing.T) {
	got := buildForwardArgv("Ubuntu", "openframe", []string{"bootstrap", "--deployment-mode=oss-tenant"})
	want := []string{"-d", "Ubuntu", "--", "openframe", "bootstrap", "--deployment-mode=oss-tenant"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv = %v, want %v", got, want)
	}

	// No user args → still a valid invocation.
	if got := buildForwardArgv("Ubuntu", "openframe", nil); !reflect.DeepEqual(got, []string{"-d", "Ubuntu", "--", "openframe"}) {
		t.Fatalf("empty-args argv = %v", got)
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
	for _, want := range []string{"not installed inside WSL", "wsl -d Ubuntu", disableEnv} {
		if !strings.Contains(msg, want) {
			t.Errorf("guidance missing %q:\n%s", want, msg)
		}
	}
}
