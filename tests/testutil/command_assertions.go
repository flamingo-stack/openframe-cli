package testutil

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
)

// Executor-seam assertions (testing plan §4).
//
// These operate on the structured command log of a *executor.MockCommandExecutor
// (RecordedCommand{Name, Args, Env}). They verify *how* the CLI constructed a
// command — which is exactly where the audit's security bugs live — without
// running any external binary. The mock keeps argv as discrete elements, so
// these can tell "passed $(x) as one literal arg" apart from "built a shell
// string", a distinction a flattened log cannot make.

// shellMetaPattern matches argv elements that would be dangerous if a shell
// ever interpreted them. Used to detect shell-string construction.
var shellMetachars = []string{"$(", "`", "&&", "||", ";", "|", ">", "<", "\n"}

// AssertNoArgContains fails if the given secret/substring appears in ANY argv
// element of ANY recorded command. Use to prove a token is never placed on the
// command line (audit I1) — credentials must travel via env/stdin instead.
func AssertNoArgContains(t *testing.T, cmds []executor.RecordedCommand, secret string) {
	t.Helper()
	if secret == "" {
		return
	}
	for _, c := range cmds {
		if strings.Contains(c.Name, secret) {
			assert.Failf(t, "secret leaked into command name",
				"command %q contains secret %q", c.String(), secret)
		}
		for i, a := range c.Args {
			if strings.Contains(a, secret) {
				assert.Failf(t, "secret leaked into argv",
					"command %q arg[%d]=%q contains secret %q", c.String(), i, a, secret)
			}
		}
	}
}

// AssertNoBashCConcatenation fails if any recorded command shells out via
// `bash -c`/`sh -c` with a payload that contains un-neutralized shell
// metacharacters. This is the regression guard for audit I3 (the WSL helm/
// kubectl path that concatenated escaped args into a bash string).
//
// It deliberately allows a `bash -c` whose script is a fixed literal with no
// metacharacters smuggled in from arguments — the failure mode we care about is
// argument-derived metacharacters reaching the shell.
func AssertNoBashCConcatenation(t *testing.T, cmds []executor.RecordedCommand, injected string) {
	t.Helper()
	for _, c := range cmds {
		script, ok := shellCScript(c)
		if !ok {
			continue
		}
		if injected != "" && strings.Contains(script, injected) {
			assert.Failf(t, "injected argument reached a shell -c string",
				"command %q embedded %q into its shell script: %q", c.String(), injected, script)
		}
	}
}

// AssertArgIsDiscrete fails unless `want` appears as a complete, standalone argv
// element (not merely as a substring of a larger concatenated string) in some
// command. Proves an argument was passed verbatim rather than spliced into a
// shell line (audit I3).
func AssertArgIsDiscrete(t *testing.T, cmds []executor.RecordedCommand, want string) {
	t.Helper()
	for _, c := range cmds {
		for _, a := range c.Args {
			if a == want {
				return
			}
		}
	}
	assert.Failf(t, "argument was not passed as a discrete argv element",
		"no recorded command had an argv element exactly equal to %q; commands=%v", want, render(cmds))
}

// AssertNoDeleteOf fails if any recorded command appears to delete a protected
// resource (e.g. the `kube-system` namespace). Regression guard for audit I7.
func AssertNoDeleteOf(t *testing.T, cmds []executor.RecordedCommand, protected string) {
	t.Helper()
	for _, c := range cmds {
		joined := c.String()
		if strings.Contains(joined, "delete") && containsArg(c, protected) {
			assert.Failf(t, "destructive command targeted a protected resource",
				"command %q deletes protected resource %q", joined, protected)
		}
	}
}

// AssertNoCurlPipeShell fails if any command pipes a remote download straight
// into a shell (curl ... | bash). Regression guard for audit I5/M1.
func AssertNoCurlPipeShell(t *testing.T, cmds []executor.RecordedCommand) {
	t.Helper()
	for _, c := range cmds {
		script, ok := shellCScript(c)
		if !ok {
			continue
		}
		low := strings.ToLower(script)
		if (strings.Contains(low, "curl") || strings.Contains(low, "wget")) &&
			strings.Contains(low, "|") &&
			(strings.Contains(low, "bash") || strings.Contains(low, "sh")) {
			assert.Failf(t, "curl|bash pattern detected",
				"command %q pipes a download into a shell: %q", c.String(), script)
		}
	}
}

// --- helpers ---

// shellCScript returns the script body of a `bash -c <script>` / `sh -c <script>`
// invocation (possibly wrapped by `wsl -d ... -u ... bash -c <script>`).
func shellCScript(c executor.RecordedCommand) (string, bool) {
	args := c.Args
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-c" && (i > 0 && isShell(args[i-1]) || isShell(c.Name)) {
			return args[i+1], true
		}
	}
	// `bash -c <script>` where bash is the command name itself.
	if isShell(c.Name) {
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "-c" {
				return args[i+1], true
			}
		}
	}
	return "", false
}

func isShell(s string) bool {
	switch s {
	case "bash", "sh", "zsh", "/bin/bash", "/bin/sh":
		return true
	}
	return false
}

func containsArg(c executor.RecordedCommand, want string) bool {
	for _, a := range c.Args {
		if a == want {
			return true
		}
	}
	return false
}

func render(cmds []executor.RecordedCommand) []string {
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.String()
	}
	return out
}

// HasShellMetachars reports whether s contains characters a shell would treat
// specially. Exposed for tests that want to assert an arg is shell-safe.
func HasShellMetachars(s string) bool {
	for _, m := range shellMetachars {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}
