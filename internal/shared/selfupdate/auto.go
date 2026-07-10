package selfupdate

import (
	"context"
	"fmt"
	"os"
	"time"

	sharedconfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"golang.org/x/mod/semver"
)

// autoUpdateEnv opts into automatic self-update. It is off unless explicitly set
// — the CLI manages live clusters, so an unexpected version change is never
// imposed.
const autoUpdateEnv = "OPENFRAME_AUTO_UPDATE"

// AutoUpdateEnabled reports whether the user opted into automatic self-update.
func AutoUpdateEnabled() bool {
	return sharedconfig.EnvBool(autoUpdateEnv)
}

// MaybeAutoUpdate, when auto-update is opted in, performs a rate-limited check
// and, if a same-major newer release is available, updates in place. It returns
// a one-line status for stderr (or "" when nothing happened). Best-effort: it
// never returns an error or changes the command's exit code, and a major-version
// bump is never auto-applied — that only surfaces a manual-update notice.
func MaybeAutoUpdate(ctx context.Context, current string, interactive bool, progress func(string)) string {
	if !AutoUpdateEnabled() || noticeSuppressed(current, interactive) {
		return ""
	}
	// Share the daily rate-limit marker with the passive notice.
	now := time.Now()
	if st := loadState(); st.LastCheck != 0 && now.Sub(time.Unix(st.LastCheck, 0)) < checkInterval {
		return ""
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	rel, err := Client{Token: os.Getenv("GITHUB_TOKEN")}.Latest(cctx)
	cancel()
	if err != nil {
		return ""
	}
	saveState(noticeState{LastCheck: now.Unix(), Latest: rel.TagName})

	if !IsNewer(current, rel.TagName) {
		return ""
	}
	if !sameMajor(current, rel.TagName) {
		// A major bump may be breaking — surface it, never auto-apply it.
		return notice(current, rel.TagName) + " (auto-update skips major versions)"
	}

	// Deadline: this runs unattended AFTER the user's command (root passes
	// context.Background), so a stalled download/verify must never hang the CLI
	// exit indefinitely.
	actx, acancel := context.WithTimeout(ctx, 10*time.Minute)
	defer acancel()
	u := Updater{Current: current, Client: Client{Token: os.Getenv("GITHUB_TOKEN")}}
	if err := u.Apply(actx, rel, progress); err != nil {
		return fmt.Sprintf("auto-update to %s failed (run `openframe update`): %v", rel.TagName, err)
	}
	return fmt.Sprintf("Auto-updated %s → %s. Run `openframe update rollback` to revert.", current, rel.TagName)
}

// sameMajor reports whether current and latest share a semver major version.
func sameMajor(current, latest string) bool {
	c, l := normalizeVersion(current), normalizeVersion(latest)
	if c == "" || l == "" {
		return false
	}
	return semver.Major(c) == semver.Major(l)
}
