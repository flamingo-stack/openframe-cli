package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
)

// TestCommandError_IsRedacted is the B5 guard: CommandError reaches user-facing
// output through the error handler even in non-verbose mode, so a registered
// secret in argv must never survive into it (the verbose prints were redacted,
// but the returned error carried the raw command line).
func TestCommandError_IsRedacted(t *testing.T) {
	const secret = "super-secret-token-12345"
	redact.RegisterSecret(secret)
	t.Cleanup(redact.ClearSecrets)

	exec := NewRealCommandExecutor(false, false)
	// A guaranteed-to-fail command carrying the secret in argv.
	_, err := exec.Execute(context.Background(), "openframe-no-such-binary", "--token", secret)
	if err == nil {
		t.Fatal("expected the command to fail")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error output leaks the registered secret: %v", err)
	}
	if !strings.Contains(err.Error(), "***") {
		t.Errorf("expected the redaction marker in place of the secret, got: %v", err)
	}
}

// TestCommandError_RedactsURLCredentials: URL-embedded credentials are scrubbed
// structurally even when never registered.
func TestCommandError_RedactsURLCredentials(t *testing.T) {
	exec := NewRealCommandExecutor(false, false)
	_, err := exec.Execute(context.Background(), "openframe-no-such-binary",
		"clone", "https://x-access-token:ghp_abcdef123456@github.com/org/repo.git")
	if err == nil {
		t.Fatal("expected the command to fail")
	}
	if strings.Contains(err.Error(), "ghp_abcdef123456") {
		t.Fatalf("error output leaks a URL-embedded credential: %v", err)
	}
}
