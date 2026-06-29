package redact

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Secret redaction for log/debug output (audit I4).
//
// External-command logging (verbose/debug mode) can echo full helm/git command
// lines and URLs that carry credentials. Register secrets as soon as they are
// collected, then route any command line through Redact before printing.

var (
	secretsMu sync.RWMutex
	secrets   = map[string]struct{}{}
)

// urlCredentialPattern matches the "user:pass@" segment of a URL so credentials
// embedded in URLs are scrubbed even if the exact value was never registered.
var urlCredentialPattern = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://)([^/@:\s]+):([^/@\s]+)@`)

const redactionMarker = "***"

// RegisterSecret records a value that must never appear in printed output.
// Empty and very short values are ignored to avoid over-redacting common text.
func RegisterSecret(secret string) {
	if len(secret) < 4 {
		return
	}
	secretsMu.Lock()
	secrets[secret] = struct{}{}
	secretsMu.Unlock()
}

// ClearSecrets removes all registered secrets (intended for tests).
func ClearSecrets() {
	secretsMu.Lock()
	secrets = map[string]struct{}{}
	secretsMu.Unlock()
}

// Redact removes registered secrets and URL-embedded credentials from s.
func Redact(s string) string {
	// Scrub URL credentials structurally first (catches unregistered tokens).
	out := urlCredentialPattern.ReplaceAllString(s, "$1$2:"+redactionMarker+"@")

	secretsMu.RLock()
	values := make([]string, 0, len(secrets))
	for v := range secrets {
		values = append(values, v)
	}
	secretsMu.RUnlock()

	// Replace longer secrets first so a secret that is a substring of another
	// does not partially unmask it.
	sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })

	for _, v := range values {
		out = strings.ReplaceAll(out, v, redactionMarker)
	}
	return out
}
