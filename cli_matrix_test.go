package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// CLI invocation matrix: run the REAL binary with a spread of argument
// combinations and validate exit codes and output. Everything here is
// machine-independent by construction — no docker/k3d/helm, no network, no
// cluster, no interactive prompts (stdin is closed, so the CLI is in its
// non-interactive mode). Anything that needs real tools lives in the e2e
// workflow (.github/workflows/test.yml), not here.

var (
	matrixBinOnce sync.Once
	matrixBinPath string
	matrixBinErr  error
)

// matrixBin builds the CLI once per test process.
func matrixBin(t *testing.T) string {
	t.Helper()
	matrixBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "of-cli-matrix")
		if err != nil {
			matrixBinErr = err
			return
		}
		matrixBinPath = filepath.Join(dir, "openframe-matrix-test")
		out, err := exec.Command("go", "build", "-o", matrixBinPath, ".").CombinedOutput()
		if err != nil {
			matrixBinErr = err
			matrixBinPath = ""
			_ = os.RemoveAll(dir)
			t.Logf("build output: %s", out)
		}
	})
	if matrixBinErr != nil {
		t.Fatalf("building test binary: %v", matrixBinErr)
	}
	return matrixBinPath
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// runMatrix executes the binary in a hermetic environment: isolated HOME (no
// ~/.openframe state), no kubeconfig, update checks off, WSL forwarding off
// (so the matrix behaves identically on native Windows), stdin closed.
func runMatrix(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	home := t.TempDir()
	cmd := exec.Command(matrixBin(t), args...)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
		"KUBECONFIG="+filepath.Join(home, "no-such-kubeconfig"),
		"OPENFRAME_NO_UPDATE_CHECK=1",
		"OPENFRAME_NO_WSL_FORWARD=1",
	)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuf, &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("running %v: %v", args, err)
		}
	}
	return ansiRe.ReplaceAllString(outBuf.String(), ""), ansiRe.ReplaceAllString(errBuf.String(), ""), exitCode
}

func TestCLIMatrix(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantExit int
		contains []string // matched against stdout+stderr, ANSI-stripped
		absent   []string
	}{
		// ---- happy read-only surface -----------------------------------
		{"version", []string{"--version"}, 0, []string{"dev"}, nil},
		{"root help", []string{"--help"}, 0,
			[]string{"Available Commands", "cluster", "app", "bootstrap", "prerequisites", "update"}, nil},
		{"app help lists subcommands", []string{"app", "--help"}, 0,
			[]string{"install", "upgrade", "status", "access", "uninstall"}, nil},
		{"cluster help lists subcommands", []string{"cluster", "--help"}, 0,
			[]string{"create", "delete", "list", "status", "cleanup"}, nil},
		{"update help lists subcommands", []string{"update", "--help"}, 0,
			[]string{"check", "rollback"}, nil},
		{"completion bash", []string{"completion", "bash"}, 0, []string{"openframe"}, nil},
		{"completion zsh", []string{"completion", "zsh"}, 0, []string{"openframe"}, nil},
		{"completion fish", []string{"completion", "fish"}, 0, []string{"openframe"}, nil},
		{"completion powershell", []string{"completion", "powershell"}, 0, []string{"openframe"}, nil},
		// Rollback with no prior update: clean offline no-op, exit 0.
		{"update rollback with nothing saved", []string{"update", "rollback"}, 0,
			[]string{"No previous version"}, nil},

		// ---- unknown surface ---------------------------------------------
		{"unknown command", []string{"bogus"}, 1, []string{"unknown command"}, nil},
		{"unknown root flag", []string{"--bogus"}, 1, []string{"unknown flag"}, nil},
		{"unknown update flag", []string{"update", "--bogus"}, 1, []string{"unknown flag"}, nil},
		// Removed/legacy flags must fail loudly, not be silently ignored.
		{"removed --github-branch", []string{"app", "install", "--github-branch", "x"}, 1,
			[]string{"unknown flag: --github-branch"}, nil},
		{"legacy --deployment-mode", []string{"app", "install", "--deployment-mode", "oss"}, 1,
			[]string{"unknown flag: --deployment-mode"}, nil},

		// ---- flag/arg validation (parse-time, before any gate) ------------
		{"non-numeric --nodes", []string{"cluster", "create", "x", "--nodes", "abc"}, 1,
			[]string{`invalid argument "abc"`}, nil},
		{"non-bool --prune", []string{"app", "upgrade", "--prune=banana"}, 1,
			[]string{"invalid argument"}, nil},
		{"bootstrap too many args", []string{"bootstrap", "a", "b"}, 1,
			[]string{"accepts at most 1 arg"}, nil},
		{"bootstrap invalid cluster name", []string{"bootstrap", "Invalid_Name", "--non-interactive"}, 1,
			[]string{"is invalid", "hyphens"}, nil},

		// ---- command-level guards (fail fast, no cluster contact) ---------
		{"upgrade ref+sync mutually exclusive", []string{"app", "upgrade", "--ref", "x", "--sync"}, 1,
			[]string{"mutually exclusive"}, nil},
		{"uninstall non-interactive needs --yes", []string{"app", "uninstall"}, 1,
			[]string{"--yes", "non-interactive"}, nil},
		{"install with unknown context", []string{"app", "install", "--context", "no-such", "--non-interactive", "--dry-run"}, 1,
			[]string{`could not use context "no-such"`}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runMatrix(t, tc.args...)
			combined := stdout + "\n" + stderr
			if code != tc.wantExit {
				t.Errorf("exit = %d, want %d\noutput:\n%s", code, tc.wantExit, combined)
			}
			for _, want := range tc.contains {
				if !strings.Contains(combined, want) {
					t.Errorf("output missing %q:\n%s", want, combined)
				}
			}
			for _, banned := range tc.absent {
				if strings.Contains(combined, banned) {
					t.Errorf("output must not contain %q:\n%s", banned, combined)
				}
			}
		})
	}
}

// TestCLIMatrix_MachineOutputsOnStdout: script-facing outputs (--version,
// completion scripts) must land on STDOUT — piping them must work.
func TestCLIMatrix_MachineOutputsOnStdout(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"completion", "bash"}} {
		stdout, _, code := runMatrix(t, args...)
		if code != 0 {
			t.Errorf("%v: exit %d", args, code)
		}
		if strings.TrimSpace(stdout) == "" {
			t.Errorf("%v: stdout is empty — machine output must go to stdout", args)
		}
	}
}

// TestCLIMatrix_SilentHelpIsQuiet: --silent must not decorate even trivial
// read-only commands with the logo.
func TestCLIMatrix_SilentVersion(t *testing.T) {
	stdout, stderr, code := runMatrix(t, "--silent", "--version")
	if code != 0 {
		t.Fatalf("exit %d (stderr: %s)", code, stderr)
	}
	if strings.Contains(stdout+stderr, "Bootstrapper") {
		t.Error("--silent leaked the logo banner")
	}
}
