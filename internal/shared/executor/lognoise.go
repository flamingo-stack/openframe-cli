package executor

import (
	"regexp"
	"strings"
)

// Subprocess log-noise filtering (0.4.9 verification observation): k3d logs
// every step to stderr in logrus format, so a failed `k3d cluster create`
// embedded a wall of `INFO[0000] ...` progress lines into the user-facing
// error, drowning the one ERRO/FATA line that explains the failure. The lines
// print "outside the CLI's own formatting" because they arrive inside error
// text — no k3d output is ever streamed live (the executor captures both
// stdout and stderr).
//
// Only informational records are dropped; warnings, errors, fatals, and any
// non-logrus line pass through untouched. Helm and docker do not use logrus,
// so their stderr is unaffected.

var (
	// Colored/TTY-style logrus records: `INFO[0007] Using config file...`.
	logrusInfoLine = regexp.MustCompile(`^\s*(?:INFO|DEBU|TRAC)\[[^\]]*\]`)
	// Plain logfmt records: `time="..." level=info msg="..."`.
	logfmtInfoLine = regexp.MustCompile(`^\s*time="[^"]*" level=(?:info|debug|trace) `)
)

// stripLogNoise removes informational logrus/logfmt records from subprocess
// output, preserving everything else (including logrus WARN/ERRO/FATA lines).
func stripLogNoise(s string) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if logrusInfoLine.MatchString(line) || logfmtInfoLine.MatchString(line) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// errorDetail prepares captured stderr for embedding into an error message:
// log noise is stripped first; when NOTHING survives (the child only logged
// informational lines before dying) the raw text is kept — losing the only
// available detail would be worse than the noise.
func errorDetail(stderr string) string {
	filtered := strings.TrimSpace(stripLogNoise(stderr))
	if filtered != "" {
		return filtered
	}
	return strings.TrimSpace(stderr)
}
