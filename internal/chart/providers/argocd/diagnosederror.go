package argocd

// diagnosedError is an error whose message already embeds its own root-cause
// diagnosis (pod crash logs, warning events, per-app health messages). The
// generic error handler recognizes it via the SelfDiagnosed method (matched
// structurally by internal/shared/errors) and suppresses its pattern-matched
// "friendly hint": matching on the embedded diagnostics misfires — a pod log
// line like "connect: connection refused" made the handler claim the CLUSTER
// was unreachable while it was serving the very diagnostics being printed.
type diagnosedError struct {
	msg string
}

func (e *diagnosedError) Error() string       { return e.msg }
func (e *diagnosedError) SelfDiagnosed() bool { return true }

// selfDiagnosedError wraps a fully-built diagnostic message as an error that
// the generic handler will not decorate with a misfiring hint.
func selfDiagnosedError(msg string) error {
	return &diagnosedError{msg: msg}
}
