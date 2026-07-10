package config

import (
	"os"
	"strings"
)

// EnvBool reports whether the named environment variable is set to a truthy
// value: 1, true, yes, or on (case-insensitive). Everything else — including
// "0" and "false" — is false.
//
// This is the ONLY way boolean OPENFRAME_* switches may be read. The old
// `os.Getenv(name) != ""` pattern treated `FLAG=0` and `FLAG=false` as ON —
// for OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY that meant a user writing `=0` to
// say "off" silently DISABLED release-signature verification (audit B5/T2).
func EnvBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
