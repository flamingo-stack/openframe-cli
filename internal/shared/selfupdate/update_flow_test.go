package selfupdate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Offline guards for the update/rollback flow. The live trust chain (real
// release, cosign bundle, checksums) is exercised end-to-end by the CI step
// "Update: apply the real latest release, then roll back"; these tests pin the
// decisions that must hold without a network.

// releaseAPI serves a minimal /releases/latest + /releases/tags/<tag>.
func releaseAPI(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	body := fmt.Sprintf(`{"tag_name":%q,"html_url":"https://example/rel"}`, tag)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/flamingo-stack/openframe-cli/releases/latest",
			"/repos/flamingo-stack/openframe-cli/releases/tags/" + tag:
			_, _ = w.Write([]byte(body))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestCheck_DevBuildIsFlaggedAndNotNewer: a dev build must be reported as such
// so `openframe update` refuses to replace it (an unversioned binary has
// nothing to compare against and no rollback point to restore).
func TestCheck_DevBuildIsFlaggedAndNotNewer(t *testing.T) {
	srv := releaseAPI(t, "9.9.9")
	u := Updater{Current: "dev", Client: Client{APIBase: srv.URL}}

	st, _, err := u.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !st.DevBuild {
		t.Error("a dev build must be flagged DevBuild")
	}
	if st.Available {
		t.Error("no update may be offered to a dev build")
	}
}

// TestCheck_ReleaseBuildSeesNewerRelease is the control case.
func TestCheck_ReleaseBuildSeesNewerRelease(t *testing.T) {
	srv := releaseAPI(t, "9.9.9")
	u := Updater{Current: "0.0.1", Client: Client{APIBase: srv.URL}}

	st, rel, err := u.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if st.DevBuild {
		t.Error("0.0.1 is a release version, not a dev build")
	}
	if !st.Available {
		t.Error("9.9.9 must be offered to 0.0.1")
	}
	if rel.TagName != "9.9.9" {
		t.Errorf("tag = %q", rel.TagName)
	}
}

// TestCheck_ExplicitTagBothSpellings: `update 0.4.7` and `update v0.4.7` must
// both resolve against a bare-tagged release (this repo tags without the "v").
func TestCheck_ExplicitTagBothSpellings(t *testing.T) {
	srv := releaseAPI(t, "0.4.7")
	u := Updater{Current: "0.0.1", Client: Client{APIBase: srv.URL}}

	for _, spelling := range []string{"0.4.7", "v0.4.7"} {
		_, rel, err := u.Check(context.Background(), spelling)
		if err != nil {
			t.Fatalf("Check(%q): %v", spelling, err)
		}
		if rel.TagName != "0.4.7" {
			t.Errorf("Check(%q) resolved to %q", spelling, rel.TagName)
		}
	}
}

// TestApply_RefusesNativeWindows: the native Windows launcher forwards the CLI
// into WSL, so the Linux binary there is the one that self-updates; replacing
// the Windows executable would be wrong (and it is locked while running).
func TestApply_RefusesNativeWindowsWithGuidance(t *testing.T) {
	u := Updater{Current: "0.0.1", GOOS: "windows"}
	err := u.Apply(context.Background(), Release{TagName: "9.9.9"}, nil)
	if err == nil {
		t.Fatal("Apply must refuse on the native Windows launcher")
	}
	if !strings.Contains(err.Error(), "WSL") {
		t.Errorf("the refusal should point at the WSL binary, got: %v", err)
	}
}

// TestRollback_ConsumesThePointOnce: rollback restores the retained binary and
// clears the rollback point, so a second rollback is a clean no-op rather than
// bouncing between two versions.
func TestRollback_ConsumesThePointOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh stub binaries; unix-only")
	}
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	exe := filepath.Join(dir, "openframe")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\necho 2.0.0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	prev := filepath.Join(dir, "prev")
	if err := os.WriteFile(prev, []byte("#!/bin/sh\necho 1.0.0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := savePrevious(prev); err != nil {
		t.Fatal(err)
	}

	u := Updater{Current: "2.0.0", GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, exePath: exe}
	if err := u.Rollback(context.Background(), nil); err != nil {
		t.Fatalf("first rollback: %v", err)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "echo 1.0.0") {
		t.Fatalf("binary not restored: %q", got)
	}

	// Point consumed: a second rollback finds nothing to restore.
	if _, ok := PreviousVersion(); ok {
		t.Error("the rollback point must be cleared after use")
	}
	if err := u.Rollback(context.Background(), nil); err == nil {
		t.Error("a second rollback must fail (nothing saved), not restore again")
	}
}
