package selfupdate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// stubNotice isolates the state file under a temp HOME, clears the opt-out env,
// and pins the timeNow/fetchLatest seams. It returns the stubbed "now" and a
// counter of fetchLatest invocations so tests can assert the 24h gate held.
func stubNotice(t *testing.T, rel Release, fetchErr error) (time.Time, *int) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// os.UserHomeDir reads USERPROFILE on Windows — without this the state
	// file lands in (and leaks between tests via) the real user profile.
	t.Setenv("USERPROFILE", home)
	t.Setenv("OPENFRAME_NO_UPDATE_CHECK", "")
	at := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	calls := 0
	origNow, origFetch := timeNow, fetchLatest
	timeNow = func() time.Time { return at }
	fetchLatest = func(context.Context) (Release, error) {
		calls++
		return rel, fetchErr
	}
	t.Cleanup(func() { timeNow, fetchLatest = origNow, origFetch })
	return at, &calls
}

func TestMaybeNotifyFirstRunChecksAndWritesState(t *testing.T) {
	at, calls := stubNotice(t, Release{TagName: "v2.0.0"}, nil)

	msg := MaybeNotify(context.Background(), "v1.0.0", true)
	if *calls != 1 {
		t.Fatalf("first run should query once, got %d calls", *calls)
	}
	if !strings.Contains(msg, "v1.0.0 → v2.0.0") {
		t.Fatalf("expected upgrade notice, got %q", msg)
	}
	st := loadState()
	if st.Latest != "v2.0.0" || st.LastCheck != at.Unix() {
		t.Fatalf("state not persisted: %+v", st)
	}
}

func TestMaybeNotifyFreshStateServesCacheWithoutNetwork(t *testing.T) {
	at, calls := stubNotice(t, Release{}, errors.New("network must not be hit"))
	saveState(noticeState{LastCheck: at.Add(-time.Hour).Unix(), Latest: "v2.0.0"})

	if msg := MaybeNotify(context.Background(), "v1.0.0", true); !strings.Contains(msg, "v2.0.0") {
		t.Fatalf("expected cached notice, got %q", msg)
	}
	// Up-to-date binary against the same cache stays silent — still no network.
	if msg := MaybeNotify(context.Background(), "v2.0.0", true); msg != "" {
		t.Fatalf("up-to-date should be silent, got %q", msg)
	}
	if *calls != 0 {
		t.Fatalf("fresh cache must not trigger a network call, got %d", *calls)
	}
}

func TestMaybeNotifyStaleStateRefreshes(t *testing.T) {
	at, calls := stubNotice(t, Release{TagName: "v3.0.0"}, nil)
	saveState(noticeState{LastCheck: at.Add(-checkInterval - time.Minute).Unix(), Latest: "v2.0.0"})

	msg := MaybeNotify(context.Background(), "v1.0.0", true)
	if *calls != 1 {
		t.Fatalf("stale cache should re-query once, got %d calls", *calls)
	}
	if !strings.Contains(msg, "v3.0.0") {
		t.Fatalf("expected refreshed notice, got %q", msg)
	}
	st := loadState()
	if st.Latest != "v3.0.0" || st.LastCheck != at.Unix() {
		t.Fatalf("state not refreshed: %+v", st)
	}
}

func TestMaybeNotifyCorruptStateBehavesLikeFirstRun(t *testing.T) {
	at, calls := stubNotice(t, Release{TagName: "v2.0.0"}, nil)
	p, err := stateFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	msg := MaybeNotify(context.Background(), "v1.0.0", true)
	if *calls != 1 {
		t.Fatalf("corrupt state should fall through to a query, got %d calls", *calls)
	}
	if !strings.Contains(msg, "v2.0.0") {
		t.Fatalf("expected notice despite corrupt state, got %q", msg)
	}
	// The corrupt file is replaced with a valid one on the way out.
	b, err := os.ReadFile(p) //nolint:gosec // G304: test-owned temp path
	if err != nil {
		t.Fatal(err)
	}
	var st noticeState
	if err := json.Unmarshal(b, &st); err != nil || st.LastCheck != at.Unix() {
		t.Fatalf("state file not rewritten cleanly: %q (err %v)", b, err)
	}
}

func TestMaybeNotifyStateWriteFailureIsSilent(t *testing.T) {
	_, calls := stubNotice(t, Release{TagName: "v2.0.0"}, nil)
	// A regular file where the state directory should be makes MkdirAll fail.
	if err := os.WriteFile(filepath.Join(os.Getenv("HOME"), ".openframe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if msg := MaybeNotify(context.Background(), "v1.0.0", true); !strings.Contains(msg, "v2.0.0") {
		t.Fatalf("write failure must not eat the notice, got %q", msg)
	}
	if *calls != 1 {
		t.Fatalf("expected one query, got %d", *calls)
	}
	// Best effort held: nothing persisted, so the next run queries again.
	if MaybeNotify(context.Background(), "v1.0.0", true); *calls != 2 {
		t.Fatalf("unpersisted state should re-query, got %d calls", *calls)
	}
}

func TestMaybeNotifyUpToDateAfterCheckIsSilent(t *testing.T) {
	at, calls := stubNotice(t, Release{TagName: "v1.0.0"}, nil)
	if msg := MaybeNotify(context.Background(), "v1.0.0", true); msg != "" {
		t.Fatalf("up-to-date check should be silent, got %q", msg)
	}
	// The successful (if quiet) check still stamps the rate limit.
	if st := loadState(); *calls != 1 || st.LastCheck != at.Unix() {
		t.Fatalf("calls=%d state=%+v, want one stamped check", *calls, st)
	}
}

func TestMaybeNotifyFetchErrorStaysSilent(t *testing.T) {
	_, calls := stubNotice(t, Release{}, errors.New("boom"))
	if msg := MaybeNotify(context.Background(), "v1.0.0", true); msg != "" {
		t.Fatalf("fetch failure must be silent, got %q", msg)
	}
	if *calls != 1 {
		t.Fatalf("expected one attempted query, got %d", *calls)
	}
	if st := loadState(); st.LastCheck != 0 {
		t.Fatalf("failed check must not stamp the rate limit: %+v", st)
	}
}

func TestMaybeNotifySuppression(t *testing.T) {
	_, calls := stubNotice(t, Release{TagName: "v2.0.0"}, nil)

	t.Setenv("OPENFRAME_NO_UPDATE_CHECK", "1")
	if msg := MaybeNotify(context.Background(), "v1.0.0", true); msg != "" {
		t.Fatalf("opt-out env should silence the notice, got %q", msg)
	}
	t.Setenv("OPENFRAME_NO_UPDATE_CHECK", "")
	if msg := MaybeNotify(context.Background(), "v1.0.0", false); msg != "" {
		t.Fatalf("non-interactive should silence the notice, got %q", msg)
	}
	if msg := MaybeNotify(context.Background(), "dev", true); msg != "" {
		t.Fatalf("dev build should silence the notice, got %q", msg)
	}
	if *calls != 0 {
		t.Fatalf("suppressed paths must never query, got %d calls", *calls)
	}
}

// TestStateHomeLookupFailure covers the stateFile error branch: with no
// resolvable home, loading yields the zero state and saving is a silent no-op.
func TestStateHomeLookupFailure(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "") // the Windows home source for os.UserHomeDir
	if st := loadState(); st != (noticeState{}) {
		t.Fatalf("expected zero state without a home dir, got %+v", st)
	}
	saveState(noticeState{LastCheck: 1, Latest: "v1.0.0"}) // must not panic
}
