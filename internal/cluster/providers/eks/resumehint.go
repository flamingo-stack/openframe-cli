package eks

import "github.com/openframe/internal/shared/resumehint"

// resumeHintError carries a resume instruction that survives the generic
// interruption handler. On Ctrl+C that handler prints only "Operation cancelled
// by user." and discards err.Error(), so a hint wrapped only as message text is
// lost. internal/shared/errors surfaces the hint via the ResumeHint() method
// even for an interrupted operation. (The GKE twin: gke/resumehint.go.)
//
// This is a thin alias over the shared implementation in
// internal/shared/resumehint to avoid duplicating the type/logic across
// provider packages (see OPENFRAM-001 for cross-package import aliasing
// conventions).
type resumeHintError = resumehint.Error

// withResumeHint attaches hint to err structurally (not just in the message
// text), so it survives the interruption handler that drops err.Error().
func withResumeHint(err error, hint string) error {
	return resumehint.WithResumeHint(err, hint)
}
