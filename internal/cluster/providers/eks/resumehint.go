package eks

import "yy.foundation.im/base/internal/shared/errors"

// resumeHintError carries a resume instruction that survives the generic
// interruption handler. On Ctrl+C that handler prints only "Operation cancelled
// by user." and discards err.Error(), so a hint wrapped only as message text is
// lost. internal/shared/errors surfaces the hint via the ResumeHint() method
// even for an interrupted operation. (The GKE twin: gke/resumehint.go.)
type resumeHintError = errors.ResumeHintError

// withResumeHint attaches hint to err structurally (not just in the message
// text), so it survives the interruption handler that drops err.Error().
func withResumeHint(err error, hint string) error {
	return errors.WithResumeHint(err, hint)
}
