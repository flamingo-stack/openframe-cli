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
	// distroEnv overrides which WSL distribution OpenFrame uses. When unset, the
	// WSL *default* distribution is used (no `-d` flag). Hardcoding a name like
	// "Ubuntu" breaks on hosts whose distro is registered as "Ubuntu-24.04",
	// "Ubuntu-22.04", etc. (WSL_E_DISTRO_NOT_FOUND).
	distroEnv = "OPENFRAME_WSL_DISTRO"
	// BinaryInWSL is the OpenFrame executable name expected on the PATH in WSL.
	BinaryInWSL = "openframe"
	// disableEnv, when set, bypasses forwarding and runs natively on Windows
	// (unsupported; provided as a debugging escape hatch).
	disableEnv = "OPENFRAME_NO_WSL_FORWARD"
)

// wslDistroArgs returns the `-d <distro>` selector when OPENFRAME_WSL_DISTRO is
// set, else nil so the WSL default distribution is targeted.
func wslDistroArgs() []string {
	if d := strings.TrimSpace(os.Getenv(distroEnv)); d != "" {
		return []string{"-d", d}
	}
	return nil
}

// wslArgvWith builds the argv for `wsl <distroArgs> -- <cmd...>`. Pure/testable.
func wslArgvWith(distroArgs []string, cmd ...string) []string {
	out := append([]string{}, distroArgs...)
	out = append(out, "--")
	return append(out, cmd...)
}

// wslArgv builds the wsl argv for the configured distribution (the WSL default
// unless OPENFRAME_WSL_DISTRO is set).
func wslArgv(cmd ...string) []string {
	return wslArgvWith(wslDistroArgs(), cmd...)
}

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
// forwarded environment, and the child's exit code. It auto-installs the
// matching Linux release into WSL when missing, and returns an error only when
// WSL / the WSL openframe binary are unavailable (with setup guidance).
func Forward(version string, args []string) (int, error) {
	if err := ensureOpenframeInWSL(version, runtime.GOARCH); err != nil {
		return 1, err
	}
	// Resolve the concrete binary path: a PATH-installed openframe, else the
	// absolute install dir (~/.openframe/bin is not necessarily on PATH). ensure
	// above already verified this resolves to a runnable binary.
	bin, err := resolveWSLBinaryPath()
	if err != nil || bin == "" {
		return 1, notInstalledError()
	}
	forwardArgv := append([]string{bin}, args...)
	cmd := exec.Command("wsl", wslArgv(forwardArgv...)...) // #nosec G204 -- bin is resolved from a fixed lookup; user args are the CLI's own args
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = withWSLEnv(os.Environ(), os.LookupEnv)
	// The child's own output already surfaced any failure; propagate its code.
	return exitCodeOf(cmd.Run()), nil
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

// wslBinaryLookupScript resolves the openframe binary path inside WSL: the
// PATH-resolved binary if present, otherwise the absolute path where the
// launcher installs it. `command -v` fails (prints nothing) when openframe is
// not on PATH, so the `||` branch falls back to ~/.openframe/bin — which is
// where both the release download and OPENFRAME_WSL_BINARY install it, and which
// is not necessarily on the WSL PATH. It is a constant (no interpolated input).
const wslBinaryLookupScript = "command -v " + BinaryInWSL + ` 2>/dev/null || printf '%s' "$HOME/.openframe/bin/` + BinaryInWSL + `"`

// resolveWSLBinaryPath returns the absolute path of the openframe binary inside
// WSL (PATH-resolved or the install dir). The path is not guaranteed to exist —
// callers verify it is runnable.
func resolveWSLBinaryPath() (string, error) {
	out, err := exec.Command("wsl", wslArgv("bash", "-lc", wslBinaryLookupScript)...).Output() // #nosec G204 -- script is a constant
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// wslBinaryStatus reports whether openframe is runnable inside WSL. A non-nil
// error means WSL ITSELF is unavailable (not installed, or no distro registered)
// and carries actionable setup guidance. err==nil with present==false means WSL
// works but the binary is simply not installed yet (installable) — the caller
// may auto-install and then re-check.
func wslBinaryStatus() (present bool, err error) {
	bin, rerr := resolveWSLBinaryPath()
	if rerr != nil {
		// Distinguish "WSL/​distro not available" (fatal, actionable) from an
		// otherwise-unknown failure (treat as "binary absent" so install can try).
		if guide := wslUnavailableError(rerr); guide != nil {
			return false, guide
		}
		return false, nil
	}
	if bin == "" {
		return false, nil
	}
	check := exec.Command("wsl", wslArgv("bash", "-lc", "test -x "+shellSingleQuote(bin))...) // #nosec G204 -- bin single-quoted
	if check.Run() != nil {
		return false, nil
	}
	return true, nil
}

// verifyOpenframeInWSL checks that the openframe binary is runnable inside WSL,
// returning setup guidance if not.
func verifyOpenframeInWSL() error {
	present, err := wslBinaryStatus()
	if err != nil {
		return err
	}
	if !present {
		return notInstalledError()
	}
	return nil
}

// wslStderr extracts the stderr captured by exec (Output() stores it on
// *exec.ExitError), falling back to the error text.
func wslStderr(err error) string {
	var ee *exec.ExitError
	if stderrors.As(err, &ee) {
		return string(ee.Stderr)
	}
	return err.Error()
}

// wslUnavailableError categorizes a failed `wsl` invocation into actionable
// guidance — WSL not installed, or no/wrong distro — or returns nil when the
// failure is not a WSL-availability problem (caller falls back to
// notInstalledError). This stops every WSL hiccup from collapsing into the
// misleading "OpenFrame is not installed inside WSL".
func wslUnavailableError(err error) error {
	if err == nil {
		return nil
	}
	// wsl.exe itself is missing → the WSL feature is not installed.
	if stderrors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf(`WSL is not installed on this system.

On Windows OpenFrame runs inside WSL2. Install it, reboot if prompted, then re-run:

    wsl --install

Set %s=1 to bypass and run natively on Windows (unsupported)`, disableEnv)
	}
	low := strings.ToLower(wslStderr(err))
	switch {
	case strings.Contains(low, "no distribution"),
		strings.Contains(low, "wsl_e_distro_not_found"),
		strings.Contains(low, "has no installed distributions"):
		return fmt.Errorf(`WSL has no usable distribution (%s).

Install one and re-run, or point OpenFrame at an existing distro:

    wsl --install -d Ubuntu
    # or, if you already have one: set %s=<name>   (list with: wsl -l -q)

Set %s=1 to bypass and run natively on Windows (unsupported)`, distroLabel(), distroEnv, disableEnv)
	}
	return nil
}

// distroLabel describes the targeted WSL distribution for user-facing messages.
func distroLabel() string {
	if d := strings.TrimSpace(os.Getenv(distroEnv)); d != "" {
		return d
	}
	return "the default distro"
}

func notInstalledError() error {
	return fmt.Errorf(`OpenFrame is not installed inside WSL (%s)

On Windows the cluster runs in WSL2 and OpenFrame must run there too. A tagged
release is auto-installed into WSL automatically; for a dev/local build point
%s at the Linux binary you built and re-run:

    set %s=C:\path\to\openframe-linux-amd64   (PowerShell: $env:%s="...")

Or install it manually inside WSL, then re-run:

    wsl
    # place the openframe linux binary on your PATH

If your WSL distro is not the default one, select it with %s=<name>
(list them with: wsl -l -q).
Set %s=1 to bypass and run natively on Windows (unsupported)`,
		distroLabel(), localBinaryEnv, localBinaryEnv, localBinaryEnv, distroEnv, disableEnv)
}
