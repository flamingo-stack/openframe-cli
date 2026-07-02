package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestCurrentMatchesRuntime(t *testing.T) {
	if got, want := Current(), OS(runtime.GOOS); got != want {
		t.Fatalf("Current() = %q, want %q", got, want)
	}
}

func TestOSHelpersMutuallyExclusive(t *testing.T) {
	n := 0
	for _, b := range []bool{IsWindows(), IsMac(), IsLinux()} {
		if b {
			n++
		}
	}
	// On the three supported OSes exactly one holds; on any other, none do.
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		if n != 1 {
			t.Fatalf("expected exactly one OS helper true on %s, got %d", runtime.GOOS, n)
		}
	default:
		if n != 0 {
			t.Fatalf("expected no OS helper true on %s, got %d", runtime.GOOS, n)
		}
	}
}

func TestInstallDocsHintFor(t *testing.T) {
	docs := InstallDocs{
		Darwin:  "mac hint",
		Linux:   "linux hint",
		Windows: "windows hint",
		Default: "default hint",
	}
	cases := map[OS]string{
		Darwin:      "mac hint",
		Linux:       "linux hint",
		Windows:     "windows hint",
		OS("plan9"): "default hint", // unknown OS → Default
	}
	for os, want := range cases {
		if got := docs.hintFor(os); got != want {
			t.Errorf("hintFor(%q) = %q, want %q", os, got, want)
		}
	}
}

func TestInstallDocsHintFor_FallsBackWhenOSEmpty(t *testing.T) {
	// A known OS with an empty entry must fall back to Default, not return "".
	docs := InstallDocs{Default: "fallback"}
	for _, os := range []OS{Darwin, Linux, Windows} {
		if got := docs.hintFor(os); got != "fallback" {
			t.Errorf("hintFor(%q) with empty entry = %q, want %q", os, got, "fallback")
		}
	}
}

func TestInstallHint_KnownTools(t *testing.T) {
	for _, tool := range []string{"docker", "kubectl", "k3d", "helm"} {
		hint := InstallHint(tool)
		if hint == "" {
			t.Errorf("InstallHint(%q) is empty", tool)
		}
		if !strings.Contains(strings.ToLower(hint), tool) {
			t.Errorf("InstallHint(%q) = %q, expected it to mention the tool", tool, hint)
		}
		if !strings.Contains(hint, "http") {
			t.Errorf("InstallHint(%q) = %q, expected a documentation URL", tool, hint)
		}
	}
}

func TestInstallHint_EveryToolHasGuidanceForEveryOS(t *testing.T) {
	// Reqs 21/30: no OS (including Windows) may be left without guidance.
	for tool, docs := range toolDocs {
		for _, os := range []OS{Darwin, Linux, Windows, OS("solaris")} {
			if h := docs.hintFor(os); h == "" {
				t.Errorf("tool %q has no install guidance for %q", tool, os)
			}
		}
	}
}

func TestInstallHint_UnknownTool(t *testing.T) {
	hint := InstallHint("nonesuch")
	if !strings.Contains(hint, "nonesuch") {
		t.Fatalf("InstallHint(unknown) = %q, expected it to name the tool", hint)
	}
}
