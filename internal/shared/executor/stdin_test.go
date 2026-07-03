package executor

import (
	"context"
	"runtime"
	"testing"
)

// TestExecuteWithOptions_StdinIsPiped proves the executor wires ExecuteOptions.Stdin
// to the process stdin end-to-end, using `cat` as an echo. This is what lets
// helm read values from `-f -` without a temp file.
func TestExecuteWithOptions_StdinIsPiped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("cat is not available on Windows shells")
	}
	exec := NewRealCommandExecutor(false, false)

	const payload = "fullnameOverride: argocd\n"
	res, err := exec.ExecuteWithOptions(context.Background(), ExecuteOptions{
		Command: "cat",
		Stdin:   []byte(payload),
	})
	if err != nil {
		t.Fatalf("cat with stdin: %v", err)
	}
	if res.Stdout != payload {
		t.Fatalf("stdout = %q, want %q (stdin was not piped through)", res.Stdout, payload)
	}
}

func TestExecuteWithOptions_NoStdinIsHarmless(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a unix echo")
	}
	exec := NewRealCommandExecutor(false, false)
	// No Stdin set → the command still runs normally.
	res, err := exec.ExecuteWithOptions(context.Background(), ExecuteOptions{
		Command: "echo",
		Args:    []string{"ok"},
	})
	if err != nil || res.Stdout != "ok\n" {
		t.Fatalf("echo without stdin: stdout=%q err=%v", res.Stdout, err)
	}
}
