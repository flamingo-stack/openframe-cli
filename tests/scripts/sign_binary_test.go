// Package scripts_test covers the repo's release shell scripts; it contains
// no production code.
package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for scripts/sign-binary.sh, the release-signing dispatcher invoked by
// GoReleaser build hooks. The real signing tools (codesign, xcrun, jsign via
// java, curl, jq, zip) are replaced with PATH stubs that record their argv to
// a call log, so these tests validate the OPENFRAME_SIGN gate, the per-OS
// dispatch, and the exact flags each tool receives — without certificates or
// network. The real signatures are verified post-publish by the
// verify-*-signature jobs in .github/workflows/release.yml.

const stubBody = `#!/bin/sh
printf '%s %s\n' "$(basename "$0")" "$*" >> "$CALL_LOG"
exit 0
`

// curl prints CURL_BODY (the fake AAD token response); jq prints JQ_OUT (the
// fake extracted token) after draining stdin.
const curlStubBody = `#!/bin/sh
printf '%s %s\n' curl "$*" >> "$CALL_LOG"
printf '%s' "$CURL_BODY"
`

const jqStubBody = `#!/bin/sh
printf '%s %s\n' jq "$*" >> "$CALL_LOG"
cat > /dev/null
printf '%s' "$JQ_OUT"
`

type signScriptRun struct {
	exitCode int
	output   string
	calls    []string
}

// darwinEnv / windowsEnv are the full sets the release workflow provides.
func darwinEnv() map[string]string {
	return map[string]string{
		"OPENFRAME_SIGN":    "1",
		"SIGNING_IDENTITY":  "Developer ID Application: Test (TEAM123)",
		"KEYCHAIN_PATH":     "/fake/signing.keychain-db",
		"APPLE_ID_USERNAME": "notary@example.com",
		"APPLE_ID_PASSWORD": "app-specific-pass",
		"APPLE_TEAM_ID":     "TEAM123",
	}
}

func windowsEnv(jsignJar string) map[string]string {
	return map[string]string{
		"OPENFRAME_SIGN":                  "1",
		"JSIGN_JAR":                       jsignJar,
		"AZURE_TENANT_ID":                 "tenant-guid",
		"AZURE_CLIENT_ID":                 "client-guid",
		"AZURE_CLIENT_SECRET":             "client-secret",
		"AZURE_SIGNING_ENDPOINT":          "https://eus.codesigning.azure.net",
		"AZURE_CODE_SIGNING_ACCOUNT_NAME": "flamingo-account",
		"AZURE_CERTIFICATE_PROFILE_NAME":  "flamingo-profile",
		"CURL_BODY":                       `{"access_token":"fake-token"}`,
		"JQ_OUT":                          "fake-token",
	}
}

func runSignScript(t *testing.T, env map[string]string, args ...string) signScriptRun {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("sign-binary.sh runs only on the macOS release runner")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	stubDir := t.TempDir()
	for name, body := range map[string]string{
		"codesign": stubBody,
		"xcrun":    stubBody,
		"zip":      stubBody,
		"java":     stubBody,
		"curl":     curlStubBody,
		"jq":       jqStubBody,
	} {
		require.NoError(t, os.WriteFile(filepath.Join(stubDir, name), []byte(body), 0o755))
	}
	callLog := filepath.Join(t.TempDir(), "calls.log")

	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "sign-binary.sh"))
	require.NoError(t, err)
	require.FileExists(t, script)

	cmd := exec.Command("bash", append([]string{script}, args...)...)
	cmd.Env = []string{
		"PATH=" + stubDir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"TMPDIR=" + t.TempDir(),
		"CALL_LOG=" + callLog,
	}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, runErr := cmd.CombinedOutput()
	run := signScriptRun{output: string(out)}
	if runErr != nil {
		var exitErr *exec.ExitError
		require.ErrorAs(t, runErr, &exitErr, "script did not run: %v\n%s", runErr, out)
		run.exitCode = exitErr.ExitCode()
	}
	if logBytes, err := os.ReadFile(callLog); err == nil {
		run.calls = strings.Split(strings.TrimSpace(string(logBytes)), "\n")
		if len(run.calls) == 1 && run.calls[0] == "" {
			run.calls = nil
		}
	}
	return run
}

