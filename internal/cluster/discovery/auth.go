package discovery

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// AuthFlow is the single gcloud login flow every GKE-touching command goes
// through: it checks the auth state and, in an interactive session, offers to
// run `gcloud auth login` (and `gcloud auth application-default login` when
// terraform needs ADC) right there — so the user never has to leave the CLI.
// Non-interactive sessions never prompt (CI rule) and get an actionable error
// instead.
type AuthFlow struct {
	exec executor.CommandExecutor
	// interactive reports whether prompting is allowed; seam for tests.
	interactive func() bool
	// confirm asks the user; seam for tests.
	confirm func(message string) (bool, error)
	// runLogin runs a gcloud login command attached to the terminal (browser
	// flows print URLs and read stdin, so it must bypass the capturing
	// executor); seam for tests.
	runLogin func(ctx context.Context, args ...string) error
}

// NewAuthFlow builds the production flow on the given executor (used for the
// non-interactive status probes, so they stay mockable).
func NewAuthFlow(exec executor.CommandExecutor) *AuthFlow {
	return &AuthFlow{
		exec:        exec,
		interactive: func() bool { return !sharedUI.IsNonInteractive() },
		confirm: func(message string) (bool, error) {
			return sharedUI.ConfirmActionInteractive(message, true)
		},
		runLogin: func(ctx context.Context, args ...string) error {
			cmd := osexec.CommandContext(ctx, "gcloud", args...) // #nosec G204 -- fixed argv assembled by this package, no user input
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

// NewAuthFlowWithSeams is the test constructor.
func NewAuthFlowWithSeams(exec executor.CommandExecutor, interactive func() bool,
	confirm func(string) (bool, error), runLogin func(context.Context, ...string) error) *AuthFlow {
	return &AuthFlow{exec: exec, interactive: interactive, confirm: confirm, runLogin: runLogin}
}

// Ensure makes sure gcloud is authenticated; with requireADC it also ensures
// Application Default Credentials, which terraform's google provider uses —
// `cluster create --type gke` needs both, discovery/use only the first.
func (f *AuthFlow) Ensure(ctx context.Context, requireADC bool) error {
	d := &GKEDiscoverer{exec: f.exec}
	switch d.AuthStatus(ctx) {
	case CLIMissing:
		return fmt.Errorf("gcloud is not installed — install the Google Cloud SDK first (https://cloud.google.com/sdk/docs/install)")
	case NotAuthenticated:
		if err := f.login(ctx, "Google Cloud login required. Log in now (opens a browser)?",
			[]string{"auth", "login"},
			func() bool { return d.AuthStatus(ctx) == Authenticated },
			"gcloud is not authenticated — run 'gcloud auth login'"); err != nil {
			return err
		}
	}

	if !requireADC {
		return nil
	}
	if f.hasADC(ctx) {
		return nil
	}
	return f.login(ctx, "Terraform needs Application Default Credentials. Set them up now (opens a browser)?",
		[]string{"auth", "application-default", "login"},
		func() bool { return f.hasADC(ctx) },
		"Application Default Credentials are missing — run 'gcloud auth application-default login'")
}

// hasADC probes Application Default Credentials without prompting.
func (f *AuthFlow) hasADC(ctx context.Context) bool {
	result, err := f.exec.Execute(ctx, "gcloud", "auth", "application-default", "print-access-token")
	return err == nil && result != nil && strings.TrimSpace(result.Stdout) != ""
}

// login runs one interactive login step: prompt → run → re-verify.
func (f *AuthFlow) login(ctx context.Context, prompt string, args []string, verified func() bool, manualHint string) error {
	if !f.interactive() {
		return fmt.Errorf("%s", manualHint)
	}
	confirmed, err := f.confirm(prompt)
	if err != nil {
		return err
	}
	if !confirmed {
		return fmt.Errorf("%s", manualHint)
	}
	if err := f.runLogin(ctx, args...); err != nil {
		return fmt.Errorf("gcloud %s failed: %w", strings.Join(args, " "), err)
	}
	if !verified() {
		return fmt.Errorf("login did not complete — %s", manualHint)
	}
	pterm.Success.Println("Google Cloud authentication complete")
	return nil
}
