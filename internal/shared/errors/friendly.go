package errors

import "strings"

// friendlyHint returns a plain-language, actionable hint for common low-level
// failures, or "" when none applies. It exists so non-technical users get a
// next step instead of a raw error string (req 30). It never replaces the
// underlying error — it's shown as an extra "did you mean" line.
func friendlyHint(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	switch {
	case containsAny(msg, "connection refused", "was refused", "unable to connect to the server", "connection reset"):
		return "The cluster isn't reachable — is it running? Try 'openframe cluster status'."
	case containsAny(msg, "no such host", "dns resolution", "name resolution"):
		return "The cluster address couldn't be resolved. Check your kubeconfig / current context."
	case containsAny(msg, "context deadline exceeded", "timed out", "timeout"):
		return "The operation timed out — the cluster may be slow or unreachable. Wait a moment and retry."
	case containsAny(msg, "permission denied", "forbidden", "unauthorized"):
		return "Permission was denied. Check your credentials / kubeconfig for this cluster."
	case strings.Contains(msg, "context") && strings.Contains(msg, "not exist"):
		return "That kube-context doesn't exist. Run 'kubectl config get-contexts' to see the available ones."
	case strings.Contains(msg, "docker") && containsAny(msg, "not running", "cannot connect", "daemon"):
		return "Docker doesn't appear to be running. Start Docker and try again — or run 'openframe prerequisites check'."
	default:
		return ""
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
