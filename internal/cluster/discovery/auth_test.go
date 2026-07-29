package discovery

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// authFlowHarness wires an AuthFlow with recording seams.
type authFlowHarness struct {
	flow       *AuthFlow
	mock       *executor.MockCommandExecutor
	logins     []string
	confirms   []string
	confirmAns bool
}

func newAuthHarness(t *testing.T, interactive bool) *authFlowHarness {
	t.Helper()
	h := &authFlowHarness{mock: executor.NewMockCommandExecutor(), confirmAns: true}
	h.flow = NewAuthFlowWithSeams(h.mock,
		func() bool { return interactive },
		func(msg string) (bool, error) {
			h.confirms = append(h.confirms, msg)
			return h.confirmAns, nil
		},
		func(ctx context.Context, args ...string) error {
			joined := strings.Join(args, " ")
			h.logins = append(h.logins, joined)
			// A successful login makes ITS OWN subsequent status probe pass —
			// `auth login` must not grant ADC and vice versa.
			if strings.Contains(joined, "application-default") {
				h.mock.SetResponse("application-default print-access-token", &executor.CommandResult{ExitCode: 0, Stdout: "ya29.token\n"})
			} else {
				// A fresh `auth login` makes the account listed AND able to mint a
				// token (the freshness probe).
				h.mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
				h.mock.SetResponse("gcloud auth print-access-token", &executor.CommandResult{ExitCode: 0, Stdout: "ya29.token\n"})
			}
			return nil
		})
	return h
}

func (h *authFlowHarness) loggedOut() {
	h.mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: ""})
	h.mock.SetResponse("gcloud auth print-access-token", &executor.CommandResult{ExitCode: 1, Stderr: "reauth required"})
	h.mock.SetResponse("application-default print-access-token", &executor.CommandResult{ExitCode: 1, Stderr: "no credentials"})
}

func (h *authFlowHarness) loggedIn(withADC bool) {
	h.mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	h.mock.SetResponse("gcloud auth print-access-token", &executor.CommandResult{ExitCode: 0, Stdout: "ya29.token\n"})
	if withADC {
		h.mock.SetResponse("application-default print-access-token", &executor.CommandResult{ExitCode: 0, Stdout: "ya29.token\n"})
	} else {
		h.mock.SetResponse("application-default print-access-token", &executor.CommandResult{ExitCode: 1, Stderr: "no credentials"})
	}
}

// primaryExpired models the M6 case: the account is listed ACTIVE but its token
// is expired, so `gcloud auth list` still shows it while `print-access-token`
// fails.
func (h *authFlowHarness) primaryExpired() {
	h.mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	h.mock.SetResponse("gcloud auth print-access-token", &executor.CommandResult{ExitCode: 1, Stderr: "reauth required"})
	h.mock.SetResponse("application-default print-access-token", &executor.CommandResult{ExitCode: 0, Stdout: "ya29.token\n"})
}

func TestAuthFlow_AlreadyAuthenticatedIsSilent(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedIn(true)

	require.NoError(t, h.flow.Ensure(context.Background(), true))
	assert.Empty(t, h.logins, "no login must run when already authenticated")
	assert.Empty(t, h.confirms, "no prompt must be shown when already authenticated")
}

func TestAuthFlow_InteractiveLoginRuns(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedOut()

	require.NoError(t, h.flow.Ensure(context.Background(), false))
	assert.Equal(t, []string{"auth login"}, h.logins)
}

func TestAuthFlow_ADCRequiredRunsBothLogins(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedOut()

	require.NoError(t, h.flow.Ensure(context.Background(), true))
	assert.Equal(t, []string{"auth login", "auth application-default login"}, h.logins)
}

func TestAuthFlow_ADCOnlyWhenCLICredsPresent(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedIn(false)

	require.NoError(t, h.flow.Ensure(context.Background(), true))
	assert.Equal(t, []string{"auth application-default login"}, h.logins)
}

// M6: an expired primary login (listed ACTIVE but no mintable token) must be
// caught and re-logged-in BEFORE the ADC step, not surface later as a raw error.
func TestAuthFlow_ExpiredPrimaryLoginReLogsIn(t *testing.T) {
	h := newAuthHarness(t, true)
	h.primaryExpired()

	require.NoError(t, h.flow.Ensure(context.Background(), true))
	assert.Equal(t, []string{"auth login"}, h.logins,
		"expired primary login must trigger re-login, and before any ADC step")
}

func TestAuthFlow_ExpiredPrimaryNonInteractiveErrors(t *testing.T) {
	h := newAuthHarness(t, false)
	h.primaryExpired()

	err := h.flow.Ensure(context.Background(), true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired", "the error must name the expired login, with the gcloud auth login hint")
	assert.Contains(t, err.Error(), "gcloud auth login")
	assert.Empty(t, h.logins)
}

func TestAuthFlow_NonInteractiveNeverPrompts(t *testing.T) {
	h := newAuthHarness(t, false)
	h.loggedOut()

	err := h.flow.Ensure(context.Background(), true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gcloud auth login")
	assert.Empty(t, h.logins, "non-interactive sessions must never launch a login")
	assert.Empty(t, h.confirms)
}

func TestAuthFlow_DeclinedGivesManualHint(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedOut()
	h.confirmAns = false

	err := h.flow.Ensure(context.Background(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gcloud auth login")
	assert.Empty(t, h.logins)
}

func TestAuthFlow_FailedLoginSurfaces(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedOut()
	h.flow.runLogin = func(ctx context.Context, args ...string) error {
		return errors.New("browser exploded")
	}

	err := h.flow.Ensure(context.Background(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "browser exploded")
}

func TestAuthFlow_UnverifiedLoginFails(t *testing.T) {
	h := newAuthHarness(t, true)
	h.loggedOut()
	// Login "succeeds" but the status probe still reports logged-out.
	h.flow.runLogin = func(ctx context.Context, args ...string) error { return nil }

	err := h.flow.Ensure(context.Background(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not complete")
}
