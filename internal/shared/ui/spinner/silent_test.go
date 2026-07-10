package spinner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what was
// written. The leak this guards against goes to the real stdout, so asserting on
// an injected writer would not see it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// TestSpinner_SilentSuppressesSuccessButNotFailure locks --silent's contract
// ("suppress all output except errors") for the spinner.
//
// The spinner prints its final line via pterm.Success.WithWriter(s.out), and
// that WithWriter overrides the io.Discard writer ui.SetSilent() installs on the
// package-level printers. Every spinner in the CLI therefore printed its Success
// line to stdout under --silent. A failure, by contrast, must still be reported.
func TestSpinner_SilentSuppressesSuccessButNotFailure(t *testing.T) {
	// SetSilent is process-global and irreversible by design; snapshot the
	// printers it rewires so the rest of the package's tests keep their output.
	info, success, warning, errp := pterm.Info, pterm.Success, pterm.Warning, pterm.Error
	t.Cleanup(func() { pterm.Info, pterm.Success, pterm.Warning, pterm.Error = info, success, warning, errp })

	ui.SetSilent()

	// New() snapshots os.Stdout, so it must be constructed inside the capture.
	out := captureStdout(t, func() {
		sp := New()
		sp.Start("installing")
		sp.Success("installed")
	})
	if strings.Contains(out, "installed") {
		t.Errorf("--silent must not print the spinner's success line to stdout; got %q", out)
	}

	// A failure must still surface: --silent suppresses everything EXCEPT errors.
	var errOut bytes.Buffer
	pterm.Error = *pterm.Error.WithWriter(&errOut)
	sp := New()
	sp.Start("installing")
	sp.Fail("install failed")

	if got := errOut.String(); !strings.Contains(got, "install failed") {
		t.Errorf("--silent must still report failures; got %q", got)
	}
}
