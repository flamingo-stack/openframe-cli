package k3d

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// B3 contract guards for the inotify bump: it must never reach an interactive
// sudo password prompt — skip on macOS (no fs.inotify.* keys), skip when the
// limits already suffice, and escalate only with `sudo -n` (fail, don't prompt).

func TestInotify_DarwinIsSkippedEntirely(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	m := NewK3dManager(mock, false)

	require.NoError(t, m.increaseInotifyLimitsFor(context.Background(), "darwin"))
	assert.Zero(t, mock.GetCommandCount(), "macOS has no inotify sysctls; nothing may run (the old code ran `sudo sysctl` and prompted for a password)")
}

func TestInotify_SufficientLimitsSkipSudo(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sysctl -n", &executor.CommandResult{ExitCode: 0, Stdout: "999999\n", Duration: time.Millisecond})
	m := NewK3dManager(mock, false)

	require.NoError(t, m.increaseInotifyLimitsFor(context.Background(), "linux"))
	for _, rc := range mock.Commands() {
		assert.NotEqualf(t, "sudo", rc.Name, "no privilege escalation when limits already suffice: %v", rc)
	}
}

func TestInotify_LowLimitsEscalateWithSudoN(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sysctl -n", &executor.CommandResult{ExitCode: 0, Stdout: "8192\n", Duration: time.Millisecond})
	m := NewK3dManager(mock, false)

	require.NoError(t, m.increaseInotifyLimitsFor(context.Background(), "linux"))

	var sawSudo bool
	for _, rc := range mock.Commands() {
		if rc.Name != "sudo" {
			continue
		}
		sawSudo = true
		require.NotEmpty(t, rc.Args)
		assert.Equalf(t, "-n", rc.Args[0], "sudo must run non-interactively (-n) so it can never prompt for a password: %v", rc.Args)
	}
	assert.True(t, sawSudo, "low limits must trigger the sysctl write")
}

func TestInotify_SudoFailureIsActionableNotAPrompt(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sysctl -n", &executor.CommandResult{ExitCode: 0, Stdout: "8192\n", Duration: time.Millisecond})
	mock.SetResponse("sudo -n sysctl", &executor.CommandResult{ExitCode: 1, Stderr: "sudo: a password is required", Duration: time.Millisecond})
	m := NewK3dManager(mock, false)

	err := m.increaseInotifyLimitsFor(context.Background(), "linux")
	require.Error(t, err, "missing passwordless sudo surfaces as an error (downgraded to a warning by the caller)")
	assert.Contains(t, err.Error(), "sudo sysctl -w", "error must carry the manual command since we refused to prompt")
}

func TestInotify_WindowsWSLUsesSudoN(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	m := NewK3dManager(mock, false)

	require.NoError(t, m.increaseInotifyLimitsFor(context.Background(), "windows"))
	cmds := mock.Commands()
	require.Len(t, cmds, 1)
	assert.Equal(t, "wsl", cmds[0].Name)
	assert.Truef(t, strings.Contains(cmds[0].String(), "sudo -n sysctl"), "WSL branch must also be prompt-free: %s", cmds[0])
}
