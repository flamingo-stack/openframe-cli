package gke

// resumeHintError carries a resume instruction that survives the generic
// interruption handler. On Ctrl+C that handler prints only "Operation cancelled
// by user." and discards err.Error(), so a hint wrapped only as message text is
// lost. internal/shared/errors surfaces the hint via the ResumeHint() method
// even for an interrupted operation.
type resumeHintError struct {
	err  error
	hint string
}

func (e *resumeHintError) Error() string     { return e.err.Error() }
func (e *resumeHintError) Unwrap() error      { return e.err }
func (e *resumeHintError) ResumeHint() string { return e.hint }

// withResumeHint attaches hint to err structurally (not just in the message
// text), so it survives the interruption handler that drops err.Error().
func withResumeHint(err error, hint string) error {
	return &resumeHintError{err: err, hint: hint}
}
