package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
)

// checkInterval is how often the passive notice re-queries GitHub. Between
// checks the last-known-latest version is served from the on-disk cache.
const checkInterval = 24 * time.Hour

// noticeState is the cached result of the passive update check.
type noticeState struct {
	LastCheck int64  `json:"lastCheck"` // unix seconds
	Latest    string `json:"latest"`
}

func stateFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openframe", "state", "update-check.json"), nil
}

func loadState() noticeState {
	var s noticeState
	p, err := stateFile()
	if err != nil {
		return s
	}
	// p is a fixed CLI-owned path (~/.openframe/state/update-check.json), not
	// user-controlled input.
	if b, err := os.ReadFile(p); err == nil { //nolint:gosec // G304: fixed CLI-owned path
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func saveState(s noticeState) {
	p, err := stateFile()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return
	}
	if b, err := json.Marshal(s); err == nil {
		_ = os.WriteFile(p, b, 0o600)
	}
}

// noticeSuppressed reports whether the passive update notice must stay silent:
// an explicit opt-out, a non-interactive environment (CI, pipes, machine
// output all report non-interactive via ui.IsNonInteractive), or a dev build
// with no comparable version.
func noticeSuppressed(current string, interactive bool) bool {
	if sharedconfig.EnvBool("OPENFRAME_NO_UPDATE_CHECK") {
		return true
	}
	if !interactive {
		return true
	}
	return normalizeVersion(current) == ""
}

// MaybeNotify performs a best-effort, rate-limited check and returns a one-line
// upgrade notice, or "" when none is warranted, the check is suppressed, or any
// error occurs. It never blocks longer than a short timeout and never surfaces
// errors — callers print the result to stderr so machine output stays clean.
func MaybeNotify(ctx context.Context, current string, interactive bool) string {
	if noticeSuppressed(current, interactive) {
		return ""
	}
	now := time.Now()
	if st := loadState(); st.Latest != "" && st.LastCheck != 0 &&
		now.Sub(time.Unix(st.LastCheck, 0)) < checkInterval {
		// Serve from cache within the interval — no network call.
		if IsNewer(current, st.Latest) {
			return notice(current, st.Latest)
		}
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	rel, err := Client{Token: GitHubToken()}.Latest(ctx)
	if err != nil {
		return "" // best effort; stay silent on any failure
	}
	saveState(noticeState{LastCheck: now.Unix(), Latest: rel.TagName})
	if IsNewer(current, rel.TagName) {
		return notice(current, rel.TagName)
	}
	return ""
}

func notice(current, latest string) string {
	return fmt.Sprintf("A new OpenFrame CLI release is available: %s → %s. Run `openframe update` to upgrade.", current, latest)
}
