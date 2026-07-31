package ui

import (
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/pterm/pterm"
)

// ShowContextHeader prints one dim line naming the kube-context the command
// is about to act against — the cheapest possible guard against running a
// destructive command at the wrong cluster. Skipped when the output is not a
// terminal (machine consumers), under --silent, and when no context exists.
func ShowContextHeader() {
	if IsSilent() || !isTerminalEnvironment() || os.Getenv("TERM") == "dumb" {
		return
	}
	_, current, err := k8s.LoadContexts(k8s.DefaultKubeconfigPath())
	if err != nil || current == "" {
		return
	}
	g := Glyphs()
	pterm.FgGray.Printfln("%s kube-context %s", g.Bullet, current)
}