// findCall returns the index of the first recorded call containing every
// fragment, or -1.
func findCall(calls []string, fragments ...string) int {
	for i, call := range calls {
		matched := true
		for _, f := range fragments {
			if !strings.Contains(call, f) {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}

func fakeBinary(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openframe")
	require.NoError(t, os.WriteFile(path, []byte("fake binary"), 0o755))
	return path
}

func TestSignBinaryScript_GateSkipsWithoutOptIn(t *testing.T) {
	run := runSignScript(t, nil, "darwin", "arm64", fakeBinary(t))
	assert.Equal(t, 0, run.exitCode, run.output)
	assert.Contains(t, run.output, "skipping darwin/arm64")
	assert.Empty(t, run.calls, "no signing tool must be invoked without the opt-in gate")
}

func TestSignBinaryScript_LinuxIsNeverSigned(t *testing.T) {
	run := runSignScript(t, map[string]string{"OPENFRAME_SIGN": "1"}, "linux", "amd64", fakeBinary(t))
	assert.Equal(t, 0, run.exitCode, run.output)
	assert.Contains(t, run.output, "not signed (by design)")
	assert.Empty(t, run.calls)
}

func TestSignBinaryScript_RequiresArguments(t *testing.T) {
	run := runSignScript(t, map[string]string{"OPENFRAME_SIGN": "1"}, "darwin", "arm64")
	assert.NotEqual(t, 0, run.exitCode, "missing <path> must fail")
	assert.Contains(t, run.output, "usage:")
}

func TestSignBinaryScript_DarwinSignsAndNotarizes(t *testing.T) {
	bin := fakeBinary(t)
	run := runSignScript(t, darwinEnv(), "darwin", "arm64", bin)
	require.Equal(t, 0, run.exitCode, run.output)

	sign := findCall(run.calls,
		"codesign --sign Developer ID Application: Test (TEAM123)",
		"--keychain /fake/signing.keychain-db",
		"--timestamp", "--options runtime", "--force", bin)
	require.NotEqual(t, -1, sign, "codesign --sign call missing or wrong flags:\n%s", strings.Join(run.calls, "\n"))

	verify := findCall(run.calls, "codesign --verify --strict", bin)
	require.NotEqual(t, -1, verify, "codesign --verify call missing")

	notarize := findCall(run.calls,
		"xcrun notarytool submit",
		"--apple-id notary@example.com",
		"--team-id TEAM123",
		"--wait")
	require.NotEqual(t, -1, notarize, "notarytool submit call missing or wrong flags")

	assert.Less(t, sign, verify, "must verify after signing")
	assert.Less(t, verify, notarize, "must notarize only a verified signature")
	assert.Contains(t, run.output, "signed and notarized")
}

func TestSignBinaryScript_DarwinRequiresIdentity(t *testing.T) {
	env := darwinEnv()
	delete(env, "SIGNING_IDENTITY")
	run := runSignScript(t, env, "darwin", "arm64", fakeBinary(t))
	assert.NotEqual(t, 0, run.exitCode)
	assert.Equal(t, -1, findCall(run.calls, "codesign"), "must not sign without an identity")
}

func TestSignBinaryScript_WindowsSignsViaTrustedSigning(t *testing.T) {
	bin := fakeBinary(t)
	run := runSignScript(t, windowsEnv("/fake/jsign.jar"), "windows", "amd64", bin)
	require.Equal(t, 0, run.exitCode, run.output)

	token := findCall(run.calls, "curl", "https://login.microsoftonline.com/tenant-guid/oauth2/v2.0/token")
	require.NotEqual(t, -1, token, "AAD token request missing:\n%s", strings.Join(run.calls, "\n"))

	sign := findCall(run.calls,
		"java -jar /fake/jsign.jar",
		"--storetype TRUSTEDSIGNING",
		"--keystore eus.codesigning.azure.net", // https:// prefix must be stripped for jsign
		"--storepass fake-token",
		"--alias flamingo-account/flamingo-profile",
		"--tsaurl http://timestamp.acs.microsoft.com",
		"--tsmode RFC3161",
		bin)
	require.NotEqual(t, -1, sign, "jsign call missing or wrong flags:\n%s", strings.Join(run.calls, "\n"))
	assert.Less(t, token, sign, "token must be fetched before signing")
}

func TestSignBinaryScript_WindowsFailsWithoutToken(t *testing.T) {
	env := windowsEnv("/fake/jsign.jar")
	env["CURL_BODY"] = `{"error":"invalid_client"}`
	env["JQ_OUT"] = ""
	run := runSignScript(t, env, "windows", "amd64", fakeBinary(t))
	assert.NotEqual(t, 0, run.exitCode)
	assert.Contains(t, run.output, "failed to obtain Azure Trusted Signing token")
	assert.Equal(t, -1, findCall(run.calls, "java"), "must not invoke jsign without a token")
}

func TestSignBinaryScript_WindowsRequiresAzureConfig(t *testing.T) {
	env := windowsEnv("/fake/jsign.jar")
	delete(env, "AZURE_CLIENT_SECRET")
	run := runSignScript(t, env, "windows", "amd64", fakeBinary(t))
	assert.NotEqual(t, 0, run.exitCode)
	assert.Empty(t, run.calls, "must fail fast before calling any tool")
}
