package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/pterm/pterm"
)

// captureUI redirects every printer ShowCleanupSummary writes to and returns
// what was printed.
func captureUI(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	info, warning, success := pterm.Info, pterm.Warning, pterm.Success
	basic := pterm.DefaultBasicText
	t.Cleanup(func() {
		pterm.Info, pterm.Warning, pterm.Success = info, warning, success
		pterm.DefaultBasicText = basic
	})
	pterm.Info = *pterm.Info.WithWriter(&buf)
	pterm.Warning = *pterm.Warning.WithWriter(&buf)
	pterm.Success = *pterm.Success.WithWriter(&buf)
	pterm.DefaultBasicText = *pterm.DefaultBasicText.WithWriter(&buf)
	fn()
	return buf.String()
}

// TestShowCleanupSummary_ReportsRealCounts (M2.1): the summary must describe
// what happened. The old one printed "Removed unused Docker images / Freed up
// disk space / Optimized cluster performance" unconditionally — the same text
// whether cleanup removed twenty objects or none.
func TestShowCleanupSummary_ReportsRealCounts(t *testing.T) {
	ui := NewOperationsUI()
	out := captureUI(t, func() {
		ui.ShowCleanupSummary("dev", models.CleanupResult{
			ApplicationsDeleted: 3,
			ReleasesRemoved:     2,
			NamespacesDeleted:   1,
		})
	})

	for _, want := range []string{"3 ArgoCD application(s)", "2 Helm release(s)", "1 namespace(s)"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary must report %q; got:\n%s", want, out)
		}
	}
	// Counts that are zero are not printed as noise.
	if strings.Contains(out, "node(s) pruned") {
		t.Errorf("a zero count must not be listed; got:\n%s", out)
	}
	if strings.Contains(out, "Freed up disk space") {
		t.Errorf("the summary must not claim un-measured outcomes; got:\n%s", out)
	}
}

// TestShowCleanupSummary_EmptyClusterSaysSo: removing nothing must read as
// "nothing to remove", not as a list of accomplishments.
func TestShowCleanupSummary_EmptyClusterSaysSo(t *testing.T) {
	out := captureUI(t, func() { NewOperationsUI().ShowCleanupSummary("dev", models.CleanupResult{}) })

	if !strings.Contains(out, "Nothing to remove") {
		t.Errorf("an empty cleanup must say so; got:\n%s", out)
	}
}

// TestShowCleanupSummary_PartialFailureIsVisible: cleanup swallows phase errors
// by design so a broken cluster can still be torn down. That is only safe if
// the user is told which phases failed — otherwise "cleanup completed" is a lie
// and the leftover resources are a surprise.
func TestShowCleanupSummary_PartialFailureIsVisible(t *testing.T) {
	result := models.CleanupResult{ReleasesRemoved: 1}
	result.AddFailure("Kubernetes namespaces", errors.New("connection refused"))

	out := captureUI(t, func() { NewOperationsUI().ShowCleanupSummary("dev", result) })

	if strings.Contains(out, "cleanup completed") {
		t.Errorf("a partial cleanup must not be reported as completed; got:\n%s", out)
	}
	for _, want := range []string{"finished with problems", "Kubernetes namespaces", "connection refused", "some resources may remain"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary must surface %q; got:\n%s", want, out)
		}
	}
}
