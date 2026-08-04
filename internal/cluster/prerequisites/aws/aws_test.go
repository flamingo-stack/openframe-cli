package aws

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// stubLookPath makes only the named commands resolvable, restoring the real
// lookup after the test.
func stubLookPath(t *testing.T, available ...string) {
	t.Helper()
	orig := lookPath
	t.Cleanup(func() { lookPath = orig })
	lookPath = func(cmd string) (string, error) {
		for _, name := range available {
			if cmd == name {
				return "/usr/bin/" + cmd, nil
			}
		}
		return "", errors.New("not found: " + cmd)
	}
}

// stubRunCommand records every invocation and never executes anything.
func stubRunCommand(t *testing.T, err error) *[][]string {
	t.Helper()
	orig := runCommand
	t.Cleanup(func() { runCommand = orig })
	var calls [][]string
	runCommand = func(name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return err
	}
	return &calls
}

func stubAwsVersionOutput(t *testing.T, out string, err error) {
	t.Helper()
	orig := awsVersionOutput
	t.Cleanup(func() { awsVersionOutput = orig })
	awsVersionOutput = func() ([]byte, error) { return []byte(out), err }
}

// isV2Output must find the version token anywhere in the combined output:
// wrapper noise on stderr (e.g. Python deprecation warnings) can precede the
// version line, and a prefix check then misreported v2 as v1.
func TestIsV2Output(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want bool
	}{
		{"clean v2", "aws-cli/2.15.30 Python/3.11.8 Darwin/23.4.0 exe/x86_64\n", true},
		{"v2 behind stderr noise", "DeprecationWarning: Python 3.8 support ends soon\naws-cli/2.15.30 Python/3.8.18 Linux/6.5\n", true},
		{"clean v1", "aws-cli/1.29.0 Python/3.10.12 Linux/6.5 botocore/1.31.0\n", false},
		{"garbage", "command not found: aws", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isV2Output([]byte(tc.out)); got != tc.want {
				t.Errorf("isV2Output(%q) = %v, want %v", tc.out, got, tc.want)
			}
		})
	}
}

func TestInstallLinux_PrefersAptAndUsesNonInteractiveSudo(t *testing.T) {
	// apt and dnf both present: apt must win (fallback order is positional).
	stubLookPath(t, "apt", "dnf")
	calls := stubRunCommand(t, nil)
	stubAwsVersionOutput(t, "aws-cli/2.15.30 Python/3.11.8 Linux/6.5\n", nil)

	if err := NewAwsInstaller().installLinux(); err != nil {
		t.Fatalf("install must succeed when apt installs v2: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("want exactly one package-manager invocation, got %v", *calls)
	}
	got := strings.Join((*calls)[0], " ")
	// sudo -n: never prompt for a password mid-flow.
	if want := "sudo -n apt install -y awscli"; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestInstallLinux_FallsBackToDnfWhenAptAbsent(t *testing.T) {
	stubLookPath(t, "dnf", "yum")
	calls := stubRunCommand(t, nil)
	stubAwsVersionOutput(t, "aws-cli/2.15.30 Python/3.11.8 Linux/6.5\n", nil)

	if err := NewAwsInstaller().installLinux(); err != nil {
		t.Fatalf("install must succeed via dnf: %v", err)
	}
	got := strings.Join((*calls)[0], " ")
	if want := "sudo -n dnf install -y awscli2"; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

// Older repos (e.g. Ubuntu 22.04) ship legacy v1, whose `aws eks get-token`
// output kubectl no longer accepts — a v1 result is a failure, not a success.
func TestInstallLinux_RejectsInstalledV1(t *testing.T) {
	stubLookPath(t, "apt")
	stubRunCommand(t, nil)
	stubAwsVersionOutput(t, "aws-cli/1.29.0 Python/3.10.12 Linux/6.5 botocore/1.31.0\n", nil)

	err := NewAwsInstaller().installLinux()
	if err == nil {
		t.Fatal("a v1 install must be rejected")
	}
	if !strings.Contains(err.Error(), "v1") || !strings.Contains(err.Error(), "v2") {
		t.Fatalf("error must explain the v1/v2 mismatch, got: %v", err)
	}
}

func TestInstallLinux_NoPackageManager(t *testing.T) {
	stubLookPath(t) // nothing resolvable
	calls := stubRunCommand(t, nil)

	err := NewAwsInstaller().installLinux()
	if err == nil {
		t.Fatal("expected an error when no package manager exists")
	}
	if len(*calls) != 0 {
		t.Fatalf("no command may run without a package manager, got %v", *calls)
	}
}

// A failing package manager must fall through to the next one, not abort.
func TestInstallLinux_FailedManagerFallsThrough(t *testing.T) {
	stubLookPath(t, "apt", "dnf")
	orig := runCommand
	t.Cleanup(func() { runCommand = orig })
	var calls [][]string
	runCommand = func(name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		if len(calls) == 1 {
			return fmt.Errorf("apt broke")
		}
		return nil
	}
	stubAwsVersionOutput(t, "aws-cli/2.15.30 Python/3.11.8 Linux/6.5\n", nil)

	if err := NewAwsInstaller().installLinux(); err != nil {
		t.Fatalf("dnf fallback must succeed after apt failure: %v", err)
	}
	if len(calls) != 2 || calls[1][2] != "dnf" {
		t.Fatalf("want apt then dnf, got %v", calls)
	}
}

func TestGetInstallHelp_NonEmpty(t *testing.T) {
	if NewAwsInstaller().GetInstallHelp() == "" {
		t.Fatal("install help must not be empty")
	}
}
