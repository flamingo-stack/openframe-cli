package executor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin the security contract of the WSL command wrapper (audit I3):
// tool arguments must always be passed as DISCRETE argv elements, never spliced
// into a shell -c string, and the wrapper script must be a constant into which
// no argument is ever interpolated.

// contains reports whether argv has an element exactly equal to want.
func containsExact(argv []string, want string) bool {
	for _, a := range argv {
		if a == want {
			return true
		}
	}
	return false
}

// scriptOf returns the script passed to `bash -c` in a wrapped argv.
func scriptOf(t *testing.T, argv []string) string {
	t.Helper()
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == "-c" {
			return argv[i+1]
		}
	}
	t.Fatalf("no `-c <script>` found in argv: %v", argv)
	return ""
}

func TestBuildWSL_HelmArgsArePassedDiscretely(t *testing.T) {
	injected := "x=$(touch /tmp/pwn)"
	cmd, argv := buildWSLCommand("helm", []string{"upgrade", "--install", "--set", injected}, "runner")

	require.Equal(t, "wsl", cmd)
	// The malicious value must appear verbatim as its own argv element...
	assert.True(t, containsExact(argv, injected),
		"injected value must be a discrete argv element, got: %v", argv)
	// ...and must NOT be embedded inside the bash -c script.
	script := scriptOf(t, argv)
	assert.NotContains(t, script, injected,
		"injected value must never be spliced into the shell script")
	// The script must be exactly the constant — no per-call construction.
	assert.Equal(t, helmWSLScript, script)
}

func TestBuildWSL_KubectlJsonpathSurvivesVerbatim(t *testing.T) {
	jp := "jsonpath={.items[*].metadata.name}"
	_, argv := buildWSLCommand("kubectl", []string{"get", "pods", "-o", jp}, "runner")

	assert.True(t, containsExact(argv, jp),
		"jsonpath with {} and $ must pass through unescaped as a discrete arg: %v", argv)
	assert.Equal(t, kubectlWSLScript, scriptOf(t, argv))
}

func TestBuildWSL_HostileUserIsSanitized(t *testing.T) {
	_, argv := buildWSLCommand("kubectl", []string{"get", "ns"}, "evil; rm -rf /")

	// -u must be followed by the safe fallback, never the hostile value.
	var userArg string
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == "-u" {
			userArg = argv[i+1]
		}
	}
	assert.Equal(t, "runner", userArg)
	// And the hostile value must appear nowhere in the argv.
	assert.False(t, strings.Contains(strings.Join(argv, "\x00"), "rm -rf"),
		"hostile user value leaked into argv: %v", argv)
}

func TestSanitizeWSLUser(t *testing.T) {
	cases := map[string]string{
		"":               "runner",
		"runner":         "runner",
		"dev_user-1":     "dev_user-1",
		"evil; rm -rf /": "runner",
		"a b":            "runner",
		"$(whoami)":      "runner",
		"-flag":          "runner", // must start with letter/underscore
	}
	for in, want := range cases {
		assert.Equalf(t, want, sanitizeWSLUser(in), "sanitizeWSLUser(%q)", in)
	}
}

func TestBuildWSL_FiltersContextFlags(t *testing.T) {
	// helm: --kube-context (space and = forms) removed.
	_, helmArgv := buildWSLCommand("helm",
		[]string{"upgrade", "--kube-context", "k3d-x", "--namespace", "argocd"}, "runner")
	assert.False(t, containsExact(helmArgv, "--kube-context"), "helm argv: %v", helmArgv)
	assert.False(t, containsExact(helmArgv, "k3d-x"), "context value must be dropped: %v", helmArgv)
	assert.True(t, containsExact(helmArgv, "argocd"), "unrelated args must survive: %v", helmArgv)

	_, helmArgv2 := buildWSLCommand("helm", []string{"upgrade", "--kube-context=k3d-x"}, "runner")
	assert.False(t, containsExact(helmArgv2, "--kube-context=k3d-x"), "= form must be dropped: %v", helmArgv2)

	// kubectl: --context removed.
	_, kArgv := buildWSLCommand("kubectl", []string{"get", "pods", "--context", "k3d-x"}, "runner")
	assert.False(t, containsExact(kArgv, "--context"), "kubectl argv: %v", kArgv)
	assert.False(t, containsExact(kArgv, "k3d-x"), "kubectl context value must be dropped: %v", kArgv)
}

func TestBuildWSL_K3dUsesSudoAndPassesArgs(t *testing.T) {
	cmd, argv := buildWSLCommand("k3d", []string{"cluster", "create", "my-cluster"}, "runner")
	require.Equal(t, "wsl", cmd)
	assert.True(t, containsExact(argv, "sudo"))
	assert.True(t, containsExact(argv, "-E"))
	assert.True(t, containsExact(argv, "my-cluster"))
	// No bash -c shell wrapping for k3d.
	assert.False(t, containsExact(argv, "-c"), "k3d should not be wrapped in bash -c: %v", argv)
}

func TestBuildWSL_HomeIsDiscretePositional(t *testing.T) {
	_, argv := buildWSLCommand("helm", []string{"version"}, "runner")
	// /home/runner is passed as a positional after the script (so it is $1),
	// not interpolated into the script body.
	assert.True(t, containsExact(argv, "/home/runner"), "argv: %v", argv)
	assert.NotContains(t, scriptOf(t, argv), "/home/runner")
}

func TestBuildWSL_UnknownCommandPassthrough(t *testing.T) {
	cmd, argv := buildWSLCommand("docker", []string{"ps"}, "runner")
	assert.Equal(t, "docker", cmd)
	assert.Equal(t, []string{"ps"}, argv)
}

// TestWSLScripts_KubeconfigEditIsBackedUpAndAtomic is the I8 guard: the WSL
// scripts must back up the kubeconfig and replace it atomically (temp + mv),
// never edit it in place with `sed -i`.
func TestWSLScripts_KubeconfigEditIsBackedUpAndAtomic(t *testing.T) {
	for name, script := range map[string]string{"helm": helmWSLScript, "kubectl": kubectlWSLScript} {
		assert.NotContainsf(t, script, "sed -i", "%s script must not edit kubeconfig in place", name)
		assert.Containsf(t, script, ".kube/config.openframe.bak", "%s script must back up kubeconfig", name)
		assert.Containsf(t, script, `mv "$HOME/.kube/config.openframe.tmp" "$HOME/.kube/config"`,
			"%s script must replace kubeconfig atomically via mv", name)
	}
}
