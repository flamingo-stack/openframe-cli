package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnnotationEscaping(t *testing.T) {
	if got := escapeAnnotationData("a%b\nc\rd"); got != "a%25b%0Ac%0Dd" {
		t.Fatalf("data escape = %q", got)
	}
	if got := escapeAnnotationProperty("t: a,b"); got != "t%3A a%2Cb" {
		t.Fatalf("property escape = %q", got)
	}
}

func TestAppendStepSummary_WritesOnlyInActions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")

	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITHUB_STEP_SUMMARY", path)
	AppendStepSummary("### nope")
	if _, err := os.Stat(path); err == nil {
		t.Fatal("outside Actions nothing must be written")
	}

	t.Setenv("GITHUB_ACTIONS", "true")
	AppendStepSummary("### hello")
	AppendStepSummary("more")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "### hello") || !strings.Contains(string(data), "more") {
		t.Fatalf("summary = %q", data)
	}
}

func TestInGitHubActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")
	if InGitHubActions() {
		t.Fatal("empty env → not in Actions")
	}
	t.Setenv("GITHUB_ACTIONS", "true")
	if !InGitHubActions() {
		t.Fatal("GITHUB_ACTIONS=true → in Actions")
	}
}
