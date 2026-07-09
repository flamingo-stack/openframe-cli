package target

import (
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

// UIPrompter implements k8s.Prompter using the shared interactive UI helpers.
// This is the production prompter for the context-selection flow.
type UIPrompter struct{}

// Confirm asks a yes/no question with a default answer.
func (UIPrompter) Confirm(message string, defaultYes bool) (bool, error) {
	return ui.ConfirmActionInteractive(message, defaultYes)
}

// Choose presents a list and returns the selected index.
func (UIPrompter) Choose(label string, options []string) (int, error) {
	idx, _, err := ui.SelectFromList(label, options)
	return idx, err
}

var _ k8s.Prompter = UIPrompter{}
