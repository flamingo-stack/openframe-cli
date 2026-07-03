// Package wsllauncher makes the native Windows build of OpenFrame re-run itself
// inside WSL (Option 1). On Windows the cluster (Docker + k3d) lives in WSL2,
// and the native Kubernetes client cannot reach it from a native Windows
// process — so instead of shelling individual tools into WSL, the whole CLI is
// forwarded to a Linux build running inside WSL, where client-go/helm/k3d all
// work natively.
//
// The Linux build (running inside WSL) has runtime.GOOS == "linux" and therefore
// never forwards, so there is no recursion.
package wsllauncher

import (
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// Distro is the WSL distribution OpenFrame runs inside.
	Distro = "Ubuntu"
	// BinaryInWSL is the OpenFrame executable name expected on the PATH in WSL.
	BinaryInWSL = "openframe"
	// disableEnv, when set, bypasses forwarding and runs natively on Windows
	// (unsupported; provided as a debugging escape hatch).
	disableEnv = "OPENFRAME_NO_WSL_FORWARD"
)

// forwardedEnvVars are host (Windows) environment variables shared into WSL via
// WSLENV so credentials/config reach the Linux process.
var forwardedEnvVars = []string{
	"GITHUB_TOKEN",
	"OPENFRAME_GITHUB_TOKEN",
}

// ShouldForward reports whether this process must re-run itself inside WSL: only
// the native Windows build forwards, and only when not explicitly disabled.
func ShouldForward() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	return os.Getenv(disableEnv) == ""
}

// Forward re-runs `openframe <args>` inside WSL, passing through stdio, the
// forwarded environment, and the child's exit code. It returns an error only
// when WSL / the WSL openframe binary are unavailable (with setup guidance).
func Forward(args []string) (int, error) {
	if err := verifyOpenframeInWSL(); err != nil {
		return 1, err
	}
	cmd := exec.Command("wsl", buildForwardArgv(Distro, BinaryInWSL, args)...) // #nosec G204 -- fixed distro/binary; user args are the CLI's own args
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = withWSLEnv(os.Environ(), os.LookupEnv)
	// The child's own output already surfaced any failure; propagate its code.
	return exitCodeOf(cmd.Run()), nil
}

// buildForwardArgv builds the argv for `wsl -d <distro> -- <binary> <args...>`.
// The `--` guarantees the remaining tokens are treated as the command line, not
// wsl flags.
func buildForwardArgv(distro, binary string, args []string) []string {
	out := []string{"-d", distro, "--", binary}
	return append(out, args...)
}

// withWSLEnv returns env with WSLENV extended so the forwarded vars that are
// actually set on the host are shared into WSL.
func withWSLEnv(env []string, lookup func(string) (string, bool)) []string {
	var share []string
	for _, v := range forwardedEnvVars {
		if _, ok := lookup(v); ok {
			share = append(share, v)
		}
	}
	if len(share) == 0 {
		return env
	}

	// Extract any existing WSLENV from env itself (kept consistent with what we
	// filter out) and merge our forwarded vars into it.
	var existing string
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if strings.HasPrefix(kv, "WSLENV=") {
			existing = strings.TrimPrefix(kv, "WSLENV=")
			continue
		}
		out = append(out, kv)
	}
	return append(out, "WSLENV="+mergeWSLENV(existing, share))
}

// mergeWSLENV merges add into an existing WSLENV value, de-duplicating by
// variable name (WSLENV entries may carry a "/flags" suffix, e.g. "PATH/l").
func mergeWSLENV(existing string, add []string) string {
	seen := map[string]bool{}
	var parts []string
	for _, p := range strings.Split(existing, ":") {
		if p == "" {
			continue
		}
		name := strings.SplitN(p, "/", 2)[0]
		if !seen[name] {
			seen[name] = true
			parts = append(parts, p)
		}
	}
	for _, v := range add {
		if !seen[v] {
			seen[v] = true
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, ":")
}

// exitCodeOf maps a Cmd.Run() error to a process exit code.
func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if stderrors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}

// verifyOpenframeInWSL checks that the openframe binary is runnable inside WSL,
// returning setup guidance if not.
func verifyOpenframeInWSL() error {
	// `command -v` is a shell builtin, so run it via bash inside the distro.
	check := exec.Command("wsl", "-d", Distro, "--", "bash", "-lc", "command -v "+BinaryInWSL) // #nosec G204 -- fixed distro/binary name
	if err := check.Run(); err != nil {
		return notInstalledError()
	}
	return nil
}

func notInstalledError() error {
	return fmt.Errorf(`OpenFrame is not installed inside WSL (%s)

On Windows the cluster runs in WSL2 and OpenFrame must run there too. Install
the Linux build inside your WSL distro and put it on your PATH, then re-run:

    wsl -d %s
    # inside WSL: download the openframe linux binary and place it on your PATH

Set %s=1 to bypass and run natively on Windows (unsupported)`,
		Distro, Distro, disableEnv)
}
