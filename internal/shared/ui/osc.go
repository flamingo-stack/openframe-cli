package ui

import (
	"fmt"
	"os"
)

// Terminal OSC escape helpers: clickable hyperlinks (OSC 8) and desktop
// notifications (OSC 9). Both degrade to nothing-fancy when the output is not
// an interactive terminal, so logs and pipes never accumulate escape garbage.

// hyperlinksSupported reports whether the terminal is known to render OSC 8
// hyperlinks. Allowlist, not blanket: a terminal that does NOT support OSC 8
// may print the raw escape bytes into the text, which is worse than no link.
func hyperlinksSupported() bool {
	if !isTerminalEnvironment() || os.Getenv("TERM") == "dumb" {
		return false
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "ghostty", "vscode", "Hyper", "Tabby":
		return true
	}
	// Konsole and every VTE-based terminal (GNOME Terminal, Tilix, …).
	return os.Getenv("KONSOLE_VERSION") != "" || os.Getenv("VTE_VERSION") != ""
}

// Hyperlink renders text as a clickable OSC 8 link where supported; elsewhere
// it returns the plain text, appending the URL when the text does not already
// carry it.
func Hyperlink(url, text string) string {
	if url == "" {
		return text
	}
	if text == "" {
		text = url
	}
	if !hyperlinksSupported() {
		if text == url {
			return url
		}
		return text + " (" + url + ")"
	}
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}

// Notify announces a finished long-running operation: an OSC 9 desktop
// notification (iTerm2, WezTerm, ghostty, Konsole show a toast; others ignore
// the sequence) plus BEL, so even terminals without OSC 9 mark the tab for a
// user who switched away during a 20-minute install. No-op when the output is
// not a terminal or --silent is in effect.
func Notify(message string) {
	if !isTerminalEnvironment() || IsSilent() || os.Getenv("TERM") == "dumb" {
		return
	}
	fmt.Fprintf(os.Stdout, "\x1b]9;%s\x07\a", message)
}
