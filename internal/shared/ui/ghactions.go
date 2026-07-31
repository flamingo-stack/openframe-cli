package ui

import (
	"fmt"
	"os"
	"strings"
)

// GitHub Actions workflow-command helpers. All are no-ops outside an Actions
// job, so callers wire them unconditionally: stages fold into ::group::
// blocks, failures surface as job/PR annotations, and the closing summary
// lands in the job's Step Summary panel — the places a person actually looks
// when a 20-minute workflow run fails.

// InGitHubActions reports whether the process runs inside a GitHub Actions job.
func InGitHubActions() bool { return os.Getenv("GITHUB_ACTIONS") == "true" }

// GroupStart opens a foldable log group titled title.
func GroupStart(title string) {
	if InGitHubActions() {
		fmt.Printf("::group::%s\n", escapeAnnotationData(title))
	}
}

// GroupEnd closes the innermost log group.
func GroupEnd() {
	if InGitHubActions() {
		fmt.Println("::endgroup::")
	}
}

// ErrorAnnotation surfaces a failure as an ::error:: annotation (shown in the
// job summary and on the PR when applicable).
func ErrorAnnotation(title, message string) {
	if !InGitHubActions() {
		return
	}
	fmt.Printf("::error title=%s::%s\n", escapeAnnotationProperty(title), escapeAnnotationData(message))
}

// AppendStepSummary appends a markdown fragment to the job's Step Summary.
// Best-effort: a missing or unwritable summary file is silently skipped.
func AppendStepSummary(markdown string) {
	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if !InGitHubActions() || path == "" {
		return
	}
	// The runner pre-creates the summary file; 0600 only applies if we are
	// first, and nothing else needs to read it but the runner (root there).
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 G703 -- GITHUB_STEP_SUMMARY is the Actions runner's own contract: it names the file we are MEANT to append to, and we only ever append markdown
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(markdown + "\n")
}

// escapeAnnotationData escapes a workflow-command data value per the Actions
// runner's rules.
func escapeAnnotationData(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A")
	return r.Replace(s)
}

// escapeAnnotationProperty escapes a workflow-command property value (titles
// need the extra ':' and ',' escapes).
func escapeAnnotationProperty(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C")
	return r.Replace(s)
}
